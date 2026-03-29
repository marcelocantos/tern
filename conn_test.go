// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io"
	"math/big"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
	"github.com/quic-go/quic-go"
)

// generateTestCert creates a self-signed TLS certificate for testing.
func generateTestCert(t testing.TB) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	pool := x509.NewCertPool()
	parsedCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatal(err)
	}
	pool.AddCert(parsedCert)

	return cert, pool
}

// relayEnv holds the URL and options needed to connect to a relay.
type relayEnv struct {
	url  string
	opts []Option
}

// localRelay starts a test relay (WebTransport + raw QUIC) and returns
// a relayEnv configured for raw QUIC connections.
func localRelay(t *testing.T) relayEnv {
	t.Helper()
	// If an external instrumented relay is running (e.g. for coverage),
	// use it instead of starting our own.
	if qURL := os.Getenv("TERN_TEST_QUIC_URL"); qURL != "" {
		u, _ := url.Parse(qURL)
		return relayEnv{
			url: qURL,
			opts: []Option{
				WithTLS(&tls.Config{InsecureSkipVerify: true}),
				WithQUICPort(u.Port()),
			},
		}
	}
	return localRelayTB(t)
}

// localRelayTB is the shared implementation for tests and benchmarks.
func localRelayTB(t testing.TB) relayEnv {
	t.Helper()

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}

	// Start WebTransport server.
	wtUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(wtUDP)
	t.Cleanup(func() { srv.Close() })

	// Start raw QUIC server sharing the same hub.
	qsrv := NewQUICServer("127.0.0.1:0", tlsCfg, "", srv.Hub())
	qUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go qsrv.ServeWithTLS(qUDP, tlsCfg)
	t.Cleanup(func() { qsrv.Close() })

	wtPort := wtUDP.LocalAddr().(*net.UDPAddr).Port
	qPort := qUDP.LocalAddr().(*net.UDPAddr).Port

	// Default: raw QUIC. The URL host is used by both WT and QUIC paths.
	return relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(wtPort),
		opts: []Option{
			WithTLS(&tls.Config{RootCAs: pool}),
			WithQUICPort(strconv.Itoa(qPort)),
		},
	}
}

// localRelayWT starts a test relay and returns a relayEnv configured
// for WebTransport connections (backward compat / browser path).
func localRelayWT(t *testing.T) relayEnv {
	t.Helper()

	// If an external instrumented relay is running, use it.
	if wtURL := os.Getenv("TERN_TEST_WT_URL"); wtURL != "" {
		return relayEnv{
			url: wtURL,
			opts: []Option{
				WithTLS(&tls.Config{InsecureSkipVerify: true}),
				WithWebTransport(),
			},
		}
	}

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatal(err)
	}

	go srv.Serve(conn)
	t.Cleanup(func() { srv.Close() })

	addr := conn.LocalAddr().(*net.UDPAddr)
	return relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(addr.Port),
		opts: []Option{
			WithTLS(&tls.Config{RootCAs: pool}),
			WithWebTransport(),
		},
	}
}

// liveRelay returns a relayEnv for tern.fly.dev if TERN_TOKEN is set.
// Skips the test otherwise. Uses WebTransport since the live relay may
// not yet have a raw QUIC port.
func liveRelay(t *testing.T) relayEnv {
	t.Helper()
	token := os.Getenv("TERN_TOKEN")
	if token == "" {
		t.Skip("TERN_TOKEN not set; skipping live test")
	}

	env := relayEnv{
		url: "https://tern.fly.dev:443",
		opts: []Option{
			WithToken(token),
			WithWebTransport(),
		},
	}

	// Probe connectivity — skip if the relay isn't reachable over QUIC/UDP.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	probe, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Skipf("live relay not reachable: %v", err)
	}
	probe.CloseNow()

	return env
}

// connectPair registers a backend and connects a client, returning both.
func connectPair(t *testing.T, env relayEnv) (*Conn, *Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal("register:", err)
	}
	t.Cleanup(func() { b.CloseNow() })

	c, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}
	t.Cleanup(func() { c.CloseNow() })

	return b, c
}

