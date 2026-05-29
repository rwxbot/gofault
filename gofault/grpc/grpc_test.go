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