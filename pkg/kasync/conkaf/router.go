package conkaf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/kasync"
)

func New(consumerCfg, producerCfg *kafka.ConfigMap, errTopic string) *Router {
	return &Router{
		consumerCfg: consumerCfg,
		producerCfg: producerCfg,
		errTopic:    errTopic,
		routeGrps:   make(map[string]*TopicRouteGroup),
	}
}

type Router struct {
	io.Closer
	consumerCfg, producerCfg *kafka.ConfigMap
	routeGrps                map[string]*TopicRouteGroup
	errTopic                 string
	stop                     chan interface{}
}

type TopicRouteGroup struct {
	topic       string
	defHandler  kasync.MsgHandler
	msgResolver kasync.ResolveMsgName
	handlers    map[string]kasync.MsgHandler
}

//MsgHandler returns the handler for the route.
func (rg *TopicRouteGroup) MsgHandler(msgName string) (kasync.MsgHandler, error) {
	h, ok := rg.handlers[msgName]
	if !ok {
		if rg.defHandler == nil {
			return nil, status.ErrNotFound().WithMessage(fmt.Sprintf("handler for given message %s not found", msgName))
		}
		return rg.defHandler, nil
	}
	return h, nil
}

func (rg *TopicRouteGroup) HandleMsg(msgName string, handler kasync.MsgHandler) {
	rg.handlers[msgName] = handler
}

func (rg *TopicRouteGroup) SetMsgNameResolver(r kasync.ResolveMsgName) {
	rg.msgResolver = r
}

func (rg *TopicRouteGroup) ResolveMsgName(msg interface{}) (string, error) {
	return rg.msgResolver(msg)
}

func (r *Router) NewRouteGrp(topic string, defHandler kasync.MsgHandler) kasync.RouteGroup {
	rg := &TopicRouteGroup{topic: topic,
		msgResolver: ResolveMsgName,
		defHandler:  defHandler,
		handlers:    make(map[string]kasync.MsgHandler)}
	r.routeGrps[topic] = rg
	return rg
}

func (r *Router) RqTopics() []string {
	var topics []string
	for t := range r.routeGrps {
		topics = append(topics, t)
	}
	return topics
}

func (r *Router) Listen() error {
	r.stop = make(chan interface{})
	c, err := kafka.NewConsumer(r.consumerCfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = c.SubscribeTopics(r.RqTopics(), nil)
	if err != nil {
		panic(err)
	}
	log.Println("Consumer is now listening!", r.RqTopics())

loop:
	for {
		select {

		case <-r.stop:
			close(r.stop)
			break loop
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				log.Println("Message received on topic (", *e.TopicPartition.Topic, ")\nkey:", string(e.Key), "\nmsg:", string(e.Value))
				r.callHandler(e)
			case *kafka.OffsetsCommitted:
				log.Println("Offset committed: ", e)
			default:
				log.Println("Unknown event ", ev)
			}
		}
	}
	return nil
}

func (r *Router) Close() error {
	r.stop <- struct{}{}
	return nil
}

func (r *Router) callHandler(msg *kafka.Message) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, kasync.CtxKeyMsgID, string(msg.Key))

	topic := *msg.TopicPartition.Topic

	rg, err := r.RouteGroup(topic)
	if err != nil {
		r.writeErr(kasync.MsgTypeUnk, msg.Key, r.errTopic, err)
		return err
	}

	msgName, err := rg.ResolveMsgName(msg)

	if err != nil {
		r.writeErr(kasync.MsgTypeUnk, msg.Key, r.errTopic, err)
		return err
	}

	if msgName != kasync.MsgHdrValUnk {
		ctx = context.WithValue(ctx, kasync.CtxKeyMsgName, msgName)
	}

	h, err := rg.MsgHandler(msgName)
	if err != nil {
		r.writeErr(msgName, msg.Key, r.errTopic, err)
		return err
	}

	if err := h(ctx, msg.Value); err != nil {
		r.writeErr(msgName, msg.Key, r.errTopic, err)
		return err
	}
	return nil
}

func (r *Router) RouteGroup(topic string) (kasync.RouteGroup, error) {
	rg, ok := r.routeGrps[topic]
	if !ok {
		return nil, status.ErrBadRequest().WithMessage(fmt.Sprintf("routing group for given topic name '%s' not found", topic))
	}
	return rg, nil
}

func (r *Router) MsgHandler(topic, msgName string) (kasync.MsgHandler, error) {
	rg, ok := r.routeGrps[topic]
	if !ok {
		return nil, status.ErrBadRequest().WithMessage(fmt.Sprintf("routing group for given topic name '%s' not found", topic))
	}
	return rg.MsgHandler(msgName)
}

func (r *Router) writeErr(msgName string, msgKey []byte, errTopic string, err error) error {
	if err == nil {
		return nil
	}
	errSvc, ok := err.(status.ErrServiceStatus)
	if !ok {
		errSvc = status.ErrInternal().WithError(err)
	}
	b, _ := json.Marshal(errSvc)

	p, err := kafka.NewProducer(r.producerCfg)
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

func ResolveMsgName(msg interface{}) (string, error) {
	return headerByKey(msg.(*kafka.Message).Headers, kasync.MsgHdrMsgName), nil
}
