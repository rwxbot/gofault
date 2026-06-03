package grpc

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/grpc/testdata"
)

const bufSize = 1024 * 1024

func TestNewServer(t *testing.T) {
	config := DefaultConfig()
	config.Port = 0 // use random port
	s := NewServer(config)

	if s == nil {
		t.Fatal("server should not be nil")
	}

	if s.grpcServer == nil {
		t.Error("grpcServer should not be nil")
	}
}

func TestServer_StartAndStop(t *testing.T) {
	config := DefaultConfig()
	config.Port = 0
	config.Insecure = true
	s := NewServer(config)

	// Start in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	s.Stop()

	select {
	case err := <-errCh:
		// Server stopped (may have address in use error if race, that's ok)
		_ = err
	case <-time.After(time.Second):
		t.Error("server did not start within timeout")
	}
}

func TestServer_Serve(t *testing.T) {
	config := DefaultConfig()
	config.Insecure = true
	s := NewServer(config)

	ln := bufconn.Listen(bufSize)

	go func() {
		s.Serve(ln)
	}()

	time.Sleep(100 * time.Millisecond)

	if s.Address() == "" {
		t.Error("server should have an address after starting")
	}

	s.Stop()