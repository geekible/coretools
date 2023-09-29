package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog"
)

type HttpMethod int

const (
	_ HttpMethod = iota
	GET
	POST
	PUT
	DELETE
)

type httpServer struct {
	mux         *chi.Mux
	servicePort int
}

type HttpServerBuilder struct {
	server *httpServer
}

func NewHttpServerBuilder() *HttpServerBuilder {
	return &HttpServerBuilder{&httpServer{
		mux: chi.NewMux(),
	}}
}

func (hsb *HttpServerBuilder) SetPort(port int) *HttpServerBuilder {
	hsb.server.servicePort = port
	return hsb
}

func (hsb *HttpServerBuilder) SetHttpLogging(serviceName string) *HttpServerBuilder {
	requestLogger := httplog.NewLogger(serviceName, httplog.Options{
		JSON:            true,
		Concise:         true,
		LogLevel:        "debug",
		TimeFieldFormat: time.RFC3339Nano,
		TimeFieldName:   "event_time",
	})
	hsb.server.mux.Use(httplog.RequestLogger(requestLogger))
	return hsb
}

func (hsb *HttpServerBuilder) SetCompression() *HttpServerBuilder {
	hsb.server.mux.Use(middleware.Compress(5, "application/json"))
	return hsb
}

func (hsb *HttpServerBuilder) SetCorsPolicy() *HttpServerBuilder {
	hsb.server.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	return hsb
}

func (hsb *HttpServerBuilder) AddHandler(pattern string, httpMethod HttpMethod, handlerFunc http.HandlerFunc) *HttpServerBuilder {
	switch httpMethod {
	case GET:
		hsb.server.mux.Get(pattern, handlerFunc)
	case POST:
		hsb.server.mux.Post(pattern, handlerFunc)
	case PUT:
		hsb.server.mux.Put(pattern, handlerFunc)
	case DELETE:
		hsb.server.mux.Delete(pattern, handlerFunc)
	}
	return hsb
}

func (hsb *HttpServerBuilder) AddHeartbeat() *HttpServerBuilder {
	hsb.server.mux.Use(middleware.Heartbeat("/"))
	return hsb
}

func (hsb *HttpServerBuilder) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	svr := http.Server{
		Addr:    fmt.Sprintf(":%d", hsb.server.servicePort),
		Handler: hsb.server.mux,
	}

	go func() {
		fmt.Fprintf(os.Stdout, "starting http server on port %d\n", hsb.server.servicePort)
		log.Println(svr.ListenAndServe())
		cancel()
	}()

	go func() {
		fmt.Fprint(os.Stdout, "service started. press any key to quit...\n")
		var s string
		fmt.Scanln(&s)
		svr.Shutdown(ctx)
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("service shutdown gracefully")
}
