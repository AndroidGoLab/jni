package javagen

import (
	"testing"
)

func TestIsManager(t *testing.T) {
	tests := []struct {
		obtain string
		want   bool
	}{
		{"system_service", true},
		{"constructor", false},
		{"static_factory", false},
		{"", false},
	}
	for _, tt := range tests {
		cls := &MergedClass{Obtain: tt.obtain}
		got := IsManager(cls)
		if got != tt.want {
			t.Errorf("IsManager(obtain=%q) = %v, want %v", tt.obtain, got, tt.want)
		}
	}
}

func TestManagerConstructorData(t *testing.T) {
	cls := &MergedClass{
		GoType:      "Manager",
		ServiceName: "location",
		Close:       true,
	}
	data := ManagerConstructorData(cls)
	if data.GoType != "Manager" {
		t.Errorf("GoType = %q", data.GoType)
	}
	if data.ServiceName != "location" {
		t.Errorf("ServiceName = %q", data.ServiceName)
	}
	if !data.HasClose {
		t.Error("expected HasClose=true")
	}
}
