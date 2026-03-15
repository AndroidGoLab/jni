// Package protoscan reads compiled proto Go stubs to resolve actual
// Go names for services, RPCs, and messages. Protoc-gen-go may rename
// identifiers (e.g., P2p → P2P, A2dp → A2Dp) relative to the proto
// source names. This package bridges that gap.
package protoscan

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	newClientRe = regexp.MustCompile(`^func New(\w+Client)\(`)
	rpcMethodRe = regexp.MustCompile(`FullMethodName\s*=\s*"/[^/]+/(\w+)"`)
	// Match Go interface method declarations: MethodName(ctx context.Context, ...
	goMethodRe = regexp.MustCompile(`^\t(\w+)\(ctx context\.Context,`)
	// Match Go message struct declarations: type FooRequest struct {
	msgTypeRe = regexp.MustCompile(`^type (\w+) struct \{`)
)

// GoNames holds the resolved Go identifiers for a proto package.
type GoNames struct {
	// ServiceClients maps lowercase service name → actual Go service name.
	ServiceClients map[string]string

	// RPCMethods maps lowercase RPC name → actual Go RPC name.
	RPCMethods map[string]string

	// MessageTypes maps lowercase message name → actual Go message type name.
	MessageTypes map[string]string
}

// Scan reads the *_grpc.pb.go files in a proto package directory and
// extracts the actual Go names for services and RPC methods.
func Scan(protoPackageDir string) GoNames {
	names := GoNames{
		ServiceClients: make(map[string]string),
		RPCMethods:     make(map[string]string),
		MessageTypes:   make(map[string]string),
	}

	// Scan both _grpc.pb.go (for service/RPC names) and .pb.go (for message names).
	matches, _ := filepath.Glob(filepath.Join(protoPackageDir, "*.pb.go"))
	for _, path := range matches {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()

			// Match client constructors: func NewXxxClient(
			if m := newClientRe.FindStringSubmatch(line); m != nil {
				goName := strings.TrimSuffix(m[1], "Client")
				names.ServiceClients[strings.ToLower(goName)] = goName
			}

			// Match RPC full method names (wire names): "/package.Service/Method"
			if m := rpcMethodRe.FindStringSubmatch(line); m != nil {
				names.RPCMethods[strings.ToLower(m[1])] = m[1]
			}

			// Match Go interface method declarations (actual Go names).
			if m := goMethodRe.FindStringSubmatch(line); m != nil {
				goMethod := m[1]
				names.RPCMethods[strings.ToLower(goMethod)] = goMethod
			}

			// Match Go message struct type declarations.
			if m := msgTypeRe.FindStringSubmatch(line); m != nil {
				goType := m[1]
				names.MessageTypes[strings.ToLower(goType)] = goType
			}
		}
		_ = f.Close()
	}

	return names
}

// ResolveService returns the actual Go service name, falling back to
// the input if no resolution is found.
func (n GoNames) ResolveService(protoServiceName string) string {
	if resolved, ok := n.ServiceClients[strings.ToLower(protoServiceName)]; ok {
		return resolved
	}
	return protoServiceName
}

// ResolveRPC returns the actual Go RPC method name, falling back to
// the input if no resolution is found.
func (n GoNames) ResolveRPC(rpcName string) string {
	if resolved, ok := n.RPCMethods[strings.ToLower(rpcName)]; ok {
		return resolved
	}
	return rpcName
}

// ResolveMessage returns the actual Go message type name (as generated
// by protoc-gen-go), falling back to the input if no resolution is found.
// Protoc-gen-go may rename identifiers (e.g., P2p -> P2P, A2dp -> A2Dp).
func (n GoNames) ResolveMessage(msgName string) string {
	if resolved, ok := n.MessageTypes[strings.ToLower(msgName)]; ok {
		return resolved
	}
	return msgName
}
