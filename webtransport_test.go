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
	"math/big"
	"net"
	"strconv"
	"testing"
	"time"
)

// generateTestCert creates a self-signed TLS certificate for testing.
func generateTestCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
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

// startWTRelay starts a WebTransport relay server on an ephemeral port
// and returns the URL (https://...) for connecting.
func startWTRelay(t *testing.T) (string, *x509.CertPool) {
	t.Helper()

	cert, pool := generateTestCert(t)

	srv, err := NewWebTransportServer("127.0.0.1:0", cert, "")
	if err != nil {
		t.Fatal(err)
	}

	// Listen on a random UDP port.
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := srv.Serve(conn); err != nil && srv.Addr() != nil {
			// Server closed normally during cleanup.
		}
	}()
	t.Cleanup(func() { srv.Close() })

	addr := conn.LocalAddr().(*net.UDPAddr)
	url := "https://127.0.0.1:" + strconv.Itoa(addr.Port)
	return url, pool
}

func TestWebTransportRoundTrip(t *testing.T) {
	url, pool := startWTRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tlsConfig := &tls.Config{
		RootCAs: pool,
	}

	// Backend registers.
	backend, err := RegisterWT(ctx, url, WithWebTransport(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	if backend.InstanceID() == "" {
		t.Fatal("expected non-empty instance ID")
	}
	t.Log("instance ID:", backend.InstanceID())

	// Client connects.
	client, err := ConnectWT(ctx, url, backend.InstanceID(), WithWebTransport(tlsConfig))
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer client.CloseNow()

	// Client -> Backend
	if err := client.Send(ctx, 0, []byte("hello from client")); err != nil {
		t.Fatal("client send:", err)
	}

	_, data, err := backend.Recv(ctx)
	if err != nil {
		t.Fatal("backend recv:", err)
	}
	if string(data) != "hello from client" {
		t.Fatalf("got %q, want %q", data, "hello from client")
	}

	// Backend -> Client
	if err := backend.Send(ctx, 0, []byte("hello from backend")); err != nil {
		t.Fatal("backend send:", err)
	}

	_, data, err = client.Recv(ctx)
	if err != nil {
		t.Fatal("client recv:", err)
	}
	if string(data) != "hello from backend" {
		t.Fatalf("got %q, want %q", data, "hello from backend")
	}
}

func TestWebTransportDatagrams(t *testing.T) {
	url, pool := startWTRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tlsConfig := &tls.Config{
		RootCAs: pool,
	}

	backend, err := RegisterWT(ctx, url, WithWebTransport(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	client, err := ConnectWT(ctx, url, backend.InstanceID(), WithWebTransport(tlsConfig))
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer client.CloseNow()

	// Client -> Backend datagram
	if err := client.SendDatagram([]byte("dgram-c2b")); err != nil {
		t.Fatal("client send datagram:", err)
	}

	data, err := backend.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("backend recv datagram:", err)
	}
	if string(data) != "dgram-c2b" {
		t.Fatalf("got %q, want %q", data, "dgram-c2b")
	}

	// Backend -> Client datagram
	if err := backend.SendDatagram([]byte("dgram-b2c")); err != nil {
		t.Fatal("backend send datagram:", err)
	}

	data, err = client.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("client recv datagram:", err)
	}
	if string(data) != "dgram-b2c" {
		t.Fatalf("got %q, want %q", data, "dgram-b2c")
	}
}

func TestWebTransportMultipleMessages(t *testing.T) {
	url, pool := startWTRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tlsConfig := &tls.Config{
		RootCAs: pool,
	}

	backend, err := RegisterWT(ctx, url, WithWebTransport(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	client, err := ConnectWT(ctx, url, backend.InstanceID(), WithWebTransport(tlsConfig))
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer client.CloseNow()

	const n = 10
	// Send n messages client -> backend.
	for i := range n {
		msg := []byte("msg-" + strconv.Itoa(i))
		if err := client.Send(ctx, 0, msg); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	// Receive all n messages.
	for i := range n {
		expected := "msg-" + strconv.Itoa(i)
		_, data, err := backend.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		if string(data) != expected {
			t.Fatalf("msg %d: got %q, want %q", i, data, expected)
		}
	}
}

func TestWebTransportDatagramOnWebSocket(t *testing.T) {
	// Datagrams should fail gracefully on WebSocket connections.
	wsBase := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := Register(ctx, wsBase)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	if err := b.SendDatagram([]byte("test")); err == nil {
		t.Fatal("expected error sending datagram on WebSocket connection")
	}

	if _, err := b.RecvDatagram(ctx); err == nil {
		t.Fatal("expected error receiving datagram on WebSocket connection")
	}
}
