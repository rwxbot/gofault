package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"runtime/debug"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"github.com/gofault/gofault/core"
)

// Config holds gRPC server configuration.
type Config struct {
	// Port to listen on (default 9000).
	Port int
	// Network type (tcp, tcp4, tcp6).
	Network string
	// KeepAlive server parameters.
	KeepAlive *keepalive.ServerParameters
	// MaxConcurrentStreams sets the limit on the number of concurrent streams.
	MaxConcurrentStreams uint32
	// Insecure skips TLS (default false).
	Insecure bool
	// TLSCert and TLSKey for TLS credentials.
	TLSCert string
	TLSKey  string
	// UnaryInterceptors registers unary server interceptors.
	UnaryInterceptors []grpc.UnaryServerInterceptor
	// StreamInterceptors registers stream server interceptors.
	StreamInterceptors []grpc.StreamServerInterceptor
}

// DefaultConfig returns a default gRPC configuration.
func DefaultConfig() Config {
	return Config{
		Port:                 9000,
		Network:              "tcp",
		MaxConcurrentStreams: 100,
	}
}

// Server represents a gRPC server.
type Server struct {
	config     Config
	grpcServer *grpc.Server
	listener   net.Listener
	mu         sync.Mutex
	started    bool
}

// NewServer creates a new gRPC server.
func NewServer(config Config) *Server {
	if config.Port == 0 {
		config.Port = 9000
	}
	if config.Network == "" {
		config.Network = "tcp"
	}

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
	}

	// Build interceptors chain
	if len(config.UnaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(config.UnaryInterceptors...))
	}
	if len(config.StreamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(config.StreamInterceptors...))
	}

	if config.KeepAlive != nil {
		opts = append(opts, grpc.KeepaliveParams(*config.KeepAlive))
	}

	var creds credentials.TransportCredentials
	if config.Insecure {
		creds = insecure.NewCredentials()
	} else if config.TLSCert != "" && config.TLSKey != "" {
		var err error
		creds, err = credentials.NewServerTLSFromFile(config.TLSCert, config.TLSKey)
		if err != nil {
			panic(fmt.Sprintf("failed to load TLS certs: %v", err))
		}
	}

	if creds != nil {
		opts = append(opts, grpc.Creds(creds))
	}

	return &Server{
		config:     config,
		grpcServer: grpc.NewServer(opts...),
	}
}

// RegisterService registers a gRPC service.
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.grpcServer.RegisterService(desc, impl)
}

// Start listening on the configured port.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Network, s.config.Port)
	ln, err := net.Listen(s.config.Network, addr)
	if err != nil {