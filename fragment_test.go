// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"bytes"
	"context"
	"crypto/rand"
	"strconv"
	"testing"
	"time"
)

// TestFragmentSmallPayload sends a payload that fits in one fragment.
func TestFragmentSmallPayload(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	sf := c.Fragmenter()
	defer sf.Close()
	rf := b.Fragmenter()
	defer rf.Close()

	msg := []byte("hello fragments")
	if err := sf.Send(msg); err != nil {
		t.Fatal("send:", err)
	}

	got, err := rf.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(got) != string(msg) {
		t.Fatalf("got %q, want %q", got, msg)
	}
}

// TestFragmentLargePayload sends a payload larger than one datagram.
func TestFragmentLargePayload(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Use small max payload to force many fragments.
	sf := c.Fragmenter(WithMaxPayload(50))
	defer sf.Close()
	rf := b.Fragmenter(WithMaxPayload(50))
	defer rf.Close()

	// 500 bytes → 500 / (50-10) = 13 fragments.
	payload := make([]byte, 500)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	if err := sf.Send(payload); err != nil {
		t.Fatal("send:", err)
	}

	got, err := rf.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch: got %d bytes, want %d", len(got), len(payload))
	}
}

// TestFragmentEmptyPayload sends an empty payload.
func TestFragmentEmptyPayload(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	sf := c.Fragmenter()
	defer sf.Close()
	rf := b.Fragmenter()
	defer rf.Close()

	if err := sf.Send([]byte{}); err != nil {
		t.Fatal("send:", err)
	}

	got, err := rf.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d bytes, want 0", len(got))
	}
}

// TestFragmentMultipleMessages sends several large messages and
// verifies they all arrive correctly and in order.
func TestFragmentMultipleMessages(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	sf := c.Fragmenter(WithMaxPayload(100))
	defer sf.Close()
	rf := b.Fragmenter(WithMaxPayload(100))
	defer rf.Close()

	msgs := make([][]byte, 5)
	for i := range msgs {
		msgs[i] = []byte("message-" + strconv.Itoa(i) + "-" + string(make([]byte, 200)))
	}

	for _, msg := range msgs {
		if err := sf.Send(msg); err != nil {
			t.Fatal("send:", err)
		}
	}

	for i, want := range msgs {
		got, err := rf.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("message %d: got %d bytes, want %d", i, len(got), len(want))
		}
	}
}

// TestFragmentTimeout verifies that incomplete assemblies are dropped
// after the timeout expires.
func TestFragmentTimeout(t *testing.T) {
	// Use a mock datagrammer that only delivers some fragments.
	ch := make(chan []byte, 100)
	mock := &mockDatagram{ch: ch}

	f := NewFragmenter(mock, WithMaxPayload(50), WithTimeout(500*time.Millisecond))
	defer f.Close()

	// Manually send fragments for a 3-fragment message, but omit fragment 1.
	// Message ID 1, total 3.
	for _, idx := range []int{0, 2} {
		frame := make([]byte, fragmentHeaderSize+10)
		putUint32BE(frame[0:4], 1)
		putUint16BE(frame[4:6], uint16(idx))
		putUint16BE(frame[6:8], 3)
		copy(frame[fragmentHeaderSize:], []byte("payload..."))
		ch <- frame
	}

	// Recv should timeout — fragment 1 never arrives.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := f.Recv(ctx)
	if err == nil {
		t.Fatal("expected timeout for incomplete assembly")
	}

	// Verify the assembly was cleaned up.
	time.Sleep(time.Second)
	f.mu.Lock()
	remaining := len(f.assemblies)
	f.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("expected 0 assemblies after timeout, got %d", remaining)
	}
}

// TestFragmentBidirectional sends large messages in both directions.
func TestFragmentBidirectional(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	cf := c.Fragmenter(WithMaxPayload(80))
	defer cf.Close()
	bf := b.Fragmenter(WithMaxPayload(80))
	defer bf.Close()

	// Client → backend.
	payload1 := bytes.Repeat([]byte("A"), 300)
	cf.Send(payload1)

	// Backend → client.
	payload2 := bytes.Repeat([]byte("B"), 400)
	bf.Send(payload2)

	got1, err := bf.Recv(ctx)
	if err != nil {
		t.Fatal("recv on backend:", err)
	}
	if !bytes.Equal(got1, payload1) {
		t.Fatalf("backend got %d bytes, want %d", len(got1), len(payload1))
	}

	got2, err := cf.Recv(ctx)
	if err != nil {
		t.Fatal("recv on client:", err)
	}
	if !bytes.Equal(got2, payload2) {
		t.Fatalf("client got %d bytes, want %d", len(got2), len(payload2))
	}
}

// TestFragmentUnderPacketLoss sends a large fragmented message through
// the fault proxy with packet loss. Since fragments are datagrams,
// some may be lost and the message may not arrive. Verify no crash.
func TestFragmentUnderPacketLoss(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	cf := c.Fragmenter(WithMaxPayload(80), WithTimeout(2*time.Second))
	defer cf.Close()
	bf := b.Fragmenter(WithMaxPayload(80), WithTimeout(2*time.Second))
	defer bf.Close()

	// Send 10 large messages. Some may arrive, some may not (datagrams
	// are unreliable). The key assertion: no panics, no hangs.
	for i := range 10 {
		payload := make([]byte, 200)
		rand.Read(payload)
		payload[0] = byte(i) // tag for identification
		cf.Send(payload)
	}

	// Try to receive — we may get fewer than 10.
	received := 0
	recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
	defer recvCancel()
	for {
		_, err := bf.Recv(recvCtx)
		if err != nil {
			break
		}
		received++
	}
	t.Logf("fragment under loss: sent=10, received=%d", received)
}

// TestFragmentTooLarge verifies that payloads exceeding max fragment
// count are rejected.
func TestFragmentTooLarge(t *testing.T) {
	mock := &mockDatagram{ch: make(chan []byte, 1)}
	// maxPayload=11 means 1 byte per chunk. 65536 bytes → 65536 fragments > 65535 max.
	f := NewFragmenter(mock, WithMaxPayload(fragmentHeaderSize+1))
	defer f.Close()

	payload := make([]byte, 65536)
	err := f.Send(payload)
	if err != ErrDatagramTooLarge {
		t.Fatalf("expected ErrDatagramTooLarge, got %v", err)
	}
}

// --- helpers ---

type mockDatagram struct {
	ch chan []byte
}

func (m *mockDatagram) SendDatagram(data []byte) error {
	m.ch <- data
	return nil
}

func (m *mockDatagram) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	select {
	case d := <-m.ch:
		return d, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func putUint32BE(b []byte, v uint32) { b[0] = byte(v >> 24); b[1] = byte(v >> 16); b[2] = byte(v >> 8); b[3] = byte(v) }
func putUint16BE(b []byte, v uint16) { b[0] = byte(v >> 8); b[1] = byte(v) }
