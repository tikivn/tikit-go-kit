package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

type Config struct {
	Address             string `yaml:"address" mapstructure:"address"`
	LoadBalancingPolicy string `yaml:"load_balancing_policy" mapstructure:"load_balancing_policy"`
	ServerName          string `yaml:"server_name" mapstructure:"server_name"`
	Retries             uint   `yaml:"retries" mapstructure:"retries"`
}

func DefaultConfig() *Config {
	return &Config{
		Address:             "",
		Retries:             3,
		LoadBalancingPolicy: roundrobin.Name,
	}
}

func NewConnection(cfg Config, opts ...grpc.DialOption) (cc grpc.ClientConnInterface, err error) {
	var loadBalancingPolicy = cfg.LoadBalancingPolicy
	if len(loadBalancingPolicy) == 0 {
		loadBalancingPolicy = roundrobin.Name
	}
	rrLoadBalancing := fmt.Sprintf(`{"loadBalancingPolicy":"%s"}`, loadBalancingPolicy)

	var dialOpts = make([]grpc.DialOption, 0)
	if len(cfg.ServerName) > 0 {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName: cfg.ServerName,
		})))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	dialOpts = append(dialOpts,
		grpc.WithDefaultServiceConfig(rrLoadBalancing),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
		),
	)

	if cfg.Retries > 0 {
		retryOpts := []grpc_retry.CallOption{
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
			grpc_retry.WithCodes(codes.Aborted, codes.Internal, codes.ResourceExhausted),
			grpc_retry.WithMax(cfg.Retries),
		}
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(retryOpts...),
		))
	}

	if len(opts) > 0 {
		dialOpts = append(dialOpts, opts...)
	}
	cc, err = grpc.DialContext(context.Background(), cfg.Address,
		dialOpts...,
	)

	return
}