// setupEncryption creates matching E2E channels on both sides.
func setupEncryption(t *testing.T, b, c *Conn) {
	t.Helper()
	bKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	bSendKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b-to-c"))
	if err != nil {
		t.Fatal(err)
	}
	bRecvKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cSendKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cRecvKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b-to-c"))
	if err != nil {
		t.Fatal(err)
	}

	bCh, err := crypto.NewChannel(bSendKey, bRecvKey)
	if err != nil {
		t.Fatal(err)
	}
	cCh, err := crypto.NewChannel(cSendKey, cRecvKey)
	if err != nil {
		t.Fatal(err)
	}

	b.SetChannel(bCh)
	c.SetChannel(cCh)
}

// setupDatagramEncryption creates matching datagram channels on both sides.
func setupDatagramEncryption(t *testing.T, b, c *Conn) {
	t.Helper()
	bKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	bSendKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("dg-b-to-c"))
	if err != nil {
		t.Fatal(err)
	}
	bRecvKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("dg-c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cSendKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("dg-c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cRecvKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("dg-b-to-c"))
	if err != nil {
		t.Fatal(err)
	}

	bCh, err := crypto.NewChannel(bSendKey, bRecvKey)
	if err != nil {
		t.Fatal(err)
	}
	cCh, err := crypto.NewChannel(cSendKey, cRecvKey)
	if err != nil {
		t.Fatal(err)
	}

	b.SetDatagramChannel(bCh)
	c.SetDatagramChannel(cCh)
}

// liveRelayEnv returns the token and URL for the live relay, or empty
// strings if TERN_TOKEN is not set.
func liveRelayEnv() (token, url string) {
	token = os.Getenv("TERN_TOKEN")
	if token == "" {
		return "", ""
	}
	return token, "https://tern.fly.dev:443"
}

// localRelayB is localRelay for benchmarks.
func localRelayB(b *testing.B) relayEnv {
	b.Helper()
	return localRelayTB(b)
}

// forEachRelay runs a subtest against local (QUIC), local (WebTransport),
// and live relay environments.
func forEachRelay(t *testing.T, fn func(t *testing.T, env relayEnv)) {
	t.Run("local/quic", func(t *testing.T) { fn(t, localRelay(t)) })
	t.Run("local/webtransport", func(t *testing.T) { fn(t, localRelayWT(t)) })
	t.Run("live", func(t *testing.T) { fn(t, liveRelay(t)) })
}

// --- Tests ---

func TestStreamRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)

		if err := c.Send(ctx, []byte("hello")); err != nil {
			t.Fatal(err)
		}
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "hello" {
			t.Fatalf("got %q, want hello", data)
		}

		if err := b.Send(ctx, []byte("reply")); err != nil {
			t.Fatal(err)
		}
		data, err = c.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "reply" {
			t.Fatalf("got %q, want reply", data)
		}
	})
}

func TestEncryptedStreamRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)
		setupEncryption(t, b, c)

		if err := c.Send(ctx, []byte("secret")); err != nil {
			t.Fatal(err)
		}
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "secret" {
			t.Fatalf("got %q, want secret", data)
		}
	})
}

func TestDatagramRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)

		if err := c.SendDatagram([]byte("dgram")); err != nil {
			t.Fatal(err)
		}
		data, err := b.RecvDatagram(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "dgram" {
			t.Fatalf("got %q, want dgram", data)
		}
	})
}

func TestEncryptedDatagramRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)
		setupDatagramEncryption(t, b, c)

		if err := c.SendDatagram([]byte("encrypted-dgram")); err != nil {
			t.Fatal(err)
		}
		data, err := b.RecvDatagram(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "encrypted-dgram" {
			t.Fatalf("got %q, want encrypted-dgram", data)
		}
	})
}

// TestOpenStream verifies that OpenStream succeeds on an established Conn.
// The relay currently only bridges the primary stream; additional streams
// are not forwarded to the peer. This test confirms the stream opens without
// error. End-to-end multi-stream relay requires server-side support —
// see the TODO in session.go (bridgeClient).
func TestOpenStream(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		b, _ := connectPair(t, env)

		stream, err := b.OpenStream()
		if err != nil {
			t.Fatalf("OpenStream: %v", err)
		}
		defer stream.Close()
	})
}

func TestMultipleMessages(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)

		const n = 10
		for i := range n {
			if err := c.Send(ctx, []byte("msg-"+strconv.Itoa(i))); err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		for i := range n {
			expected := "msg-" + strconv.Itoa(i)
			data, err := b.Recv(ctx)
			if err != nil {
				t.Fatalf("recv %d: %v", i, err)
			}
			if string(data) != expected {
				t.Fatalf("msg %d: got %q, want %q", i, data, expected)
			}
		}
	})
}

