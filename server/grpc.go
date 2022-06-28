package server

import (
	"context"
	"fmt"

	"github.com/tikivn/tikit-go-kit/l"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

type grpcConfig struct {
	Addr                     Listen
	ServerUnaryInterceptors  []grpc.UnaryServerInterceptor
	ServerStreamInterceptors []grpc.StreamServerInterceptor
	ServerOption             []grpc.ServerOption
	MaxConcurrentStreams     uint32
}

func createDefaultGrpcConfig() *grpcConfig {
	// TODO: create interface to add option to logger
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.EnableClientHandlingTimeHistogram()
	config := &grpcConfig{
		Addr: Listen{
			Host: "0.0.0.0",
			Port: 10443,
		},
		ServerUnaryInterceptors: []grpc.UnaryServerInterceptor{
			//grpcotel.UnaryServerInterceptor(global.Tracer("grpc-unary")),
			grpc_prometheus.UnaryServerInterceptor,
			//grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			//grpc_ctx.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			grpc_zap.PayloadUnaryServerInterceptor(ll.Logger, func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
				if fullMethodName == "/pb.HealthService/Liveness" {
					return false
				}
				if fullMethodName == "/pb.HealthService/Readiness" {
					return false
				}
				return true
			}),
		},
		ServerStreamInterceptors: []grpc.StreamServerInterceptor{
			//grpcotel.StreamServerInterceptor(global.Tracer("grpc-stream")),
			grpc_prometheus.StreamServerInterceptor,
			//grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			//grpc_ctx.StreamServerInterceptor(),
			grpc_validator.StreamServerInterceptor(),
		},

		MaxConcurrentStreams: 1000,
	}

	return config
}

func (c *grpcConfig) ServerOptions() []grpc.ServerOption {
	return append(
		[]grpc.ServerOption{
			grpc_middleware.WithUnaryServerChain(c.ServerUnaryInterceptors...),
			grpc_middleware.WithStreamServerChain(c.ServerStreamInterceptors...),
			grpc.MaxConcurrentStreams(c.MaxConcurrentStreams),
		},
		c.ServerOption...,
	)
}

// grpcServer wraps grpc.Server setup process.
type grpcServer struct {
	server *grpc.Server
	config *grpcConfig
}

func newGrpcServer(c *grpcConfig, servers []ServiceServer) *grpcServer {
	s := grpc.NewServer(c.ServerOptions()...)
	for _, svr := range servers {
		svr.RegisterWithServer(s)
	}
	return &grpcServer{
		server: s,
		config: c,
	}
}

// Serve implements Server.Server
func (s *grpcServer) Serve() error {
	listener, err := s.config.Addr.CreateListener()
	if err != nil {
		return fmt.Errorf("failed to create listener %w", err)
	}
	ll.Info("gRPC server is starting ", l.String("addr", listener.Addr().String()))

	if err = s.server.Serve(listener); err != nil {
		ll.Info("while serving", l.Error(err))
		return fmt.Errorf("failed to serve gRPC server %w", err)
	}
	ll.Info("gRPC server ready")

	return nil
}

// Shutdown
func (s *grpcServer) Shutdown(ctx context.Context) {
	s.server.GracefulStop()
}
