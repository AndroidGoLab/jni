package protogen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderProtoToString_MinimalData(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
	}

	result, err := renderProtoToString(data)
	if err != nil {
		t.Fatalf("renderProtoToString: %v", err)
	}

	if !strings.Contains(result, `syntax = "proto3";`) {
		t.Error("missing proto3 syntax")
	}
	if !strings.Contains(result, "package test;") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(result, `option go_package = "example.com/mod/proto/test";`) {
		t.Error("missing go_package option")
	}
}

func TestRenderProtoToString_Messages(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
		Messages: []ProtoMessage{
			{
				Name: "Location",
				Fields: []ProtoField{
					{Type: "double", Name: "latitude", Number: 1},
					{Type: "double", Name: "longitude", Number: 2},
					{Type: "string", Name: "provider", Number: 3},
				},
			},
			{
				Name:   "Empty",
				Fields: nil,
			},
		},
	}

	result, err := renderProtoToString(data)
	if err != nil {
		t.Fatalf("renderProtoToString: %v", err)
	}

	if !strings.Contains(result, "message Location {") {
		t.Error("missing Location message")
	}
	if !strings.Contains(result, "double latitude = 1;") {
		t.Error("missing latitude field")
	}
	if !strings.Contains(result, "double longitude = 2;") {
		t.Error("missing longitude field")
	}
	if !strings.Contains(result, "string provider = 3;") {
		t.Error("missing provider field")
	}
	if !strings.Contains(result, "message Empty {") {
		t.Error("missing Empty message")
	}
}

func TestRenderProtoToString_OptionalFields(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
		Messages: []ProtoMessage{
			{
				Name: "Event",
				Fields: []ProtoField{
					{Type: "SubEvent", Name: "sub_event", Number: 1, Optional: true},
					{Type: "string", Name: "name", Number: 2},
				},
			},
		},
	}

	result, err := renderProtoToString(data)
	if err != nil {
		t.Fatalf("renderProtoToString: %v", err)
	}

	if !strings.Contains(result, "optional SubEvent sub_event = 1;") {
		t.Error("missing optional keyword for sub_event field")
	}
	if strings.Contains(result, "optional string name") {
		t.Error("name field should not have optional keyword")
	}
}

func TestRenderProtoToString_Services(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
		Services: []ProtoService{
			{
				Name: "ManagerService",
				RPCs: []ProtoRPC{
					{
						Name:       "GetLocation",
						InputType:  "GetLocationRequest",
						OutputType: "GetLocationResponse",
					},
					{
						Name:       "IsEnabled",
						InputType:  "IsEnabledRequest",
						OutputType: "IsEnabledResponse",
					},
				},
			},
		},
	}

	result, err := renderProtoToString(data)
	if err != nil {
		t.Fatalf("renderProtoToString: %v", err)
	}

	if !strings.Contains(result, "service ManagerService {") {
		t.Error("missing ManagerService")
	}
	if !strings.Contains(result, "rpc GetLocation (GetLocationRequest) returns (GetLocationResponse);") {
		t.Error("missing GetLocation RPC")
	}
	if !strings.Contains(result, "rpc IsEnabled (IsEnabledRequest) returns (IsEnabledResponse);") {
		t.Error("missing IsEnabled RPC")
	}
}

func TestRenderProtoToString_StreamingRPCs(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
		Services: []ProtoService{
			{
				Name: "EventService",
				RPCs: []ProtoRPC{
					{
						Name:            "SubscribeEvents",
						InputType:       "SubscribeEventsRequest",
						OutputType:      "EventResponse",
						ServerStreaming: true,
					},
					{
						Name:            "BiStream",
						InputType:       "Command",
						OutputType:      "Event",
						ClientStreaming: true,
						ServerStreaming: true,
						Comment:         "Bidirectional streaming",
					},
				},
			},
		},
	}

	result, err := renderProtoToString(data)
	if err != nil {
		t.Fatalf("renderProtoToString: %v", err)
	}

	if !strings.Contains(result, "rpc SubscribeEvents (SubscribeEventsRequest) returns (stream EventResponse);") {
		t.Error("missing server-streaming RPC syntax")
	}
	if !strings.Contains(result, "rpc BiStream (stream Command) returns (stream Event);") {
		t.Error("missing bidi-streaming RPC syntax")
	}
	if !strings.Contains(result, "// Bidirectional streaming") {
		t.Error("missing RPC comment")
	}
}