func TestPersistentInstanceID(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	myUUID := "test-device-uuid-12345"

	// Register with a persistent instance ID.
	b, err := Register(ctx, env.url, append(env.opts, WithInstanceID(myUUID))...)
	if err != nil {
		t.Fatal("register:", err)
	}
	if b.InstanceID() != myUUID {
		t.Fatalf("got instance ID %q, want %q", b.InstanceID(), myUUID)
	}

	// Client connects using the persistent ID.
	c, err := Connect(ctx, env.url, myUUID, env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}

	// Verify messaging works.
	if err := c.Send(ctx, []byte("persistent")); err != nil {
		t.Fatal(err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "persistent" {
		t.Fatalf("got %q, want persistent", data)
	}

	c.CloseNow()
	b.CloseNow()
}

func TestStreamingChannel(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer c.CloseNow()

	// Client opens a named channel.
	ch, err := c.OpenChannel("game-state")
	if err != nil {
		t.Fatal("open channel:", err)
	}
	defer ch.Close()

	// Backend accepts the channel.
	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept channel:", err)
	}
	defer bch.Close()

	if bch.Name() != "game-state" {
		t.Fatalf("got channel name %q, want game-state", bch.Name())
	}

	// Send/recv on the channel.
	if err := ch.Send(ctx, []byte("player moved")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := bch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "player moved" {
		t.Fatalf("got %q, want 'player moved'", data)
	}

	// Reverse direction.
	if err := bch.Send(ctx, []byte("state updated")); err != nil {
		t.Fatal("send:", err)
	}
	data, err = ch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "state updated" {
		t.Fatalf("got %q, want 'state updated'", data)
	}
}

// TestPersistentInstanceIDWebTransport tests WithInstanceID via the
// WebTransport path (registerWebTransport).
func TestPersistentInstanceIDWebTransport(t *testing.T) {
	env := localRelayWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	myUUID := "wt-persistent-uuid-99"

	b, err := Register(ctx, env.url, append(env.opts, WithInstanceID(myUUID))...)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	if b.InstanceID() != myUUID {
		t.Fatalf("got instance ID %q, want %q", b.InstanceID(), myUUID)
	}

	c, err := Connect(ctx, env.url, myUUID, env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer c.CloseNow()

	if err := c.Send(ctx, []byte("wt-persistent")); err != nil {
		t.Fatal(err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "wt-persistent" {
		t.Fatalf("got %q, want wt-persistent", data)
	}
}

// TestConnectToNonListeningURL verifies that Register/Connect to a
// URL that isn't listening returns a clean error (not a hang or panic).
func TestConnectToNonListeningURL(t *testing.T) {
	// Use a port that's almost certainly not listening.
	badURL := "https://127.0.0.1:19999"
	tlsOpt := WithTLS(&tls.Config{InsecureSkipVerify: true})

	t.Run("register/quic", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := Register(ctx, badURL, tlsOpt, WithQUICPort("19999"))
		if err == nil {
			t.Fatal("expected error registering to non-listening URL")
		}
		t.Logf("register quic error: %v", err)
	})

	t.Run("connect/quic", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := Connect(ctx, badURL, "fake-id", tlsOpt, WithQUICPort("19999"))
		if err == nil {
			t.Fatal("expected error connecting to non-listening URL")
		}
		t.Logf("connect quic error: %v", err)
	})

	t.Run("register/wt", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := Register(ctx, badURL, tlsOpt, WithWebTransport())
		if err == nil {
			t.Fatal("expected error registering to non-listening URL via WT")
		}
		t.Logf("register WT error: %v", err)
	})

	t.Run("connect/wt", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := Connect(ctx, badURL, "fake-id", tlsOpt, WithWebTransport())
		if err == nil {
			t.Fatal("expected error connecting to non-listening URL via WT")
		}
		t.Logf("connect WT error: %v", err)
	})
}

// TestAcceptStreamNilAcceptor verifies that AcceptChannel on a Conn
// with no acceptor (nil streamAcceptor) returns an error.
func TestAcceptStreamNilAcceptor(t *testing.T) {
	// Create a minimal Conn with nil acceptor.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c := newConn(nil, nil, io.NopCloser(nil), nil, nil, "test")
	_, err := c.AcceptChannel(ctx)
	if err == nil {
		t.Fatal("expected error from AcceptChannel with nil acceptor")
	}
	t.Logf("nil acceptor error: %v", err)
}

// TestQUICServerListenAndServeAddr tests that ListenAndServe starts the
// server and Addr returns a non-nil address.
func TestQUICServerListenAndServeAddr(t *testing.T) {
	cert, _ := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	h := newHub()
	srv := NewQUICServer("127.0.0.1:0", tlsCfg, "", h)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe(tlsCfg) }()
	// Give the server a moment to start.
	time.Sleep(50 * time.Millisecond)

	addr := srv.Addr()
	if addr == nil {
		t.Fatal("Addr() returned nil after ListenAndServe")
	}
	t.Logf("QUIC server listening on %s", addr)

	if err := srv.Close(); err != nil {
		t.Fatal("close:", err)
	}
}

