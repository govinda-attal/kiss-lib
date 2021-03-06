# Keep it Simple Stupid - Library

This library provides simple utilities that help to write less but meanigful code.

## Log(rus)

At times if you feel that passing custom logger instance is too much of instrumentation code (if yes, then continue reading this section!)

If so then you may agree that functional concrete implementation don't need to hold a reference to custom logger.

Would it be easy to simply write (and later read code) as below:

```go
package valuable

import (
	// writing less code is more (readable)!
	// writing logrus everytime we want to just write log statements is not helpful either!
	log "github.com/sirupsen/logrus"
)

// Operation is example valuable operation that logs valuable information
func Operation() {
	// if you notice here we didn't inject a special logger instance
	// less code is more (readable)!
	log.Infoln("this is a valuable message from a valuable operation")
}

```

If you like this, quick way to enable this is by importing a package to register/initialize default logger for logrus

```go
package main

import (
	log "github.com/sirupsen/logrus"

	// Register the default logger with accpeted default settings
	// Best done in main package at bootstrap.
	_ "github.com/govinda-attal/kiss-lib/pkg/logrus/reglog"

	"github.com/govinda-attal/kiss-lib/test/logex/valuable"
)

func main() {
	// One of based on configuration, you could override default settings
	// For example the level
	log.SetLevel(log.DebugLevel)

	log.WithFields(
		log.Fields{
			"go": "less code is more readable!",
		},
	).Infoln("this is info message")

	log.Debugln("this is debug message")

	// this logs something valuable.
	// if you notice there less instrumentation and we didn't pass custom logger instance!
	// less is more!
	valuable.Operation()
}
```

## HTTP RESTful service

A restful service is good when has following layers

1. HTTP Handler layer - responsible for protocol transalation
2. Business Service layer - responsible for real business logic devoid of how it is exposed - http or kafka, etc

Here Business layer when instantiated can be injected with dependencies like db connection
And when HTTP handler layer is instantiated - it can be injected with dependency being business layer instance

HTTP handlers should look simple for error handling and request and response data mapping. Hence a utility is written in github.com/govinda-attal/kiss-lib/pkg/httputil package.

Gorilla Mux is used for HTTP Router.
If there are any decorator's to applied to individual resource path handlers the can be injected too.

### Business Layer 
Business layer devoid of HTTP translations would look like

```go
// Greeter is the real interface that depicts business service contract.
type Greeter interface {
	// Hello returns a personalised greeting message for given argument.
	Hello(ctx context.Context, name string) (msg string, err error)
}

// srv implements Greeter interface.
type srv struct{}

func NewImpl() *srv {
	return &srv{}
}

func (s *srv) Hello(ctx context.Context, name string) (string, error) {
	msg := "Hello " + name
	return msg, nil
}

```

### HTTP Handler
HTTP Handler will manage http translations, error handling is simplified too.
```go
type restHandler struct {
	g Greeter
}

// NewHandler returns concreate handler implementation with dependency of real implementation being injected.
func NewHandler(g Greeter) *restHandler {
	return &restHandler{g}
}


func (rh *restHandler) Hello(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	name := vars["name"]

	msg, err := rh.g.Hello(r.Context(), name)
	if err != nil {
		return err
	}

	rs := status.NewUserDefined(codes.Success, msg)
	return httputil.RsRender(w, httputil.JSONRend(&rs))
}

func (rh *restHandler) Error(w http.ResponseWriter, r *http.Request) error {
	return status.ErrInternal()
}

```

### Bootstrap - Cobra and Viper

With httputil, error handling is simplified along with function decorators per handler func.
Negroni middlewares could be used that apply across mux or subrouter.

```go
var rootCmd = &cobra.Command{
	Use:   "restex",
	Short: "Starts microservice",
	Run:   startServer,
}

func registerHandler(r *mux.Router) {
	rh := NewHandler(NewImpl())

	ex := r.PathPrefix("/ex").Subrouter()
	ex.HandleFunc("/hello/{name}",
		httputil.WrapperHandler(rh.Hello /*, optional decorators */)).
		Methods("GET")
	ex.HandleFunc("/error",
		httputil.WrapperHandler(rh.Error)).
		Methods("GET")
	ex.HandleFunc("/secured/{name}",
		httputil.WrapperHandler(rh.Hello, httputil.AuthDecorator(nil))).
		Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(httputil.NotFoundHandler)
}

func startServer(cmd *cobra.Command, args []string) {
	r := mux.NewRouter()
	registerHandler(r)

	h := cors.Default().Handler(r)
	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.UseHandler(h)

	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: n,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)
	os.Exit(0)
}
```

### A Router pattern to route Kafka Messages on one or more topic(s)

This pattern is inspired from frameworks like gorilla mux router but is intended for message processing on Kafka topics.

Though, I have used Confluent Library, one can achieve similar pattern by just implementing driver interfaces in kiss-lib/pkg/kasync.

The crux boils down to:
```go
	// A router can listen to many topics, hence routing groups
	rg := r.NewRouteGrp("Greetings", h.DefaultHandler)
	// For different messages (in different formats) that arrive on the same topic, you can specify custom handler
	// By the way this step is optional if you want to have a default handler for a topic
	rg.HandleMsg("Welcome", h.Welcome)
	rg.HandleMsg("Farewell", h.Farewell)
```

A more detailed implementation is as below

```go 
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

```
