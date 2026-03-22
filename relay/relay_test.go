// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package relay_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/marcelocantos/tern/relay"
)

// startRelay starts a tern relay server for testing. We import the
// relay package under test, so we can't access the main package's
// registerRoutes directly. Instead, start the actual binary or use
// a minimal relay implementation.
//
// For now, we use a minimal inline relay that mirrors the real one.
func startRelay(t *testing.T, token string) string {
	t.Helper()

	type instance struct {
		conn *websocket.Conn
		ctx  context.Context
	}
	var inst *instance

	mux := http.NewServeMux()
	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		if token != "" {
			if r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
		}
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		conn.Write(r.Context(), websocket.MessageText, []byte("test-id"))
		inst = &instance{conn: conn, ctx: r.Context()}
		<-r.Context().Done()
	})
	mux.HandleFunc("GET /ws/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientConn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer clientConn.CloseNow()

		if inst == nil {
			return
		}

		// Bridge: backend → client.
		go func() {
			for {
				mt, data, err := inst.conn.Read(inst.ctx)
				if err != nil {
					return
				}
				clientConn.Write(r.Context(), mt, data)
			}
		}()

		// Bridge: client → backend.
		for {
			mt, data, err := clientConn.Read(r.Context())
			if err != nil {
				return
			}
			inst.conn.Write(inst.ctx, mt, data)
		}
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return "ws" + strings.TrimPrefix(ts.URL, "http")
}

func TestRegisterAndConnect(t *testing.T) {
	wsBase := startRelay(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Backend registers.
	backend, err := relay.Register(ctx, wsBase)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	defer backend.CloseNow()

	if backend.InstanceID() != "test-id" {
		t.Fatalf("InstanceID = %q, want test-id", backend.InstanceID())
	}

	// Client connects.
	client, err := relay.Connect(ctx, wsBase, backend.InstanceID())
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.CloseNow()

	// Client → backend.
	if err := client.Send(ctx, websocket.MessageText, []byte("hello")); err != nil {
		t.Fatalf("client Send: %v", err)
	}

	mt, data, err := backend.Recv(ctx)
	if err != nil {
		t.Fatalf("backend Recv: %v", err)
	}
	if mt != websocket.MessageText || string(data) != "hello" {
		t.Fatalf("backend got (%v, %q), want (text, hello)", mt, data)
	}

	// Backend → client.
	if err := backend.Send(ctx, websocket.MessageText, []byte("world")); err != nil {
		t.Fatalf("backend Send: %v", err)
	}

	mt, data, err = client.Recv(ctx)
	if err != nil {
		t.Fatalf("client Recv: %v", err)
	}
	if mt != websocket.MessageText || string(data) != "world" {
		t.Fatalf("client got (%v, %q), want (text, world)", mt, data)
	}
}

func TestRegisterWithToken(t *testing.T) {
	wsBase := startRelay(t, "secret")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Without token → fail.
	_, err := relay.Register(ctx, wsBase)
	if err == nil {
		t.Fatal("Register without token should fail")
	}

	// With wrong token → fail.
	_, err = relay.Register(ctx, wsBase, relay.WithToken("wrong"))
	if err == nil {
		t.Fatal("Register with wrong token should fail")
	}

	// With correct token → success.
	backend, err := relay.Register(ctx, wsBase, relay.WithToken("secret"))
	if err != nil {
		t.Fatalf("Register with correct token: %v", err)
	}
	defer backend.CloseNow()

	if backend.InstanceID() != "test-id" {
		t.Fatalf("InstanceID = %q, want test-id", backend.InstanceID())
	}
}
