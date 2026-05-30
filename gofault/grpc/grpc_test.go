package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
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