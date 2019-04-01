package httputil_test

import (
	"fmt"
	"log"
	"net/http"

	"github.com/govinda-attal/kiss-lib/core/status"
	"github.com/govinda-attal/kiss-lib/core/types"
	"github.com/govinda-attal/kiss-lib/pkg/httputil"
	"github.com/govinda-attal/kiss-lib/pkg/jwtkit"
	"github.com/gorilla/mux"
)

func ExampleNotFoundHandler_gorillaMux() {
	r := mux.NewRouter()
	// Set/Override default HTTP NotFound Handler offered by Gorilla Mux.
	r.NotFoundHandler = http.HandlerFunc(httputil.NotFoundHandler)
}

func ExampleWrapperHandler() {
	// Application specific handler(s) don't have to be defined in an embedded fashion. This is just for example.
	handler := func(w http.ResponseWriter, r *http.Request) error {
		// Notice that application specific handlers return 'error' if any,
		// this signature is different to standard http handler and intentionally simplifies error handling.
		return nil
	}
	r := mux.NewRouter()
	ex := r.PathPrefix("/hello").Subrouter()

	ex.HandleFunc("/world",
		httputil.WrapperHandler(handler)).
		Methods("GET")
}

func ExampleAuthHandler() {
	// Verifier to verify authenticity of JWT bearer token.
	v, err := jwtkit.NewRSAVerifier("path to public certificate file")

	if err != nil {
		log.Fatalf("Fatal error: %v", err)
	}

	// Application specific handler(s) don't have to be defined in an embedded fashion. This is just for example.
	handler := func(w http.ResponseWriter, r *http.Request) error {
		// This is called only when a HTTP request is submitted with a valid JWT bearer token.
		return nil
	}
	r := mux.NewRouter()
	ex := r.PathPrefix("/hello").Subrouter()

	// HTTP POST calls to /hello/world is secured with Authorization header bearer token.
	ex.HandleFunc("/world",
		httputil.WrapperHandler(httputil.AuthHandler(handler, v))).
		Methods("POST")
}

func ExampleRqBind_json() {

	// Application specific handler(s) don't have to be defined in an embedded fashion. This is just for example.
	// Following shows an example handler for HTTP Methods with VERBs that will have a JSON message within request body.
	_ = func(w http.ResponseWriter, r *http.Request) error {
		// Notice that application specific handlers return 'error' if any,
		var rq struct{ Name, Place string }
		// JSONBinder with given varible
		if err := httputil.RqBind(r, httputil.JSONBind(&rq)); err != nil {
			return err
		}
		fmt.Println("Name:", rq.Name, "\n", "Place:", rq.Place)
		return nil
	}
}

func ExampleRqBind_mpFormData() {

	// Application specific handler(s) don't have to be defined in an embedded fashion. This is just for example.
	// Following shows an example handler for HTTP Methods with VERBs that will have a multipart/form-data request	.
	_ = func(w http.ResponseWriter, r *http.Request) error {

		type Person struct {
			Name, Place string
		}
		dMap := map[string]interface{}{
			//Text
			"text": "",
			//JSON Body
			"person": &Person{},
		}
		fMap := map[string]*types.FileObj{
			"file": nil,
		}
		const _24K = (1 << 10) * 24
		if err := httputil.RqBind(r, httputil.MPFormBind(dMap, fMap, _24K)); err != nil {
			return err
		}

		txt := dMap["text"].(string)  // String
		p := dMap["person"].(*Person) // Person Struct
		fObj := fMap["file"]          // file raw bytes

		b := make([]byte, fObj.Size())
		if _, err := fObj.Reader().Read(b); err != nil {
			return err
		}
		// Prints multipart contents to the terminal
		fmt.Println(txt, p, b)
		return httputil.RsRender(w, httputil.JSONRend(status.Success()))
	}
}

func ExampleRsRender_json() {

	_ = func(w http.ResponseWriter, r *http.Request) error {
		var rs struct{ Name, Place string }
		rs.Name = "John Doe"
		rs.Place = "Sydney"
		// A JSON HTTP Response with default 200 status code.
		return httputil.RsRender(w, httputil.JSONRend(&rs))
	}
}

func ExampleRsRenderWithStatus_json() {

	_ = func(w http.ResponseWriter, r *http.Request) error {
		var rs struct{ Name, Place string }
		rs.Name = "John Doe"
		rs.Place = "Sydney"
		// A JSON HTTP Response with 302 status code.
		return httputil.RsRenderWithStatus(w, httputil.JSONRend(&rs), http.StatusFound)
	}
}
