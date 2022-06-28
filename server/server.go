package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tikivn/tikit-go-kit/l"

	"google.golang.org/grpc"
)

var (
	isShuttingDown = false
)

var ll = l.New()

// Server is the framework instance.
type Server struct {
	grpcServer    *grpcServer
	gatewayServer *gatewayServer
	config        *Config
}

// New creates a server intstance.
func New(opts ...Option) (*Server, error) {
	c := createConfig(opts)

	ll.Info("Create grpc server")
	grpcServer := newGrpcServer(c.Grpc, c.ServiceServers)
	// if err != nil {
	// 	return nil, fmt.Errorf("Faild to create grpc server. %w", err)
	// }

	conn, err := grpc.Dial(c.Grpc.Addr.String(), grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*50)),
		grpc.WithChainUnaryInterceptor(),
	)
	if err != nil {
		return nil, fmt.Errorf("fail to dial gRPC server. %w", err)
	}

	ll.Info("Create gateway server")
	gatewayServer, err := newGatewayServer(c.Gateway, conn, c.ServiceServers)
	if err != nil {
		return nil, fmt.Errorf("fail to create gateway server. %w", err)
	}

	return &Server{
		grpcServer:    grpcServer,
		gatewayServer: gatewayServer,
		config:        c,
	}, nil
}

// Start starts gRPC and Gateway servers.
func (s *Server) Start() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.gatewayServer.Serve(); err != nil {
			ll.Error("Error starting http server, ", l.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.grpcServer.Serve(); err != nil {
			ll.Error("Error starting gRPC server, ", l.Error(err))
		}
	}()

	wg.Wait()
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, ss := range s.config.ServiceServers {
		ss.Close(ctx)
	}

	s.gatewayServer.Shutdown(ctx)
	s.grpcServer.Shutdown(ctx)
}

// Serve starts gRPC and Gateway servers.
func (s *Server) Serve() error {
	stop := make(chan os.Signal, 1)
	errch := make(chan error)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.gatewayServer.Serve(); err != nil {
			ll.Error("Error starting http server, ", l.Error(err))
			errch <- err
		}
	}()

	go func() {
		if err := s.grpcServer.Serve(); err != nil {
			ll.Error("Error starting gRPC server, ", l.Error(err))
			errch <- err
		}
	}()

	// shutdown
	for {
		select {
		case <-stop:
			ll.Info("Shutting down server")

			isShuttingDown = true

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			for _, ss := range s.config.ServiceServers {
				ss.Close(ctx)
			}

			s.gatewayServer.Shutdown(ctx)
			s.grpcServer.Shutdown(ctx)
			return nil
		case err := <-errch:
			return err
		}
	}
}

func IsServerShuttingDown() bool {
	return isShuttingDown
}