// TestWebTransportServerListenAndServeAddr tests ListenAndServe/Addr
// for the WebTransport server.
func TestWebTransportServerListenAndServeAddr(t *testing.T) {
	cert, _ := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	time.Sleep(50 * time.Millisecond)

	addr := srv.Addr()
	if addr == nil {
		t.Fatal("Addr() returned nil after ListenAndServe")
	}
	t.Logf("WT server listening on %s", addr)

	if err := srv.Close(); err != nil {
		t.Fatal("close:", err)
	}
}

// TestRegisterWithTokenWebTransport tests token validation via the
// WebTransport path, including the token rejection case.
func TestRegisterWithTokenWebTransport(t *testing.T) {
	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "wt-secret")
	if err != nil {
		t.Fatal(err)
	}

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(udpConn)
	t.Cleanup(func() { srv.Close() })

	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	relayURL := "https://127.0.0.1:" + strconv.Itoa(port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wrong token should fail.
	_, err = Register(ctx, relayURL,
		WithTLS(&tls.Config{RootCAs: pool}),
		WithWebTransport(),
		WithToken("wrong-token"),
	)
	if err == nil {
		t.Fatal("expected error with wrong WT token")
	}
	t.Logf("wrong WT token: %v", err)

	// Correct token should succeed.
	b, err := Register(ctx, relayURL,
		WithTLS(&tls.Config{RootCAs: pool}),
		WithWebTransport(),
		WithToken("wt-secret"),
	)
	if err != nil {
		t.Fatal("register with correct WT token:", err)
	}
	defer b.CloseNow()
}

// TestConnectToOccupiedInstanceWebTransport tests that a second client
// connecting to the same instance via WebTransport gets rejected.
func TestConnectToOccupiedInstanceWebTransport(t *testing.T) {
	env := localRelayWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	// First client connects.
	c1, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal("connect c1:", err)
	}
	defer c1.CloseNow()

	// Verify first client works.
	if err := c1.Send(ctx, []byte("wt-c1")); err != nil {
		t.Fatal("c1 send:", err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal("b recv:", err)
	}
	if string(data) != "wt-c1" {
		t.Fatalf("got %q, want wt-c1", data)
	}

	// Second client should be rejected.
	c2, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Logf("second WT client rejected at connect: %v", err)
		return
	}
	defer c2.CloseNow()

	err = c2.Send(ctx, []byte("probe"))
	if err == nil {
		_, err = c2.Recv(ctx)
	}
	if err == nil {
		t.Fatal("expected error for second WT client")
	}
	t.Logf("second WT client rejected (deferred): %v", err)
}

// TestConnectToNonExistentInstanceWebTransport tests connecting to a
// non-existent instance via WebTransport (handleClient 404 path).
func TestConnectToNonExistentInstanceWebTransport(t *testing.T) {
	env := localRelayWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := Connect(ctx, env.url, "does-not-exist-wt-99", env.opts...)
	if err != nil {
		t.Logf("connect to non-existent WT: %v", err)
		return
	}
	t.Fatal("expected error connecting to non-existent WT instance")
}

