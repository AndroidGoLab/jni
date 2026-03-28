package javagen

import (
	"testing"
)

func TestMergedAbstractCallbackMethod_VoidReturn(t *testing.T) {
	m := MergedAbstractCallbackMethod{
		JavaMethod: "onEvent",
		Returns:    "void",
		Params:     []MergedParam{{JavaType: "int", GoName: "arg0"}},
	}

	if m.JavaReturnType() != "void" {
		t.Errorf("JavaReturnType = %q, want void", m.JavaReturnType())
	}
	if m.HasReturn() {
		t.Error("HasReturn should be false for void")
	}
}

func TestMergedAbstractCallbackMethod_IntReturn(t *testing.T) {
	m := MergedAbstractCallbackMethod{
		JavaMethod: "getCount",
		Returns:    "int",
		Params:     nil,
	}

	if m.JavaReturnType() != "int" {
		t.Errorf("JavaReturnType = %q, want int", m.JavaReturnType())
	}
	if !m.HasReturn() {
		t.Error("HasReturn should be true for int")
	}
	if m.JavaCastReturn() != "(Integer)" {
		t.Errorf("JavaCastReturn = %q, want (Integer)", m.JavaCastReturn())
	}
	if m.JavaUnboxReturn() != ".intValue()" {
		t.Errorf("JavaUnboxReturn = %q, want .intValue()", m.JavaUnboxReturn())
	}
}

func TestMergedAbstractCallbackMethod_BooleanReturn(t *testing.T) {
	m := MergedAbstractCallbackMethod{Returns: "boolean"}
	if m.JavaCastReturn() != "(Boolean)" {
		t.Errorf("JavaCastReturn = %q", m.JavaCastReturn())
	}
	if m.JavaUnboxReturn() != ".booleanValue()" {
		t.Errorf("JavaUnboxReturn = %q", m.JavaUnboxReturn())
	}
}

func TestMergedAbstractCallbackMethod_ObjectReturn(t *testing.T) {
	m := MergedAbstractCallbackMethod{Returns: "android.bluetooth.BluetoothDevice"}
	if m.JavaCastReturn() != "(android.bluetooth.BluetoothDevice)" {
		t.Errorf("JavaCastReturn = %q", m.JavaCastReturn())
	}
	if m.JavaUnboxReturn() != "" {
		t.Errorf("JavaUnboxReturn should be empty for objects, got %q", m.JavaUnboxReturn())
	}
}

func TestMergedAbstractCallbackMethod_JavaParamList(t *testing.T) {
	m := MergedAbstractCallbackMethod{
		Params: []MergedParam{
			{JavaType: "int", GoName: "arg0"},
			{JavaType: "android.bluetooth.le.ScanResult", GoName: "arg1"},
		},
	}

	want := "int arg0, android.bluetooth.le.ScanResult arg1"
	if got := m.JavaParamList(); got != want {
		t.Errorf("JavaParamList = %q, want %q", got, want)
	}
}

func TestMergedAbstractCallbackMethod_JavaArgList(t *testing.T) {
	m := MergedAbstractCallbackMethod{
		Params: []MergedParam{
			{JavaType: "int", GoName: "arg0"},
			{JavaType: "android.bluetooth.le.ScanResult", GoName: "arg1"},
			{JavaType: "boolean", GoName: "arg2"},
		},
	}

	want := "Integer.valueOf(arg0), arg1, Boolean.valueOf(arg2)"
	if got := m.JavaArgList(); got != want {
		t.Errorf("JavaArgList = %q, want %q", got, want)
	}
}

func TestMergedAbstractCallbackMethod_EmptyParams(t *testing.T) {
	m := MergedAbstractCallbackMethod{Params: nil}
	if got := m.JavaParamList(); got != "" {
		t.Errorf("JavaParamList for no params = %q, want empty", got)
	}
	if got := m.JavaArgList(); got != "" {
		t.Errorf("JavaArgList for no params = %q, want empty", got)
	}
}

func TestMergedAbstractCallback_AdapterClassName(t *testing.T) {
	tests := []struct {
		javaClass string
		want      string
	}{
		{"android.bluetooth.le.ScanCallback", "ScanCallbackAdapter"},
		{"android.bluetooth.BluetoothGattCallback", "BluetoothGattCallbackAdapter"},
		{"SomeCallback", "SomeCallbackAdapter"},
	}
	for _, tt := range tests {
		acb := MergedAbstractCallback{JavaClass: tt.javaClass}
		if got := acb.AdapterClassName(); got != tt.want {
			t.Errorf("AdapterClassName(%q) = %q, want %q", tt.javaClass, got, tt.want)
		}
	}
}

func TestMergedAbstractCallback_JavaPackage(t *testing.T) {
	tests := []struct {
		javaClass string
		want      string
	}{
		{"android.bluetooth.le.ScanCallback", "android.bluetooth.le"},
		{"SomeCallback", ""},
	}
	for _, tt := range tests {
		acb := MergedAbstractCallback{JavaClass: tt.javaClass}
		if got := acb.JavaPackage(); got != tt.want {
			t.Errorf("JavaPackage(%q) = %q, want %q", tt.javaClass, got, tt.want)
		}
	}
}
