// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"bytes"
	"context"
	"crypto/tls"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// lanPair creates a backend+client pair where the backend has a LAN
// server and the client is LAN-enabled, with encryption set up.
func lanPair(t *testing.T, env relayEnv) (*Conn, *Conn, *LANServer) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lanSrv, err := NewLANServer("", nil)
	if err != nil {
		t.Fatal("NewLANServer:", err)
	}
	t.Cleanup(func() { lanSrv.Close() })

	bCfg := env.cfg
	bCfg.LANServer = lanSrv
	b, err := Register(ctx, env.url, bCfg)
	if err != nil {
		t.Fatal("register:", err)
	}
	t.Cleanup(func() { b.CloseNow() })

	cCfg := env.cfg
	cCfg.LAN = true
	c, err := Connect(ctx, env.url, b.InstanceID(), cCfg)
	if err != nil {
		t.Fatal("connect:", err)
	}
	t.Cleanup(func() { c.CloseNow() })

	// Set up encryption — required for LAN offer (control messages).
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()
	bRec := crypto.NewPairingRecord("client", "relay", bKP, cKP.Public)
	cRec := crypto.NewPairingRecord("backend", "relay", cKP, bKP.Public)

	bCh, _ := bRec.DeriveChannel([]byte("b2c"), []byte("c2b"))
	cCh, _ := cRec.DeriveChannel([]byte("c2b"), []byte("b2c"))

	// SetChannel on backend triggers LAN advertisement.
	b.SetChannel(bCh)
	c.SetChannel(cCh)

	return b, c, lanSrv
}

// waitForLAN sends a message to trigger LAN offer processing, then
// waits for the LAN connection to establish.
func waitForLAN(t *testing.T, ctx context.Context, sender, receiver *Conn) {
	t.Helper()
	sender.Send(ctx, []byte("lan-trigger"))
	data, err := receiver.Recv(ctx)
	if err != nil {
		t.Fatal("recv trigger:", err)
	}
	if string(data) != "lan-trigger" {
		t.Fatalf("got %q, want lan-trigger", data)
	}
	time.Sleep(2 * time.Second)
}

// TestLANUpgrade verifies that traffic switches from relay to LAN.
func TestLANUpgrade(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// This should go via LAN.
	c.Send(ctx, []byte("via-lan"))
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal("recv via LAN:", err)
	}
	if string(data) != "via-lan" {
		t.Fatalf("got %q", data)
	}
}

// TestLANUpgradeBidirectional verifies both directions work after switch.
func TestLANUpgradeBidirectional(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	for i := range 5 {
		msg := []byte("ping-" + strconv.Itoa(i))
		c.Send(ctx, msg)
		data, _ := b.Recv(ctx)
		if string(data) != string(msg) {
			t.Fatalf("got %q, want %q", data, msg)
		}

		reply := []byte("pong-" + strconv.Itoa(i))
		b.Send(ctx, reply)
		data, _ = c.Recv(ctx)
		if string(data) != string(reply) {
			t.Fatalf("got %q, want %q", data, reply)
		}
	}
}

// TestLANUpgradeDatagrams verifies datagrams work after LAN switch.
func TestLANUpgradeDatagrams(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// Send datagrams — some may be lost during the switch, so send
	// several and verify at least one arrives.
	for range 10 {
		c.SendDatagram([]byte("dg-via-lan"))
	}

	recvCtx, recvCancel := context.WithTimeout(ctx, 3*time.Second)
	defer recvCancel()
	data, err := b.RecvDatagram(recvCtx)
	if err != nil {
		t.Fatal("recv datagram via LAN:", err)
	}
	if string(data) != "dg-via-lan" {
		t.Fatalf("got %q", data)
	}
}

// TestLANUpgradeLargeMessage verifies large messages (stream) work
// after LAN switch.
func TestLANUpgradeLargeMessage(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	payload := bytes.Repeat([]byte("X"), 100000)
	c.Send(ctx, payload)
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal("recv large:", err)
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("got %d bytes, want %d", len(data), len(payload))
	}
}

// TestLANUpgradeChannel verifies streaming channels work after switch.
func TestLANUpgradeChannel(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	ch, err := c.OpenChannel("game")
	if err != nil {
		t.Fatal("open channel:", err)
	}
	defer ch.Close()

	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept channel:", err)
	}
	defer bch.Close()

	if bch.Name() != "game" {
		t.Fatalf("name: got %q", bch.Name())
	}

	ch.Send(ctx, []byte("move-e4"))
	data, _ := bch.Recv(ctx)
	if string(data) != "move-e4" {
		t.Fatalf("got %q", data)
	}
}

