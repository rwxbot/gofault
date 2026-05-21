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

// bind reserves the configured listener. The address passed to net.Listen
// must be host:port — the old code built "tcp:9000" from the *network* and
// passed it as the address, so Go tried to resolve a host named "tcp" and
// every default-config Start failed.
func (s *Server) bind() (net.Listener, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil, fmt.Errorf("server already started")
	}
	if err := s.validateTLS(); err != nil {
		return nil, err
	}

	addr := fmt.Sprintf(":%d", s.config.Port)
	ln, err := net.Listen(s.config.Network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = ln
	s.started = true
	return ln, nil
}

// Start listening on the configured port. It blocks until the server stops.
func (s *Server) Start() error {
	ln, err := s.bind()
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(ln)
}

// StartInBackground binds the listener synchronously — so bind failures are
// reported to the caller — and then serves on its own goroutine.
// GRPCServerModule.OnBoot must use this: the module lifecycle runs OnBoot
// hooks sequentially before starting the HTTP server, and a blocking Start
// there hangs App.Start forever.
func (s *Server) StartInBackground() error {
	ln, err := s.bind()
	if err != nil {
		return err
	}
	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
			slog.Default().Error("grpc server stopped with error", "error", err)
		}
	}()
	return nil
}

// validateTLS rejects a half-configured TLS setup. NewServer only builds
// credentials when BOTH TLSCert and TLSKey are set; with exactly one set it
// silently leaves creds nil, and the server would serve plaintext while the
// operator believes the endpoint is encrypted. Fail at bind time instead.
// Callers hold s.mu or run before the server is shared; config is fixed at
// construction.
func (s *Server) validateTLS() error {
	if s.config.Insecure {
		return nil
	}
	if (s.config.TLSCert == "") != (s.config.TLSKey == "") {
		return fmt.Errorf("grpc: TLS is half-configured: TLSCert and TLSKey must both be set (or Insecure must be true); refusing to serve plaintext")
	}
	return nil
}

// Serve starts the gRPC server on the given listener.
func (s *Server) Serve(ln net.Listener) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return fmt.Errorf("server already started")
	}
	if err := s.validateTLS(); err != nil {
		s.mu.Unlock()
		return err
	}
	s.listener = ln
	s.started = true
	s.mu.Unlock()

	return s.grpcServer.Serve(ln)
}

// Stop gracefully shuts down the gRPC server.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}

// ImmediateStop immediately shuts down the gRPC server.
func (s *Server) ImmediateStop() {
	s.grpcServer.Stop()
}

// Address returns the listening address, or empty string if not started.
func (s *Server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// logRPCPanic records a recovered panic. Every interceptor here must convert a
// panic into an error: recovering without returning one leaves the named return
// values at their zero value, so gRPC would report the call as a success with a
// nil response and no error.
func logRPCPanic(method, kind string, r interface{}) {
	slog.Default().Error("grpc panic recovered",
		"kind", kind, "method", method, "panic", r, "stack", string(debug.Stack()))
}

// panicError converts a recovered panic into a gRPC error.
func panicError(r interface{}) error {
	return status.Errorf(codes.Internal, "panic recovered: %v", r)
}

// methodName extracts the full method name, tolerating a nil info.
func methodName(fullMethod string) string { return fullMethod }

// UnaryInterceptor adapts a function to grpc.UnaryServerInterceptor.
//
// The adapter recovers from panics in fn and reports them as Internal errors.
// It intentionally does not invoke the next handler: fn is the whole operation.
// Chain RecoveryInterceptor or LoggingInterceptor separately when they are
// needed alongside it.
func UnaryInterceptor(fn func(ctx context.Context, req interface{}) (interface{}, error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logRPCPanic(methodName(infoFullMethod(info)), "unary", r)
				resp, err = nil, panicError(r)
			}
		}()
		return fn(ctx, req)
	}
}

// StreamInterceptor adapts a function to grpc.StreamServerInterceptor.
func StreamInterceptor(fn func(srv interface{}, ss grpc.ServerStream) error) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("grpc stream panic recovered: %v\n", r)
			}
		}()
		return fn(srv, ss)
	}
}

// RecoveryInterceptor returns a unary interceptor that recovers from panics.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		defer func() {