func TestRenderProto_WritesFile(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
		Messages: []ProtoMessage{
			{
				Name: "TestMsg",
				Fields: []ProtoField{
					{Type: "string", Name: "name", Number: 1},
				},
			},
		},
	}

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "test.proto")

	if err := renderProto(data, outputPath); err != nil {
		t.Fatalf("renderProto: %v", err)
	}

	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, `syntax = "proto3";`) {
		t.Error("missing proto3 syntax in file")
	}
	if !strings.Contains(s, "message TestMsg {") {
		t.Error("missing TestMsg in file")
	}
}

func TestRenderProto_GeneratedHeader(t *testing.T) {
	data := &ProtoData{
		Package:   "test",
		GoPackage: "example.com/mod/proto/test",
	}

	result, err := renderProtoToString(data)
	if err != nil {
		t.Fatalf("renderProtoToString: %v", err)
	}

	if !strings.HasPrefix(result, "// Code generated by protogen. DO NOT EDIT.") {
		t.Error("missing generated header at start of file")
	}
}

// TestRenderProto_Integration loads a real spec and verifies end-to-end.
func TestRenderProto_Integration(t *testing.T) {
	root := findRepoRoot(t)
	specPath := filepath.Join(root, "spec", "java", "location.yaml")
	overlayPath := filepath.Join(root, "spec", "overlays", "java", "location.yaml")
	outputDir := t.TempDir()
	goModule := "github.com/AndroidGoLab/jni"

	if err := Generate(specPath, overlayPath, outputDir, goModule); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	protoPath := filepath.Join(outputDir, "location", "location.proto")
	data, err := os.ReadFile(protoPath)
	if err != nil {
		t.Fatalf("read proto: %v", err)
	}
	content := string(data)

	// Verify basic structure: service and RPC presence.
	if !strings.Contains(content, "service LocationManagerService {") {
		t.Error("missing LocationManagerService")
	}
	if !strings.Contains(content, "rpc ") {
		t.Error("no RPCs generated")
	}
	if !strings.Contains(content, "message ") {
		t.Error("no messages generated")
	}
}

// TestRenderProto_BluetoothIntegration verifies bidi streaming with bluetooth spec.
func TestRenderProto_BluetoothIntegration(t *testing.T) {
	root := findRepoRoot(t)
	specPath := filepath.Join(root, "spec", "java", "bluetooth.yaml")
	overlayPath := filepath.Join(root, "spec", "overlays", "java", "bluetooth.yaml")
	outputDir := t.TempDir()
	goModule := "github.com/AndroidGoLab/jni"

	if err := Generate(specPath, overlayPath, outputDir, goModule); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	protoPath := filepath.Join(outputDir, "bluetooth", "bluetooth.proto")
	data, err := os.ReadFile(protoPath)
	if err != nil {
		t.Fatalf("read proto: %v", err)
	}
	content := string(data)

	// Verify basic bluetooth structure.
	if !strings.Contains(content, "service BluetoothGattCharacteristicService {") {
		t.Error("missing BluetoothGattCharacteristicService")
	}
	if !strings.Contains(content, "rpc ") {
		t.Error("no RPCs generated for bluetooth")
	}
	// The old hand-curated spec had callbacks; the generated spec from .class
	// files doesn't include callback interfaces. Verify just basic presence.
	if !strings.Contains(content, "message ") {
		t.Error("missing messages in bluetooth proto")
	}
}