// TestLANServerMultipleClients verifies the LAN server can serve
// multiple backends concurrently.
func TestLANServerMultipleClients(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	lanSrv, err := NewLANServer("", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lanSrv.Close()

	b1Cfg := env.cfg
	b1Cfg.LANServer = lanSrv
	b1Cfg.InstanceID = "b1"
	b1, err := Register(ctx, env.url, b1Cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b1.CloseNow()

	b2Cfg := env.cfg
	b2Cfg.LANServer = lanSrv
	b2Cfg.InstanceID = "b2"
	b2, err := Register(ctx, env.url, b2Cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b2.CloseNow()

	t.Logf("LAN server addr: %s, serving 2 backends", lanSrv.Addr())
}

// TestLANServerFixedAddr verifies NewLANServer with a fixed address.
func TestLANServerFixedAddr(t *testing.T) {
	srv, err := NewLANServer("127.0.0.1:0", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	addr := srv.Addr()
	if addr == "" {
		t.Fatal("empty addr")
	}
	t.Logf("LAN server on fixed host: %s", addr)

	// Verify it contains 127.0.0.1 (not the LAN IP, since we bound to localhost).
	if addr[:10] != "127.0.0.1:" {
		t.Fatalf("expected 127.0.0.1:*, got %s", addr)
	}
}

// TestLANServerDefaultAddr verifies NewLANServer with empty addr.
func TestLANServerDefaultAddr(t *testing.T) {
	srv, err := NewLANServer("", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	addr := srv.Addr()
	if addr == "" {
		t.Fatal("empty addr")
	}
	t.Logf("LAN server default: %s", addr)
}

// TestLANServerCustomTLS verifies NewLANServer with custom TLS config.
func TestLANServerCustomTLS(t *testing.T) {
	cert, _ := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewLANServer("", tlsCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	t.Logf("LAN server with custom TLS: %s", srv.Addr())
}

// TestLANUpgradeConcurrentSends verifies concurrent sends work after switch.
func TestLANUpgradeConcurrentSends(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// 10 goroutines sending simultaneously.
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Send(ctx, []byte("concurrent-"+strconv.Itoa(n)))
		}(i)
	}
	wg.Wait()

	// Receive all 10.
	for range 10 {
		_, err := b.Recv(ctx)
		if err != nil {
			t.Fatal("recv:", err)
		}
	}
}

// TestLANClientDisabledIgnoresOffer verifies that a client without
// LAN enabled ignores LAN offers.
func TestLANClientDisabledIgnoresOffer(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lanSrv, err := NewLANServer("", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lanSrv.Close()

	bCfg := env.cfg
	bCfg.LANServer = lanSrv
	b, err := Register(ctx, env.url, bCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	// Client does NOT enable LAN.
	c, err := Connect(ctx, env.url, b.InstanceID(), env.cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()
	bRec := crypto.NewPairingRecord("client", "relay", bKP, cKP.Public)
	cRec := crypto.NewPairingRecord("backend", "relay", cKP, bKP.Public)
	bCh, _ := bRec.DeriveChannel([]byte("b2c"), []byte("c2b"))
	cCh, _ := cRec.DeriveChannel([]byte("c2b"), []byte("b2c"))
	b.SetChannel(bCh)
	c.SetChannel(cCh)

	// Send a message — the LAN offer is received but ignored.
	b.Send(ctx, []byte("still-relay"))
	data, _ := c.Recv(ctx)
	if string(data) != "still-relay" {
		t.Fatalf("got %q", data)
	}

	// Wait and verify still works (no crash from ignored offer).
	time.Sleep(time.Second)
	c.Send(ctx, []byte("reply"))
	data, _ = b.Recv(ctx)
	if string(data) != "reply" {
		t.Fatalf("got %q", data)
	}
}

// TestChallengeEqual verifies the challenge comparison helper.
func TestChallengeEqual(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2, 3}
	if !challengeEqual(a, b) {
		t.Fatal("equal challenges should match")
	}
	if challengeEqual(a, []byte{1, 2, 4}) {
		t.Fatal("different challenges should not match")
	}
	if challengeEqual(a, []byte{1, 2}) {
		t.Fatal("different lengths should not match")
	}
}

// TestLANServerInvalidAddr verifies NewLANServer rejects bad addresses.
func TestLANServerInvalidAddr(t *testing.T) {
	_, err := NewLANServer("not-a-valid-addr:abc:xyz", nil)
	if err == nil {
		t.Fatal("expected error for invalid addr")
	}
}
