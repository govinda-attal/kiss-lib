package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/kasync/krouter"
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

	r := krouter.New(cCfg, pCfg, "errTopic")

	r.NewRouteGrp("Greetings", h.DefaultHandler).
		Invoke("Welcome", h.Welcome).
		Invoke("Farewell", h.Farewell)

	// r.NewRouteGrp("AnotherTopic", h.DefaultHandler)

	go func() {
		if err := r.Serve(); err != nil {
			log.Fatalln(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
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

func (gh *GreeterHandler) Welcome(ctx context.Context, msg *kafka.Message) error {
	var rq WelcomeRq
	err := json.Unmarshal(msg.Value, &rq)
	if err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	return gh.srv.Welcome(rq)
}

func (gh *GreeterHandler) Farewell(ctx context.Context, msg *kafka.Message) error {
	var rq FarewellRq
	err := json.Unmarshal(msg.Value, &rq)
	if err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	return gh.srv.Farewell(rq)
}

func (gh *GreeterHandler) DefaultHandler(ctx context.Context, msg *kafka.Message) error {
	log.Println("Message processed by default handler: ", "\nmsg:", string(msg.Value))
	return nil
}
