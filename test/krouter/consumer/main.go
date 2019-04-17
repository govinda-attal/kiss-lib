package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/kasync"
	"github.com/govinda-attal/kiss-lib/pkg/kasync/conkaf"
)

func main() {

	pCfg := &kafka.ConfigMap{"bootstrap.servers": "localhost"}
	cCfg := &kafka.ConfigMap{
		"bootstrap.servers":     "localhost",
		"broker.address.family": "v4",
		"group.id":              "group",
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest",
	}

	greeter := &Greeter{}
	h := NewGreeterHandler(greeter)

	// Get a new Kafka Router (here sample router is provided that uses confluent kafka go library)
	// technically you could write a router with any library that implements interfaces within the driver package
	// kiss-lib/pkg/kasync

	r := conkaf.New(cCfg, pCfg, "errTopic")

	// A router can listen to many topics, hence routing groups
	rg := r.NewRouteGrp("Greetings", h.DefaultHandler)
	// For different messages (in different formats) that arrive on the same topic, you can specify custom handler
	// By the way this step is optional if you want to have a default handler for a topic
	rg.HandleMsg("Welcome", h.Welcome)
	rg.HandleMsg("Farewell", h.Farewell)

	// For custom requirements you can have custom message name resolver
	// It can be as simple as looking to kafka message header or looking into content of the message
	customMsgNameResolver := func(msg interface{}) (string, error) {
		if _, ok := msg.(*kafka.Message); !ok {
			return kasync.MsgHdrValUnk, fmt.Errorf("invalid msg type := does the message type match the kafka library you are using")
		}
		for _, h := range msg.(*kafka.Message).Headers {
			if h.Key == kasync.MsgHdrMsgName {
				return string(h.Value), nil
			}
		}
		return kasync.MsgHdrValUnk, nil
	}
	// Set this custom message name resolver at the routing group level or per topic level
	rg.SetMsgNameResolver(customMsgNameResolver)

	// follow same steps if the router needs to listen to same topic
	// r.NewRouteGrp("AnotherTopic", h.DefaultHandler)

	go func() {
		// Router will now start consuming or listening to messages on topics
		if err := r.Listen(); err != nil {
			log.Fatalln(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	// safely close the router (active connections to multiple topics)
	r.Close()

	time.Sleep(5 * time.Second)
}

type Greeter struct{}

func (g Greeter) Welcome(rq WelcomeRq) error {
	log.Println("welcome pack sent to: ", rq.Name)
	return nil
}

func (g Greeter) Farewell(rq FarewellRq) error {
	log.Println("farewell wishes sent to: ", rq.Name)
	return nil
}

type Greetings interface {
	Welcome(rq WelcomeRq) error
	Farewell(rq FarewellRq) error
}

type WelcomeRq struct {
	Name string `json:"name"`
}
type FarewellRq struct {
	Name string `json:"name"`
}

type GreeterHandler struct {
	srv *Greeter
}

func NewGreeterHandler(g *Greeter) *GreeterHandler {
	return &GreeterHandler{srv: g}
}

func (gh *GreeterHandler) Welcome(ctx context.Context, data []byte) error {
	var rq WelcomeRq
	err := json.Unmarshal(data, &rq)
	if err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	return gh.srv.Welcome(rq)
}

func (gh *GreeterHandler) Farewell(ctx context.Context, data []byte) error {
	var rq FarewellRq
	err := json.Unmarshal(data, &rq)
	if err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	return gh.srv.Farewell(rq)
}

func (gh *GreeterHandler) DefaultHandler(ctx context.Context, data []byte) error {
	log.Println("Message processed by default handler: ", "\nmsg payload:", string(data))
	return nil
}
