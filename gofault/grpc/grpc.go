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