// TestConnectToInvalidInstanceIDWebTransport tests that a very long
// or empty instance ID is rejected by the WT handleClient path.
func TestConnectToInvalidInstanceIDWebTransport(t *testing.T) {
	env := localRelayWT(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Empty ID — the /ws/ prefix is present but the ID is empty.
	// This may fail at Connect or at the first Send/Recv.
	_, err := Connect(ctx, env.url, "", env.opts...)
	if err != nil {
		t.Logf("empty WT instance ID: %v", err)
		return
	}
	t.Fatal("expected error with empty WT instance ID")
}

// TestQUICServerAddrBeforeStart verifies Addr returns nil before ListenAndServe.
func TestQUICServerAddrBeforeStart(t *testing.T) {
	h := newHub()
	srv := NewQUICServer("127.0.0.1:0", nil, "", h)
	if addr := srv.Addr(); addr != nil {
		t.Fatalf("expected nil Addr before start, got %v", addr)
	}
}

// TestWebTransportServerAddrBeforeStart verifies Addr returns nil before
// Serve/ListenAndServe.
func TestWebTransportServerAddrBeforeStart(t *testing.T) {
	cert, _ := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}
	if addr := srv.Addr(); addr != nil {
		t.Fatalf("expected nil Addr before start, got %v", addr)
	}
}

// TestQUICServerCloseBeforeStart verifies Close does not error when called
// before ListenAndServe (listener is nil).
func TestQUICServerCloseBeforeStart(t *testing.T) {
	h := newHub()
	srv := NewQUICServer("127.0.0.1:0", nil, "", h)
	if err := srv.Close(); err != nil {
		t.Fatal("close before start:", err)
	}
}

// TestRegisterWithTokenQUIC tests token validation via the QUIC path.
func TestRegisterWithTokenQUIC(t *testing.T) {
	env := localRelayWithToken(t, "quic-secret")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wrong token: Replace the token in opts.
	wrongOpts := []Option{env.opts[0], env.opts[1], WithToken("bad-token")}
	b, err := Register(ctx, env.url, wrongOpts...)
	if err != nil {
		// QUIC may fail at register if the server closes the conn fast enough.
		t.Logf("wrong QUIC token err: %v", err)
		return
	}
	// If Register returned without error, the server may close the stream
	// on the first Send/Recv because it rejected the auth asynchronously.
	defer b.CloseNow()
	sendErr := b.Send(ctx, []byte("probe"))
	if sendErr == nil {
		_, sendErr = b.Recv(ctx)
	}
	if sendErr == nil {
		t.Fatal("expected error with wrong QUIC token")
	}
	t.Logf("wrong QUIC token deferred err: %v", sendErr)
}

// TestConnectToNonExistentInstanceQUIC tests the QUIC handleConnect path
// for a missing instance.
func TestConnectToNonExistentInstanceQUIC(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := Connect(ctx, env.url, "quic-nonexistent-xyz", env.opts...)
	if err != nil {
		t.Logf("connect to non-existent QUIC: %v", err)
		return
	}
	defer c.CloseNow()

	// The QUIC server closes the connection after the handshake. The error
	// surfaces on the first Send or Recv.
	err = c.Send(ctx, []byte("probe"))
	if err == nil {
		_, err = c.Recv(ctx)
	}
	if err == nil {
		t.Fatal("expected error for non-existent QUIC instance")
	}
	t.Logf("non-existent QUIC instance deferred: %v", err)
}

// TestSecondClientRejectedQUIC tests the QUIC handleConnect "occupied" path.
func TestSecondClientRejectedQUIC(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	c1, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal("connect c1:", err)
	}
	defer c1.CloseNow()

	// Verify c1 works.
	if err := c1.Send(ctx, []byte("c1")); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Recv(ctx); err != nil {
		t.Fatal(err)
	}

	// Second client via QUIC.
	c2, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Logf("second QUIC client rejected at connect: %v", err)
		return
	}
	defer c2.CloseNow()

	err = c2.Send(ctx, []byte("probe"))
	if err == nil {
		_, err = c2.Recv(ctx)
	}
	if err == nil {
		t.Fatal("expected error for second QUIC client")
	}
}

