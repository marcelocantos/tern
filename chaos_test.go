// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/marcelocantos/tern/faultproxy"
)

// TestChaos runs a long-duration chaos test that continuously sends
// and verifies messages while randomly cycling through network fault
// scenarios. This validates that tern's reliability layer (QUIC
// retransmits, path switching, fragmentation) handles real-world
// conditions correctly.
//
// Run with: go test -run TestChaos -timeout 600s -v
// Default duration: 60s. Override with CHAOS_DURATION=5m.
func TestChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("chaos test skipped in short mode")
	}

	duration := 60 * time.Second
	if d, err := time.ParseDuration(getEnvOr("CHAOS_DURATION", "60s")); err == nil {
		duration = d
	}

	t.Logf("chaos test: duration=%v", duration)

	// Set up relay with fault proxy.
	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}
	wtUDP, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	go srv.Serve(wtUDP)
	t.Cleanup(func() { srv.Close() })

	qsrv := NewQUICServer("127.0.0.1:0", tlsCfg, "", srv.Hub())
	qUDP, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	go qsrv.ServeWithTLS(qUDP, tlsCfg)
	t.Cleanup(func() { qsrv.Close() })

	qPort := qUDP.LocalAddr().(*net.UDPAddr).Port
	wtPort := wtUDP.LocalAddr().(*net.UDPAddr).Port

	proxy, err := faultproxy.New("127.0.0.1:" + strconv.Itoa(qPort))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { proxy.Close() })
	proxyAddr, _ := net.ResolveUDPAddr("udp", proxy.Addr())

	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	defer cancel()

	// Connect backend and client through the fault proxy.
	b, err := Register(ctx, "https://127.0.0.1:"+strconv.Itoa(wtPort), Config{
		TLS:      &tls.Config{RootCAs: pool},
		QUICPort: strconv.Itoa(proxyAddr.Port),
	})
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, "https://127.0.0.1:"+strconv.Itoa(wtPort), b.InstanceID(), Config{
		TLS:      &tls.Config{RootCAs: pool},
		QUICPort: strconv.Itoa(proxyAddr.Port),
	})
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer c.CloseNow()

	// --- Fault scenarios ---
	type scenario struct {
		name string
		opts []faultproxy.Option
	}
	scenarios := []scenario{
		{"clean", nil},
		{"latency-50ms", []faultproxy.Option{faultproxy.WithLatency(50*time.Millisecond, 20*time.Millisecond)}},
		{"latency-200ms", []faultproxy.Option{faultproxy.WithLatency(200*time.Millisecond, 50*time.Millisecond)}},
		{"loss-5%", []faultproxy.Option{faultproxy.WithPacketLoss(0.05)}},
		{"loss-15%", []faultproxy.Option{faultproxy.WithPacketLoss(0.15)}},
		{"corrupt-2%", []faultproxy.Option{faultproxy.WithCorrupt(0.02)}},
		{"latency+loss", []faultproxy.Option{
			faultproxy.WithLatency(100*time.Millisecond, 30*time.Millisecond),
			faultproxy.WithPacketLoss(0.05),
		}},
		{"bandwidth-20KB", []faultproxy.Option{faultproxy.WithBandwidth(20000)}},
		{"blackhole-brief", []faultproxy.Option{faultproxy.WithPacketLoss(1.0)}},
	}

	// --- Statistics ---
	var stats struct {
		streamSent     atomic.Int64
		streamRecvd    atomic.Int64
		streamErrors   atomic.Int64
		dgSent         atomic.Int64
		dgRecvd        atomic.Int64
		scenarioSwitch atomic.Int64
	}

	deadline := time.Now().Add(duration)
	done := make(chan struct{})

	// --- Chaos controller: randomly switch scenarios ---
	go func() {
		for time.Now().Before(deadline) {
			s := scenarios[rand.IntN(len(scenarios))]
			stats.scenarioSwitch.Add(1)

			if s.name == "blackhole-brief" {
				// Brief blackhole: 500ms-2s, then clean.
				proxy.UpdateProfile(s.opts...)
				blackholeDur := 500*time.Millisecond + time.Duration(rand.Int64N(int64(1500*time.Millisecond)))
				select {
				case <-time.After(blackholeDur):
				case <-done:
					return
				}
				proxy.UpdateProfile() // clean
			} else {
				proxy.UpdateProfile(s.opts...)
			}

			// Hold this scenario for 2-8 seconds.
			hold := 2*time.Second + time.Duration(rand.Int64N(int64(6*time.Second)))
			select {
			case <-time.After(hold):
			case <-done:
				return
			}
		}
		close(done)
	}()

	// --- Stream sender: sequential numbered messages ---
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; time.Now().Before(deadline); i++ {
			msg := make([]byte, 8)
			binary.BigEndian.PutUint64(msg, uint64(i))

			sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Send(sendCtx, msg)
			sendCancel()

			if err != nil {
				stats.streamErrors.Add(1)
				// Connection may be temporarily broken. Back off and retry.
				time.Sleep(100 * time.Millisecond)
				continue
			}
			stats.streamSent.Add(1)

			// Pace: 10-50 msgs/sec.
			time.Sleep(time.Duration(20+rand.IntN(80)) * time.Millisecond)
		}
	}()

	// --- Stream receiver: verify ordering ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		var lastSeq int64 = -1
		outOfOrder := 0

		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
			data, err := b.Recv(recvCtx)
			recvCancel()

			if err != nil {
				stats.streamErrors.Add(1)
				continue
			}
			stats.streamRecvd.Add(1)

			if len(data) == 8 {
				seq := int64(binary.BigEndian.Uint64(data))
				if seq <= lastSeq {
					outOfOrder++
					// Stream messages should be strictly ordered by QUIC.
					// Log but don't fail — we're testing under chaos.
					slog.Debug("stream out-of-order",
						"got", seq, "last", lastSeq)
				}
				lastSeq = seq
			}
		}

		if outOfOrder > 0 {
			t.Errorf("stream: %d out-of-order messages (of %d received)",
				outOfOrder, stats.streamRecvd.Load())
		}
	}()

	// --- Datagram sender: fire-and-forget ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			payload := fmt.Appendf(nil, "dg-%d", stats.dgSent.Load())
			if err := c.SendDatagram(payload); err == nil {
				stats.dgSent.Add(1)
			}
			time.Sleep(time.Duration(50+rand.IntN(100)) * time.Millisecond)
		}
	}()

	// --- Datagram receiver ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
			_, err := b.RecvDatagram(recvCtx)
			recvCancel()
			if err == nil {
				stats.dgRecvd.Add(1)
			}
		}
	}()

	// --- Large message sender (tests fragmentation under chaos) ---
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			size := 2000 + rand.IntN(8000) // 2-10KB
			payload := make([]byte, size)
			crand.Read(payload)

			sendCtx, sendCancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.Send(sendCtx, payload)
			sendCancel()

			if err == nil {
				stats.streamSent.Add(1)
			} else {
				stats.streamErrors.Add(1)
			}
			time.Sleep(time.Duration(500+rand.IntN(2000)) * time.Millisecond)
		}
	}()

	// --- Wait ---
	wg.Wait()

	// Ensure proxy is clean for final stats.
	proxy.UpdateProfile()

	proxyStats := proxy.GetStats()

	t.Logf("=== Chaos test results (%v) ===", duration)
	t.Logf("Scenarios switched: %d", stats.scenarioSwitch.Load())
	t.Logf("Stream: sent=%d recv=%d errors=%d loss=%.1f%%",
		stats.streamSent.Load(), stats.streamRecvd.Load(), stats.streamErrors.Load(),
		100*float64(stats.streamSent.Load()-stats.streamRecvd.Load())/
			float64(max(stats.streamSent.Load(), 1)))
	t.Logf("Datagrams: sent=%d recv=%d loss=%.1f%%",
		stats.dgSent.Load(), stats.dgRecvd.Load(),
		100*float64(stats.dgSent.Load()-stats.dgRecvd.Load())/
			float64(max(stats.dgSent.Load(), 1)))
	t.Logf("Proxy: forwarded=%d dropped=%d corrupted=%d",
		proxyStats.PacketsForwarded.Load(),
		proxyStats.PacketsDropped.Load(),
		proxyStats.PacketsCorrupted.Load())

	// --- Assertions ---
	// Stream messages are reliable (QUIC retransmits). We allow some
	// loss during blackhole periods (send timeout) but most should arrive.
	streamRecv := stats.streamRecvd.Load()
	streamSent := stats.streamSent.Load()
	if streamSent > 0 && streamRecv == 0 {
		t.Fatal("no stream messages received at all")
	}
	streamLoss := float64(streamSent-streamRecv) / float64(max(streamSent, 1))
	if streamLoss > 0.5 {
		t.Errorf("stream loss too high: %.1f%% (%d/%d)",
			streamLoss*100, streamSent-streamRecv, streamSent)
	}

	// Datagrams are unreliable — just verify some arrived.
	if stats.dgSent.Load() > 10 && stats.dgRecvd.Load() == 0 {
		t.Error("no datagrams received at all")
	}
}

