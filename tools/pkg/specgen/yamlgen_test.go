package specgen

import (
	"testing"
)

func TestInferGoType(t *testing.T) {
	for _, tc := range []struct {
		name      string
		fullClass string
		goPkg     string
		want      string
	}{
		{
			name:      "simple class",
			fullClass: "android.app.Activity",
			goPkg:     "app",
			want:      "Activity",
		},
		{
			name:      "strip package prefix",
			fullClass: "android.app.alarm.AlarmManager",
			goPkg:     "alarm",
			want:      "Manager",
		},
		{
			name:      "no strip when no prefix match",
			fullClass: "android.app.SearchManager",
			goPkg:     "app",
			want:      "SearchManager",
		},
		{
			name:      "inner class",
			fullClass: "android.app.SearchManager$OnCancelListener",
			goPkg:     "app",
			want:      "SearchManagerOnCancelListener",
		},
		{
			name:      "case-sensitive prefix no strip",
			fullClass: "android.app.appsearch.AppSearchManager",
			goPkg:     "appsearch",
			want:      "AppSearchManager",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := inferGoType(tc.fullClass, tc.goPkg)
			if got != tc.want {
				t.Errorf("inferGoType(%q, %q) = %q, want %q", tc.fullClass, tc.goPkg, got, tc.want)
			}
		})
	}
}

func TestInferPackageMapping(t *testing.T) {
	module := "github.com/AndroidGoLab/jni"
	for _, tc := range []struct {
		name      string
		className string
		wantPkg   string
		wantPath  string
	}{
		{
			name:      "android.app direct",
			className: "android.app.Activity",
			wantPkg:   "app",
			wantPath:  module + "/app",
		},
		{
			name:      "android.app.appsearch subpackage",
			className: "android.app.appsearch.AppSearchManager",
			wantPkg:   "appsearch",
			wantPath:  module + "/app/appsearch",
		},
		{
			name:      "android.credentials",
			className: "android.credentials.CredentialManager",
			wantPkg:   "credentials",
			wantPath:  module + "/credentials",
		},
		{
			name:      "android.service.credentials separate from android.credentials",
			className: "android.service.credentials.ClearCredentialStateRequest",
			wantPkg:   "credentials",
			wantPath:  module + "/service/credentials",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := inferPackageMapping(tc.className, module)
			if got.Package != tc.wantPkg {
				t.Errorf("inferPackageMapping(%q).Package = %q, want %q", tc.className, got.Package, tc.wantPkg)
			}
			if got.GoImport != tc.wantPath {
				t.Errorf("inferPackageMapping(%q).GoImport = %q, want %q", tc.className, got.GoImport, tc.wantPath)
			}
		})
	}
}

func TestDeduplicateGoTypes(t *testing.T) {
	t.Run("no collision", func(t *testing.T) {
		classes := []SpecClass{
			{JavaClass: "android.app.Activity", GoType: "Activity"},
			{JavaClass: "android.app.SearchManager", GoType: "SearchManager"},
		}
		result := deduplicateGoTypes(classes)
		if result[0].GoType != "Activity" || result[1].GoType != "SearchManager" {
			t.Errorf("unexpected rename: %v", result)
		}
	})

	t.Run("collision resolved by restoring full name", func(t *testing.T) {
		classes := []SpecClass{
			{JavaClass: "android.net.ipsec.ike.IkeSaProposal", GoType: "SaProposal"},
			{JavaClass: "android.net.ipsec.ike.SaProposal", GoType: "SaProposal"},
		}
		result := deduplicateGoTypes(classes)
		if result[0].GoType != "IkeSaProposal" {
			t.Errorf("expected IkeSaProposal, got %q", result[0].GoType)
		}
		if result[1].GoType != "SaProposal" {
			t.Errorf("expected SaProposal, got %q", result[1].GoType)
		}
	})

	t.Run("inner class collision", func(t *testing.T) {
		classes := []SpecClass{
			{JavaClass: "com.example.Foo$Bar", GoType: "Bar"},
			{JavaClass: "com.example.Bar", GoType: "Bar"},
		}
		result := deduplicateGoTypes(classes)
		if result[0].GoType != "FooBar" {
			t.Errorf("expected FooBar, got %q", result[0].GoType)
		}
		if result[1].GoType != "Bar" {
			t.Errorf("expected Bar, got %q", result[1].GoType)
		}
	})
}