// TestEncryptedRecvControlMessages tests that Recv in encrypted mode
// properly handles control message types (msgLANOffer, msgCutover) by
// discarding them, and the default branch for unknown types.
func TestEncryptedRecvControlMessages(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	setupEncryption(t, b, c)

	// We need to send raw encrypted frames with different message types.
	// Access the backend's channel directly to encrypt custom payloads.
	b.mu.Lock()
	bCh := b.channel
	b.mu.Unlock()

	// Encrypt a control message (msgLANOffer = 0x01).
	lanOfferPayload := append([]byte{0x01}, []byte("lan-offer-data")...)
	cipherLAN := bCh.Encrypt(lanOfferPayload)
	b.writeMu.Lock()
	writeMessage(b.stream, cipherLAN)
	b.writeMu.Unlock()

	// Encrypt a control message (msgCutover = 0x02).
	cutoverPayload := append([]byte{0x02}, []byte("cutover-data")...)
	cipherCutover := bCh.Encrypt(cutoverPayload)
	b.writeMu.Lock()
	writeMessage(b.stream, cipherCutover)
	b.writeMu.Unlock()

	// Encrypt an unknown message type (0xFF).
	unknownPayload := append([]byte{0xFF}, []byte("unknown-data")...)
	cipherUnknown := bCh.Encrypt(unknownPayload)
	b.writeMu.Lock()
	writeMessage(b.stream, cipherUnknown)
	b.writeMu.Unlock()

	// Now send a normal application message.
	if err := b.Send(ctx, []byte("after-control")); err != nil {
		t.Fatal("send app msg:", err)
	}

	// Client should skip the control messages and receive only the app message.
	data, err := c.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "after-control" {
		t.Fatalf("got %q, want after-control", data)
	}
}

// TestEncryptedRecvEmptyPlaintext tests the `len(plaintext) == 0` branch.
func TestEncryptedRecvEmptyPlaintext(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	setupEncryption(t, b, c)

	// Encrypt an empty plaintext (no message type byte).
	b.mu.Lock()
	bCh := b.channel
	b.mu.Unlock()

	cipherEmpty := bCh.Encrypt([]byte{})
	b.writeMu.Lock()
	writeMessage(b.stream, cipherEmpty)
	b.writeMu.Unlock()

	// Recv should return nil data and nil error for empty plaintext.
	data, err := c.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if data != nil {
		t.Fatalf("expected nil data for empty plaintext, got %q", data)
	}
}

// TestQUICUnknownHandshake sends a raw QUIC connection with an unrecognized
// handshake message, testing the default case in handleConnection.
func TestQUICUnknownHandshake(t *testing.T) {
	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	h := newHub()
	srv := NewQUICServer("127.0.0.1:0", tlsCfg, "", h)

	udp, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.ServeWithTLS(udp, tlsCfg)
	t.Cleanup(func() { srv.Close() })

	port := udp.LocalAddr().(*net.UDPAddr).Port
	addr := "127.0.0.1:" + strconv.Itoa(port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientTLS := &tls.Config{RootCAs: pool, NextProtos: []string{ternALPN}}
	conn, err := quic.DialAddr(ctx, addr, clientTLS, &quic.Config{EnableDatagrams: true})
	if err != nil {
		t.Fatal("dial:", err)
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStream()
	if err != nil {
		t.Fatal("open stream:", err)
	}

	// Send an unrecognized handshake.
	if err := writeMessage(stream, []byte("garbage-handshake")); err != nil {
		t.Fatal("write:", err)
	}

	// The server should close the connection.
	time.Sleep(100 * time.Millisecond)
	_, err = stream.Read(make([]byte, 1))
	if err == nil {
		t.Fatal("expected error after unknown handshake")
	}
	t.Logf("unknown handshake: %v", err)
}

// TestQUICInvalidInstanceIDConnect sends a connect with empty or very long
// instance ID via raw QUIC, testing the validation in handleConnect.
func TestQUICInvalidInstanceIDConnect(t *testing.T) {
	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	h := newHub()
	srv := NewQUICServer("127.0.0.1:0", tlsCfg, "", h)

	udp, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.ServeWithTLS(udp, tlsCfg)
	t.Cleanup(func() { srv.Close() })

	port := udp.LocalAddr().(*net.UDPAddr).Port
	addr := "127.0.0.1:" + strconv.Itoa(port)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect with a too-long instance ID (>64 chars).
	clientTLS := &tls.Config{RootCAs: pool, NextProtos: []string{ternALPN}}
	conn, err := quic.DialAddr(ctx, addr, clientTLS, &quic.Config{EnableDatagrams: true})
	if err != nil {
		t.Fatal("dial:", err)
	}

	stream, err := conn.OpenStream()
	if err != nil {
		t.Fatal("open stream:", err)
	}

	longID := strings.Repeat("x", 65)
	if err := writeMessage(stream, []byte("connect:"+longID)); err != nil {
		t.Fatal("write:", err)
	}

	// Server should close with error.
	time.Sleep(100 * time.Millisecond)
	_, err = stream.Read(make([]byte, 1))
	if err == nil {
		t.Fatal("expected error with invalid instance ID")
	}
	t.Logf("invalid instance ID: %v", err)
	conn.CloseWithError(0, "")
}

// TestQuicTLSConfigNil verifies the nil TLS config path in quicTLSConfig.
func TestQuicTLSConfigNil(t *testing.T) {
	cfg := quicTLSConfig(options{})
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.NextProtos) == 0 || cfg.NextProtos[0] != ternALPN {
		t.Fatalf("expected ALPN %q, got %v", ternALPN, cfg.NextProtos)
	}
}

// TestQuicTLSConfigWithExisting verifies that an existing TLS config is
// cloned (not mutated) and ALPN is set.
func TestQuicTLSConfigWithExisting(t *testing.T) {
	orig := &tls.Config{InsecureSkipVerify: true}
	cfg := quicTLSConfig(options{tlsConfig: orig})
	if !cfg.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify to be cloned")
	}
	if len(cfg.NextProtos) == 0 || cfg.NextProtos[0] != ternALPN {
		t.Fatalf("expected ALPN %q, got %v", ternALPN, cfg.NextProtos)
	}
	// Verify original was not mutated.
	if len(orig.NextProtos) != 0 {
		t.Fatal("original TLS config was mutated")
	}
}

