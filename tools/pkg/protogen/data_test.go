package protogen

import (
	"testing"

	"github.com/AndroidGoLab/jni/tools/pkg/javagen"
)

func TestBuildProtoData_DataClassMessages(t *testing.T) {
	merged := &javagen.MergedSpec{
		Package: "test",
		DataClasses: []javagen.MergedDataClass{
			{
				GoType: "Location",
				Fields: []javagen.MergedField{
					{GoName: "Latitude", GoType: "float64", CallSuffix: "Double"},
					{GoName: "Longitude", GoType: "float64", CallSuffix: "Double"},
					{GoName: "Accuracy", GoType: "float32", CallSuffix: "Float"},
					{GoName: "Time", GoType: "int64", CallSuffix: "Long"},
					{GoName: "Provider", GoType: "string", CallSuffix: "Object"},
					{GoName: "Speed", GoType: "float32", CallSuffix: "Float"},
					{GoName: "IsActive", GoType: "bool", CallSuffix: "Boolean"},
					{GoName: "Priority", GoType: "int32", CallSuffix: "Int"},
					{GoName: "RawByte", GoType: "byte", CallSuffix: "Byte"},
					{GoName: "Level", GoType: "int16", CallSuffix: "Short"},
				},
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	if len(data.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(data.Messages))
	}

	msg := data.Messages[0]
	if msg.Name != "Location" {
		t.Errorf("expected message name Location, got %s", msg.Name)
	}

	expectedFields := []struct {
		name  string
		ptype string
	}{
		{"latitude", "double"},
		{"longitude", "double"},
		{"accuracy", "float"},
		{"time", "int64"},
		{"provider", "string"},
		{"speed", "float"},
		{"is_active", "bool"},
		{"priority", "int32"},
		{"raw_byte", "uint32"},
		{"level", "int32"},
	}

	if len(msg.Fields) != len(expectedFields) {
		t.Fatalf("expected %d fields, got %d", len(expectedFields), len(msg.Fields))
	}

	for i, exp := range expectedFields {
		f := msg.Fields[i]
		if f.Name != exp.name {
			t.Errorf("field %d: expected name %q, got %q", i, exp.name, f.Name)
		}
		if f.Type != exp.ptype {
			t.Errorf("field %d (%s): expected type %q, got %q", i, exp.name, exp.ptype, f.Type)
		}
		if f.Number != i+1 {
			t.Errorf("field %d: expected number %d, got %d", i, i+1, f.Number)
		}
	}
}

func TestBuildProtoData_DataClassCrossReference(t *testing.T) {
	// When a data class field references another data class, use the message type.
	merged := &javagen.MergedSpec{
		Package: "test",
		DataClasses: []javagen.MergedDataClass{
			{
				GoType:    "Device",
				JavaClass: "android.bluetooth.BluetoothDevice",
				Fields: []javagen.MergedField{
					{GoName: "Name", GoType: "string", CallSuffix: "Object"},
				},
			},
			{
				GoType:    "ScanResult",
				JavaClass: "android.bluetooth.le.ScanResult",
				Fields: []javagen.MergedField{
					{GoName: "Device", GoType: "Device", CallSuffix: "Object"},
					{GoName: "RSSI", GoType: "int32", CallSuffix: "Int"},
				},
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	if len(data.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(data.Messages))
	}

	scanResult := data.Messages[1]
	if scanResult.Name != "ScanResult" {
		t.Fatalf("expected ScanResult, got %s", scanResult.Name)
	}

	if scanResult.Fields[0].Type != "Device" {
		t.Errorf("expected Device field type to be 'Device', got %q", scanResult.Fields[0].Type)
	}
}

func TestBuildProtoData_ManagerServiceWithMethods(t *testing.T) {
	merged := &javagen.MergedSpec{
		Package: "location",
		DataClasses: []javagen.MergedDataClass{
			{
				GoType:    "Location",
				JavaClass: "android.location.Location",
				Fields: []javagen.MergedField{
					{GoName: "Latitude", GoType: "float64", CallSuffix: "Double"},
				},
			},
		},
		Classes: []javagen.MergedClass{
			{
				GoType: "Manager",
				Methods: []javagen.MergedMethod{
					{
						GoName:     "GetLastKnownLocation",
						Params:     []javagen.MergedParam{{GoName: "provider", GoType: "string", IsString: true}},
						Returns:    "android.location.Location",
						GoReturn:   "*jni.Object",
						CallSuffix: "Object",
						ReturnKind: javagen.ReturnObject,
					},
					{
						GoName:     "IsProviderEnabled",
						Params:     []javagen.MergedParam{{GoName: "provider", GoType: "string", IsString: true}},
						Returns:    "boolean",
						GoReturn:   "bool",
						CallSuffix: "Boolean",
						ReturnKind: javagen.ReturnBool,
					},
					{
						GoName:     "RemoveUpdates",
						Params:     []javagen.MergedParam{{GoName: "listener", GoType: "*jni.Object", IsObject: true}},
						Returns:    "void",
						GoReturn:   "",
						CallSuffix: "Void",
						ReturnKind: javagen.ReturnVoid,
					},
				},
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	// Should have 1 service.
	if len(data.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(data.Services))
	}

	svc := data.Services[0]
	if svc.Name != "ManagerService" {
		t.Errorf("expected ManagerService, got %s", svc.Name)
	}

	if len(svc.RPCs) != 3 {
		t.Fatalf("expected 3 RPCs, got %d", len(svc.RPCs))
	}

	// Check GetLastKnownLocation RPC.
	rpc0 := svc.RPCs[0]
	if rpc0.Name != "GetLastKnownLocation" {
		t.Errorf("expected RPC name GetLastKnownLocation, got %s", rpc0.Name)
	}
	if rpc0.InputType != "GetLastKnownLocationRequest" {
		t.Errorf("expected input type GetLastKnownLocationRequest, got %s", rpc0.InputType)
	}
	if rpc0.OutputType != "GetLastKnownLocationResponse" {
		t.Errorf("expected output type GetLastKnownLocationResponse, got %s", rpc0.OutputType)
	}

	// Find the response message and check it references the Location message.
	var resp *ProtoMessage
	for _, msg := range data.Messages {
		if msg.Name == "GetLastKnownLocationResponse" {
			resp = &msg
			break
		}
	}
	if resp == nil {
		t.Fatal("GetLastKnownLocationResponse message not found")
	}
	if len(resp.Fields) != 1 || resp.Fields[0].Type != "Location" {
		t.Errorf("expected response field type Location, got %v", resp.Fields)
	}

	// Check void method produces empty response.
	var voidResp *ProtoMessage
	for _, msg := range data.Messages {
		if msg.Name == "RemoveUpdatesResponse" {
			voidResp = &msg
			break
		}
	}
	if voidResp == nil {
		t.Fatal("RemoveUpdatesResponse message not found")
	}
	if len(voidResp.Fields) != 0 {
		t.Errorf("expected empty response for void method, got %d fields", len(voidResp.Fields))
	}

	// Check request message param types.
	var req *ProtoMessage
	for _, msg := range data.Messages {
		if msg.Name == "RemoveUpdatesRequest" {
			req = &msg
			break
		}
	}
	if req == nil {
		t.Fatal("RemoveUpdatesRequest message not found")
	}
	if len(req.Fields) != 1 || req.Fields[0].Type != "int64" {
		t.Errorf("expected object param type int64, got %v", req.Fields)
	}
}

func TestBuildProtoData_SkipsIterableDataClass(t *testing.T) {
	merged := &javagen.MergedSpec{
		Package: "location",
		Classes: []javagen.MergedClass{
			{
				GoType: "Manager",
				Kind:   "",
				Methods: []javagen.MergedMethod{
					{
						GoName:     "DoSomething",
						Returns:    "void",
						ReturnKind: javagen.ReturnVoid,
						CallSuffix: "Void",
					},
				},
			},
			{
				GoType:  "gnssStatus",
				Kind:    "iterable_data",
				Methods: nil,
			},
			{
				GoType:  "settingsBuilder",
				Kind:    "builder",
				Methods: nil,
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	if len(data.Services) != 1 {
		t.Fatalf("expected 1 service (Manager only), got %d", len(data.Services))
	}
	if data.Services[0].Name != "ManagerService" {
		t.Errorf("expected ManagerService, got %s", data.Services[0].Name)
	}
}

func TestBuildProtoData_Callbacks(t *testing.T) {
	merged := &javagen.MergedSpec{
		Package: "location",
		Callbacks: []javagen.MergedCallback{
			{
				JavaInterface: "android.location.LocationListener",
				GoType:        "locationListener",
				Methods: []javagen.MergedCallbackMethod{
					{
						JavaMethod: "onLocationChanged",
						GoField:    "OnLocation",
						Params: []javagen.MergedParam{
							{GoName: "arg0", GoType: "*jni.Object", IsObject: true},
						},
					},
					{
						JavaMethod: "onProviderEnabled",
						GoField:    "OnProviderEnabled",
						Params: []javagen.MergedParam{
							{GoName: "arg0", GoType: "string", IsString: true},
						},
					},
				},
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	// Should have a service with a streaming RPC.
	if len(data.Services) != 1 {
		t.Fatalf("expected 1 service from callback, got %d", len(data.Services))
	}

	svc := data.Services[0]
	if svc.Name != "LocationListenerService" {
		t.Errorf("expected LocationListenerService, got %s", svc.Name)
	}

	if len(svc.RPCs) != 1 {
		t.Fatalf("expected 1 RPC, got %d", len(svc.RPCs))
	}

	rpc := svc.RPCs[0]
	if rpc.Name != "SubscribeLocationListener" {
		t.Errorf("expected RPC name SubscribeLocationListener, got %s", rpc.Name)
	}
	if !rpc.ServerStreaming {
		t.Error("expected ServerStreaming to be true")
	}
	if rpc.ClientStreaming {
		t.Error("expected ClientStreaming to be false for server-streaming")
	}

	// Verify event messages were created.
	msgNames := make(map[string]bool)
	for _, msg := range data.Messages {
		msgNames[msg.Name] = true
	}

	expectedMsgs := []string{
		"LocationListenerOnLocationEvent",
		"LocationListenerOnProviderEnabledEvent",
		"LocationListenerEvent",
		"SubscribeLocationListenerRequest",
	}
	for _, name := range expectedMsgs {
		if !msgNames[name] {
			t.Errorf("expected message %s not found", name)
		}
	}
}

func TestBuildProtoData_BidiCallbacks(t *testing.T) {
	merged := &javagen.MergedSpec{
		Package: "bluetooth",
		Callbacks: []javagen.MergedCallback{
			{
				JavaInterface: "android.bluetooth.BluetoothGattCallback",
				GoType:        "gattCallback",
				Methods: []javagen.MergedCallbackMethod{
					{JavaMethod: "onConnectionStateChange", GoField: "OnConnectionStateChange"},
					{JavaMethod: "onCharacteristicRead", GoField: "OnCharacteristicRead"},
					{JavaMethod: "onCharacteristicWrite", GoField: "OnCharacteristicWrite"},
				},
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	if len(data.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(data.Services))
	}

	rpc := data.Services[0].RPCs[0]
	if rpc.Name != "GattCallbackStream" {
		t.Errorf("expected GattCallbackStream, got %s", rpc.Name)
	}
	if !rpc.ClientStreaming {
		t.Error("expected ClientStreaming true for bidi")
	}
	if !rpc.ServerStreaming {
		t.Error("expected ServerStreaming true for bidi")
	}

	// Check that a Command message was created for bidi.
	msgNames := make(map[string]bool)
	for _, msg := range data.Messages {
		msgNames[msg.Name] = true
	}
	if !msgNames["GattCallbackCommand"] {
		t.Error("expected GattCallbackCommand message for bidi streaming")
	}
}

func TestBuildProtoData_PackageAndGoPackage(t *testing.T) {
	merged := &javagen.MergedSpec{Package: "location"}
	data := BuildProtoData(merged, "github.com/example/jni")

	if data.Package != "location" {
		t.Errorf("expected package location, got %s", data.Package)
	}
	if data.GoPackage != "github.com/example/jni/proto/location" {
		t.Errorf("expected go_package github.com/example/jni/proto/location, got %s", data.GoPackage)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GetLastKnownLocation", "get_last_known_location"},
		{"provider", "provider"},
		{"IsProviderEnabled", "is_provider_enabled"},
		{"minTimeMs", "min_time_ms"},
		{"UUID", "uuid"},
		{"Latitude", "latitude"},
		{"OnSatelliteStatus", "on_satellite_status"},
		{"RSSI", "rssi"},
		{"enabledOnly", "enabled_only"},
		{"arg0", "arg0"},
	}

	for _, tt := range tests {
		got := toSnakeCase(tt.input)
		if got != tt.expected {
			t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildProtoData_CapitalizesNames(t *testing.T) {
	merged := &javagen.MergedSpec{
		Package: "bluetooth",
		Classes: []javagen.MergedClass{
			{
				GoType: "leScanner",
				Methods: []javagen.MergedMethod{
					{
						GoName:     "startScan",
						Returns:    "void",
						ReturnKind: javagen.ReturnVoid,
						CallSuffix: "Void",
					},
					{
						GoName:     "getScanResults",
						Returns:    "int",
						GoReturn:   "int32",
						CallSuffix: "Int",
						ReturnKind: javagen.ReturnPrimitive,
					},
				},
			},
		},
	}

	data := BuildProtoData(merged, "example.com/mod")

	if len(data.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(data.Services))
	}

	svc := data.Services[0]
	if svc.Name != "LeScannerService" {
		t.Errorf("expected service name LeScannerService, got %s", svc.Name)
	}

	if len(svc.RPCs) != 2 {
		t.Fatalf("expected 2 RPCs, got %d", len(svc.RPCs))
	}

	rpc0 := svc.RPCs[0]
	if rpc0.Name != "StartScan" {
		t.Errorf("expected RPC name StartScan, got %s", rpc0.Name)
	}
	if rpc0.InputType != "StartScanRequest" {
		t.Errorf("expected input type StartScanRequest, got %s", rpc0.InputType)
	}
	if rpc0.OutputType != "StartScanResponse" {
		t.Errorf("expected output type StartScanResponse, got %s", rpc0.OutputType)
	}

	// Verify message names are also capitalized.
	msgNames := make(map[string]bool)
	for _, msg := range data.Messages {
		msgNames[msg.Name] = true
	}
	for _, expected := range []string{
		"StartScanRequest", "StartScanResponse",
		"GetScanResultsRequest", "GetScanResultsResponse",
	} {
		if !msgNames[expected] {
			t.Errorf("expected message %s not found in %v", expected, msgNames)
		}
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"leScanner", "LeScanner"},
		{"Manager", "Manager"},
		{"getProviders", "GetProviders"},
		{"", ""},
		{"a", "A"},
	}
	for _, tt := range tests {
		got := capitalizeFirst(tt.input)
		if got != tt.expected {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestProtoTypeFromCallSuffix(t *testing.T) {
	tests := []struct {
		suffix   string
		goType   string
		expected string
	}{
		{"Boolean", "bool", "bool"},
		{"Byte", "byte", "uint32"},
		{"Short", "int16", "int32"},
		{"Int", "int32", "int32"},
		{"Long", "int64", "int64"},
		{"Float", "float32", "float"},
		{"Double", "float64", "double"},
		{"Object", "string", "string"},
		{"Object", "*jni.Object", "int64"},
	}

	for _, tt := range tests {
		got := protoTypeFromCallSuffix(tt.suffix, tt.goType)
		if got != tt.expected {
			t.Errorf("protoTypeFromCallSuffix(%q, %q) = %q, want %q",
				tt.suffix, tt.goType, got, tt.expected)
		}
	}
}

func TestProtoTypeFromParam(t *testing.T) {
	tests := []struct {
		name     string
		param    javagen.MergedParam
		expected string
	}{
		{
			name:     "string param",
			param:    javagen.MergedParam{GoType: "string", IsString: true},
			expected: "string",
		},
		{
			name:     "bool param",
			param:    javagen.MergedParam{GoType: "bool", IsBool: true},
			expected: "bool",
		},
		{
			name:     "object param",
			param:    javagen.MergedParam{GoType: "*jni.Object", IsObject: true},
			expected: "int64",
		},
		{
			name:     "int param",
			param:    javagen.MergedParam{GoType: "int32"},
			expected: "int32",
		},
		{
			name:     "long param",
			param:    javagen.MergedParam{GoType: "int64"},
			expected: "int64",
		},
		{
			name:     "float param",
			param:    javagen.MergedParam{GoType: "float32"},
			expected: "float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := protoTypeFromParam(tt.param)
			if got != tt.expected {
				t.Errorf("protoTypeFromParam() = %q, want %q", got, tt.expected)
			}
		})
	}
}