// TestChaosWithLAN runs chaos with LAN upgrade active, testing
// path switching under fault injection.
func TestChaosWithLAN(t *testing.T) {
	if testing.Short() {
		t.Skip("chaos test skipped in short mode")
	}

	duration := 30 * time.Second
	t.Logf("chaos+LAN test: duration=%v", duration)

	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), duration+15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	var sent, recvd atomic.Int64
	deadline := time.Now().Add(duration)

	// Periodically force fallback and re-establishment.
	go func() {
		for time.Now().Before(deadline) {
			time.Sleep(3*time.Second + time.Duration(rand.Int64N(int64(4*time.Second))))
			c.router.fallbackToRelay()
			slog.Info("chaos: forced fallback")

			// Wait a bit, then the onSwitch callback will re-advertise
			// (with backoff). Meanwhile traffic flows via relay.
			time.Sleep(2 * time.Second)
		}
	}()

	// Sender.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; time.Now().Before(deadline); i++ {
			msg := []byte(fmt.Sprintf("chaos-%d", i))
			sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Send(sendCtx, msg)
			sendCancel()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			sent.Add(1)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Receiver.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
			data, err := b.Recv(recvCtx)
			recvCancel()
			if err != nil {
				continue
			}
			if bytes.HasPrefix(data, []byte("chaos-")) {
				recvd.Add(1)
			}
		}
	}()

	wg.Wait()

	t.Logf("chaos+LAN: sent=%d recv=%d loss=%.1f%%",
		sent.Load(), recvd.Load(),
		100*float64(sent.Load()-recvd.Load())/float64(max(sent.Load(), 1)))

	if sent.Load() > 0 && recvd.Load() == 0 {
		t.Fatal("no messages received")
	}
}

func getEnvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
