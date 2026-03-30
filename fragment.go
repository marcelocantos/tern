// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrDatagramTooLarge is returned when a payload exceeds the maximum
// fragmentable size (65535 fragments × max chunk size).
var ErrDatagramTooLarge = errors.New("datagram too large to fragment")

// Fragment header layout (10 bytes):
//   [0:4]  message ID  (uint32, monotonic per sender)
//   [4:6]  fragment index (uint16, 0-based)
//   [6:8]  total fragments (uint16, >= 1)
//   [8:10] reserved/flags (uint16, 0 for now)
//   [10:]  payload
const fragmentHeaderSize = 10

// DefaultFragmentTimeout is how long the receiver waits for all
// fragments of a message before discarding the partial assembly.
const DefaultFragmentTimeout = 5 * time.Second

// DefaultMaxDatagramPayload is the maximum payload per QUIC datagram.
// QUIC typically supports ~1200 bytes; we use a conservative default.
// The usable payload per fragment is this minus fragmentHeaderSize.
const DefaultMaxDatagramPayload = 1200

// Fragmenter handles splitting large datagrams into fragments and
// reassembling them on the receive side. It wraps a datagrammer.
//
// Usage:
//
//	frag := NewFragmenter(conn.dg)
//	frag.Send(largePayload)          // splits into multiple datagrams
//	assembled, _ := frag.Recv(ctx)   // reassembles from fragments
type Fragmenter struct {
	dg          datagrammer
	maxPayload  int           // max bytes per datagram (including header)
	timeout     time.Duration // reassembly timeout
	nextMsgID   atomic.Uint32
	mu          sync.Mutex
	assemblies  map[uint32]*assembly
	assembled   chan []byte // completed messages ready for delivery
	done        chan struct{}
	recvStarted sync.Once
}

type assembly struct {
	fragments [][]byte // indexed by fragment index; nil = not yet received
	total     int
	received  int
	deadline  time.Time
}

// FragmenterOption configures a Fragmenter.
type FragmenterOption func(*Fragmenter)

// WithMaxPayload sets the maximum datagram payload size (including
// the 10-byte fragment header). Default is 1200.
func WithMaxPayload(n int) FragmenterOption {
	return func(f *Fragmenter) { f.maxPayload = n }
}

// WithTimeout sets the reassembly timeout for incomplete messages.
func WithTimeout(d time.Duration) FragmenterOption {
	return func(f *Fragmenter) { f.timeout = d }
}

// NewFragmenter creates a fragmenter wrapping the given datagrammer.
func NewFragmenter(dg datagrammer, opts ...FragmenterOption) *Fragmenter {
	f := &Fragmenter{
		dg:         dg,
		maxPayload: DefaultMaxDatagramPayload,
		timeout:    DefaultFragmentTimeout,
		assemblies: make(map[uint32]*assembly),
		assembled:  make(chan []byte, 64),
		done:       make(chan struct{}),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Send fragments a payload and sends each fragment as a separate
// datagram. If the payload fits in a single datagram, it is sent
// as a single fragment (total=1).
func (f *Fragmenter) Send(data []byte) error {
	maxChunk := f.maxPayload - fragmentHeaderSize
	if maxChunk <= 0 {
		maxChunk = 1
	}

	total := (len(data) + maxChunk - 1) / maxChunk
	if total == 0 {
		total = 1 // empty payload → single fragment
	}
	if total > 65535 {
		return ErrDatagramTooLarge
	}

	msgID := f.nextMsgID.Add(1)

	for i := 0; i < total; i++ {
		start := i * maxChunk
		end := start + maxChunk
		if end > len(data) {
			end = len(data)
		}
		chunk := data[start:end]

		frame := make([]byte, fragmentHeaderSize+len(chunk))
		binary.BigEndian.PutUint32(frame[0:4], msgID)
		binary.BigEndian.PutUint16(frame[4:6], uint16(i))
		binary.BigEndian.PutUint16(frame[6:8], uint16(total))
		// frame[8:10] reserved, left as zero
		copy(frame[fragmentHeaderSize:], chunk)

		if err := f.dg.SendDatagram(frame); err != nil {
			return err
		}
	}
	return nil
}

// Recv receives the next fully reassembled message. Blocks until a
// complete message is available or the context is cancelled.
func (f *Fragmenter) Recv(ctx context.Context) ([]byte, error) {
	f.recvStarted.Do(func() {
		go f.recvLoop()
		go f.cleanupLoop()
	})

	select {
	case msg := <-f.assembled:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.done:
		return nil, context.Canceled
	}
}

// Close stops the fragmenter's background goroutines.
func (f *Fragmenter) Close() {
	select {
	case <-f.done:
	default:
		close(f.done)
	}
}

// recvLoop reads raw datagrams, extracts the fragment header, and
// feeds fragments into the assembly map.
func (f *Fragmenter) recvLoop() {
	for {
		data, err := f.dg.ReceiveDatagram(context.Background())
		if err != nil {
			select {
			case <-f.done:
				return
			default:
				continue
			}
		}

		if len(data) < fragmentHeaderSize {
			continue // runt packet, discard
		}

		msgID := binary.BigEndian.Uint32(data[0:4])
		fragIdx := int(binary.BigEndian.Uint16(data[4:6]))
		totalFrags := int(binary.BigEndian.Uint16(data[6:8]))
		payload := data[fragmentHeaderSize:]

		if totalFrags == 0 || fragIdx >= totalFrags {
			continue // invalid header
		}

		f.mu.Lock()
		a, ok := f.assemblies[msgID]
		if !ok {
			a = &assembly{
				fragments: make([][]byte, totalFrags),
				total:     totalFrags,
				deadline:  time.Now().Add(f.timeout),
			}
			f.assemblies[msgID] = a
		}

		// Store fragment if not already received.
		if a.fragments[fragIdx] == nil {
			frag := make([]byte, len(payload))
			copy(frag, payload)
			a.fragments[fragIdx] = frag
			a.received++
		}

		complete := a.received == a.total
		var assembled []byte
		if complete {
			// Calculate total size and reassemble.
			size := 0
			for _, frag := range a.fragments {
				size += len(frag)
			}
			assembled = make([]byte, 0, size)
			for _, frag := range a.fragments {
				assembled = append(assembled, frag...)
			}
			delete(f.assemblies, msgID)
		}
		f.mu.Unlock()

		if complete {
			select {
			case f.assembled <- assembled:
			case <-f.done:
				return
			}
		}
	}
}

// cleanupLoop periodically removes expired incomplete assemblies.
func (f *Fragmenter) cleanupLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-f.done:
			return
		case now := <-ticker.C:
			f.mu.Lock()
			for id, a := range f.assemblies {
				if now.After(a.deadline) {
					delete(f.assemblies, id)
				}
			}
			f.mu.Unlock()
		}
	}
}