// TestQuicAddrBadURL tests the quicAddr error path with an invalid URL.
func TestQuicAddrBadURL(t *testing.T) {
	_, err := quicAddr("://invalid", options{})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

// TestQuicAddrCustomPort tests quicAddr with a custom port override.
func TestQuicAddrCustomPort(t *testing.T) {
	addr, err := quicAddr("https://example.com:443", options{quicPort: "5555"})
	if err != nil {
		t.Fatal(err)
	}
	if addr != "example.com:5555" {
		t.Fatalf("got %q, want example.com:5555", addr)
	}
}

// TestQuicAddrDefaultPort tests quicAddr with no port override.
func TestQuicAddrDefaultPort(t *testing.T) {
	addr, err := quicAddr("https://example.com:443", options{})
	if err != nil {
		t.Fatal(err)
	}
	if addr != "example.com:4433" {
		t.Fatalf("got %q, want example.com:4433", addr)
	}
}

// TestRegisterBadURLQUIC tests that Register with an unparseable URL
// returns a clean error for the QUIC path.
func TestRegisterBadURLQUIC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Register(ctx, "://bad-url")
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

// TestConnectBadURLQUIC tests Connect with an unparseable URL.
func TestConnectBadURLQUIC(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := Connect(ctx, "://bad-url", "some-id")
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

// TestPersistentInstanceIDQUIC tests the WithInstanceID path on QUIC.
func TestPersistentInstanceIDQUIC(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	myUUID := "quic-persistent-uuid-42"
	b, err := Register(ctx, env.url, append(env.opts, WithInstanceID(myUUID))...)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	if b.InstanceID() != myUUID {
		t.Fatalf("got %q, want %q", b.InstanceID(), myUUID)
	}

	c, err := Connect(ctx, env.url, myUUID, env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer c.CloseNow()

	if err := c.Send(ctx, []byte("quic-persistent")); err != nil {
		t.Fatal(err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "quic-persistent" {
		t.Fatalf("got %q, want quic-persistent", data)
	}
}

func TestMultipleChannels(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Open two channels.
	ch1, _ := c.OpenChannel("control")
	ch2, _ := c.OpenChannel("data")
	defer ch1.Close()
	defer ch2.Close()

	bch1, _ := b.AcceptChannel(ctx)
	bch2, _ := b.AcceptChannel(ctx)
	defer bch1.Close()
	defer bch2.Close()

	// Messages on different channels are independent.
	ch1.Send(ctx, []byte("ctrl-msg"))
	ch2.Send(ctx, []byte("data-msg"))

	d1, _ := bch1.Recv(ctx)
	d2, _ := bch2.Recv(ctx)

	// Channel names tell us which is which (order may vary due to concurrency).
	msgs := map[string]string{bch1.Name(): string(d1), bch2.Name(): string(d2)}
	if msgs["control"] != "ctrl-msg" {
		t.Fatalf("control channel got %q", msgs["control"])
	}
	if msgs["data"] != "data-msg" {
		t.Fatalf("data channel got %q", msgs["data"])
	}
}
