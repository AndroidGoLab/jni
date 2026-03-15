package protogen

import (
	"strings"

	"github.com/AndroidGoLab/jni/tools/pkg/javagen"
)

// StreamingPattern classifies how a callback should be exposed as gRPC streaming.
type StreamingPattern int

const (
	// ServerStreaming means the server sends a stream of events to the client.
	ServerStreaming StreamingPattern = iota + 1
	// BidiStreaming means both client and server send streams.
	BidiStreaming
)

// bidiInterfaces is the set of Java interfaces known to require bidirectional
// streaming because they involve both read and write operations.
var bidiInterfaces = map[string]bool{
	"android.bluetooth.BluetoothGattCallback":       true,
	"android.bluetooth.BluetoothGattServerCallback": true,
}

// DetectStreamingPattern determines whether a callback should be exposed as
// server-streaming or bidirectional-streaming in the gRPC API.
//
// Known bidi interfaces (e.g. BluetoothGattCallback) are always BidiStreaming.
// Any callback with 2+ methods containing "read"/"write" (case-insensitive) is
// also BidiStreaming. Everything else is ServerStreaming.
func DetectStreamingPattern(cb *javagen.MergedCallback) StreamingPattern {
	if bidiInterfaces[cb.JavaInterface] {
		return BidiStreaming
	}

	readWriteCount := 0
	for _, m := range cb.Methods {
		lower := strings.ToLower(m.JavaMethod)
		if strings.Contains(lower, "read") || strings.Contains(lower, "write") {
			readWriteCount++
		}
	}
	if readWriteCount >= 2 {
		return BidiStreaming
	}

	return ServerStreaming
}
