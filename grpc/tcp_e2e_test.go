package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Every other end-to-end test runs over bufconn, which bypasses the kernel
// socket path entirely. This test puts a real TCP listener (127.0.0.1:0) in
// front of the server and dials it with a real client, so the wire format,
// the HTTP/2 handshake, and Address() reporting all run over an actual
// socket. It also guards the bind() address fix: the old code passed
// "tcp:9000" as the *address*, which no real dialer could ever reach.
func TestServer_RealTCPUnary(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("real TCP listen: %v", err)
	}

	cfg := DefaultConfig()
	cfg.Insecure = true
	cfg.UnaryInterceptors = []grpc.UnaryServerInterceptor{RecoveryInterceptor(), LoggingInterceptor()}
	s := NewServer(cfg)
	s.RegisterService(&EchoServiceDesc, echoServiceImpl{})

	served := make(chan error, 1)
	go func() { served <- s.Serve(ln) }()
	t.Cleanup(func() {
		s.ImmediateStop()
		<-served
	})

	// Address() must report the real socket address, not a configured
	// guess. Serve runs on its own goroutine, so wait for the listener
	// to be recorded.
	var addr string
	deadline := time.Now().Add(2 * time.Second)
	for addr == "" && time.Now().Before(deadline) {
		addr = s.Address()
		if addr == "" {
			time.Sleep(10 * time.Millisecond)
		}
	}
	if addr == "" {
		t.Fatal("Address() is empty after Serve took a real listener")
	}
	if addr != ln.Addr().String() {
		t.Errorf("Address() = %q, want %q", addr, ln.Addr().String())
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(testCodecName)))
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	t.Cleanup(func() { conn.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reply, err := NewEchoClient(conn).Echo(ctx, &EchoRequest{Message: "over the wire"})
	if err != nil {
		t.Fatalf("Echo over real TCP: %v", err)
	}
	if reply.GetMessage() != "over the wire" {
		t.Errorf("reply = %q, want %q", reply.GetMessage(), "over the wire")
	}
}

// StartInBackground must report bind failures synchronously to the caller
// instead of swallowing them on the serve goroutine. Port -1 cannot be
// bound, so the error must come back from the call itself.
func TestServer_StartInBackgroundReportsBindError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Insecure = true
	cfg.Port = -1 // ": -1" is not a bindable port
	s := NewServer(cfg)

	err := s.StartInBackground()
	if err == nil {
		s.ImmediateStop()
		t.Fatal("StartInBackground must return the bind error")
	}
	if s.Address() != "" {
		t.Errorf("Address() = %q after a failed bind, want empty", s.Address())
	}
}
