package krouter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/core/status/codes"
	"github.com/govinda-attal/kiss-lib/pkg/kasync"
)

type MsgHandler func(ctx context.Context, msg *kafka.Message) error

func New(consumerCfg, producerCfg *kafka.ConfigMap, errTopic string) *Router {
	return &Router{
		consumerCfg: consumerCfg,
		producerCfg: producerCfg,
		errTopic:    errTopic,
		routeGrps:   make(map[string]RouteGroup),
	}
}

type AsyncRouter interface {
	io.Closer
	NewRouteGrp(topic string, defHandler MsgHandler) *RouteGroup
	RqTopics() []string
	Serve() error
}

type Router struct {
	io.Closer
	consumerCfg, producerCfg *kafka.ConfigMap
	routeGrps                map[string]RouteGroup
	errTopic                 string
	NotFoundHandler          MsgHandler
	stop                     chan interface{}
}

type RouteGroup interface {
	Invoke(msgName string, handler MsgHandler) RouteGroup
	MsgHandler(msgName string) (MsgHandler, error)
}

type TopicRouteGroup struct {
	topic      string
	defHandler MsgHandler
	handlers   map[string]MsgHandler
}

//MsgHandler returns the handler for the route.
func (rg *TopicRouteGroup) MsgHandler(msgName string) (MsgHandler, error) {
	h, ok := rg.handlers[msgName]
	if !ok {
		if rg.defHandler == nil {
			return nil, status.ErrNotFound().WithMessage(fmt.Sprintf("handler for given message %s not found", msgName))
		}
		return rg.defHandler, nil
	}
	return h, nil
}

func (rg *TopicRouteGroup) Invoke(msgName string, handler MsgHandler) RouteGroup {
	rg.handlers[msgName] = handler
	return rg
}

func (kr *Router) NewRouteGrp(topic string, defHandler MsgHandler) *TopicRouteGroup {
	rg := &TopicRouteGroup{topic: topic,
		defHandler: defHandler,
		handlers:   make(map[string]MsgHandler)}
	kr.routeGrps[topic] = rg
	return rg
}

func (kr *Router) RqTopics() []string {
	var topics []string
	for t := range kr.routeGrps {
		topics = append(topics, t)
	}
	return topics
}

func (kr *Router) Serve() error {
	kr.stop = make(chan interface{})
	c, err := kafka.NewConsumer(kr.consumerCfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.SubscribeTopics(kr.RqTopics(), nil)
	if err != nil {
		panic(err)
	}
	log.Println("Consumer is now listening!", kr.RqTopics())

loop:
	for {
		select {

		case <-kr.stop:
			close(kr.stop)
			break loop
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				log.Println("Message received on topic (", *e.TopicPartition.Topic, ")\nkey:", string(e.Key), "\nmsg:", string(e.Value))
				kr.callHandler(e)
			case *kafka.OffsetsCommitted:
				log.Println("Offset committed: ", e)
			default:
				log.Println("Unknown event ", ev)
			}
			// default:
			// 	msg, _ := c.ReadMessage(100)
			// 	log.Println("Message received: ", msg.Key, "\n", msg.Value)
			// 	kr.callHandler(msg)

		}
	}
	return nil
}

func (kr *Router) Close() {
	kr.stop <- struct{}{}
}

func (kr *Router) callHandler(msg *kafka.Message) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, kasync.CtxKeyMsgID, string(msg.Key))

	topic := *msg.TopicPartition.Topic
	msgName := headerByKey(msg.Headers, kasync.MsgHdrMsgName)

	if msgName != kasync.MsgHdrValUnk {
		ctx = context.WithValue(ctx, kasync.CtxKeyMsgName, msgName)
	}
	h, err := kr.MsgHandler(topic, msgName)
	if err != nil {
		if errSvc, ok := err.(status.ErrServiceStatus); ok && errSvc.Is(codes.ErrNotFound) {
			kr.NotFoundHandler(ctx, msg)
		}
		kr.writeErr(msgName, msg.Key, kr.errTopic, err)
		return nil
	}

	if err := h(ctx, msg); err != nil {
		kr.writeErr(msgName, msg.Key, kr.errTopic, err)
	}
	return nil
}

func (kr *Router) MsgHandler(topic, msgName string) (MsgHandler, error) {
	rg, ok := kr.routeGrps[topic]
	if !ok {
		return nil, status.ErrBadRequest().WithMessage(fmt.Sprintf("routing group for given topic name '%s' not found", topic))
	}
	return rg.MsgHandler(msgName)
}

func (kr *Router) writeErr(msgName string, msgKey []byte, errTopic string, err error) error {
	if err == nil {
		return nil
	}
	errSvc, ok := err.(status.ErrServiceStatus)
	if !ok {
		errSvc = status.ErrInternal().WithError(err)
	}
	b, _ := json.Marshal(errSvc)

	p, err := kafka.NewProducer(kr.producerCfg)
	if err != nil {
		panic(err)
	}

	defer p.Close()

	return p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &errTopic,
			Partition: kafka.PartitionAny,
		},
		Key:   msgKey,
		Value: b,
		Headers: []kafka.Header{
			kafka.Header{Key: kasync.MsgHdrMsgName, Value: []byte(msgName)},
			kafka.Header{Key: kasync.MsgHdrMsgType, Value: []byte(kasync.MsgTypeErrEvent)},
		},
	}, nil)
}

func headerByKey(hdrs []kafka.Header, key string) kasync.MsgType {
	for _, h := range hdrs {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return kasync.MsgHdrValUnk
}
