package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/core/status/codes"
	"github.com/govinda-attal/kiss-lib/pkg/httputil"
	_ "github.com/govinda-attal/kiss-lib/pkg/logrus/reglog"
)

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

var (
	version string = "1.0.0"
	port    int
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.yaml", "config file (default is ./config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv()
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infoln("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	log.Println("restex version:", version)
	Execute()
}
