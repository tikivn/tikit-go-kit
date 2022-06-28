package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/tikivn/tikit-go-kit/grpc/gatewayopt"
	"github.com/tikivn/tikit-go-kit/l"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

// gatewayServer wraps gRPC gateway server setup process.
type gatewayServer struct {
	// Mux *runtime.ServeMux
	// listener *net.Listener
	// mux    *http.Handler
	server *http.Server
	config *gatewayConfig
}

type HTTPServerConfig struct {
	TLSConfig         *tls.Config
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	TLSNextProto      map[string]func(*http.Server, *tls.Conn, http.Handler)
	ConnState         func(net.Conn, http.ConnState)
}

func (c *HTTPServerConfig) applyTo(s *http.Server) {
	s.TLSConfig = c.TLSConfig
	s.ReadTimeout = c.ReadTimeout
	s.ReadHeaderTimeout = c.ReadHeaderTimeout
	s.WriteTimeout = c.WriteTimeout
	s.IdleTimeout = c.IdleTimeout
	s.MaxHeaderBytes = c.MaxHeaderBytes
	s.TLSNextProto = c.TLSNextProto
	s.ConnState = c.ConnState
}

type gatewayConfig struct {
	Addr              Listen
	MuxOptions        []runtime.ServeMuxOption
	ServerConfig      *HTTPServerConfig
	ServerMiddlewares []HTTPServerMiddleware
	ServerHandlers    []HTTPServerHandler
}

func createDefaultGatewayConfig() *gatewayConfig {
	config := &gatewayConfig{
		Addr: Listen{
			Host: "0.0.0.0",
			Port: 10080,
		},
		MuxOptions: []runtime.ServeMuxOption{
			gatewayopt.ProtoJSONMarshaler(),
			runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		},
		ServerHandlers: []HTTPServerHandler{
			PrometheusHandler,
			PprofHandler,
		},
	}

	return config
}

func newGatewayServer(c *gatewayConfig, conn *grpc.ClientConn, servers []ServiceServer) (*gatewayServer, error) {
	// init mux
	mux := runtime.NewServeMux(c.MuxOptions...)

	var handler http.Handler = mux

	//handler = otelhttp.NewHandler(handler, "")

	for i := len(c.ServerMiddlewares) - 1; i >= 0; i-- {
		handler = c.ServerMiddlewares[i](handler)
	}

	httpMux := http.NewServeMux()

	for _, h := range c.ServerHandlers {
		h(httpMux)
	}

	httpMux.Handle("/", handler)

	svr := &http.Server{
		Addr:    c.Addr.String(),
		Handler: httpMux,
	}
	if cfg := c.ServerConfig; cfg != nil {
		cfg.applyTo(svr)
	}

	for _, svr := range servers {
		err := svr.RegisterWithHandler(context.Background(), mux, conn)
		if err != nil {
			return nil, fmt.Errorf("failed to register handler. %w", err)
		}
	}

	return &gatewayServer{
		server: svr,
		// mux:    &httpMux,
		config: c,
	}, nil
}

// Serve
func (s *gatewayServer) Serve() error {

	ll.Info("http server starting at", l.String("addr", s.config.Addr.String()))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		ll.Info("Error starting http server, ", l.Error(err))
		return err
	}

	return nil
}

func (s *gatewayServer) Shutdown(ctx context.Context) {
	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// defer cancel()
	err := s.server.Shutdown(ctx)
	ll.Info("All http(s) requests finished")
	if err != nil {
		ll.Info("failed to shutdown grpc-gateway server: ", l.Error(err))
	}
}
