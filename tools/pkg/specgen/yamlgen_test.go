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

func TestChooseBestConstructor(t *testing.T) {
	t.Run("prefers Context param", func(t *testing.T) {
		ctors := []JavapConstructor{
			{Params: nil},
			{Params: []JavapParam{{JavaType: "android.content.Context"}}},
			{Params: []JavapParam{{JavaType: "int"}, {JavaType: "int"}}},
		}
		best := chooseBestConstructor(ctors)
		if len(best.Params) != 1 || best.Params[0].JavaType != "android.content.Context" {
			t.Errorf("expected Context constructor, got %+v", best)
		}
	})

	t.Run("falls back to no-arg", func(t *testing.T) {
		ctors := []JavapConstructor{
			{Params: []JavapParam{{JavaType: "int"}, {JavaType: "int"}}},
			{Params: nil},
		}
		best := chooseBestConstructor(ctors)
		if len(best.Params) != 0 {
			t.Errorf("expected no-arg constructor, got %+v", best)
		}
	})

	t.Run("falls back to first", func(t *testing.T) {
		ctors := []JavapConstructor{
			{Params: []JavapParam{{JavaType: "int"}}},
			{Params: []JavapParam{{JavaType: "int"}, {JavaType: "int"}}},
		}
		best := chooseBestConstructor(ctors)
		if len(best.Params) != 1 {
			t.Errorf("expected first constructor (1 param), got %+v", best)
		}
	})
}

func TestClassFromJavap_ConstructorObtain(t *testing.T) {
	// Reset AndroidServiceName so the test doesn't depend on runtime state.
	origSvc := AndroidServiceName
	AndroidServiceName = nil
	defer func() { AndroidServiceName = origSvc }()

	t.Run("concrete class with constructors gets obtain=constructor", func(t *testing.T) {
		jc := &JavapClass{
			FullName: "android.media.MediaRecorder",
			Constructors: []JavapConstructor{
				{Params: nil},
				{Params: []JavapParam{{JavaType: "android.content.Context"}}},
			},
			Methods: []JavapMethod{
				{Name: "start", ReturnType: "void"},
			},
		}
		cls := classFromJavap(jc, "media")
		if cls.Obtain != "constructor" {
			t.Errorf("Obtain = %q, want %q", cls.Obtain, "constructor")
		}
		// Should pick the Context constructor.
		if len(cls.ConstructorParams) != 1 {
			t.Fatalf("len(ConstructorParams) = %d, want 1", len(cls.ConstructorParams))
		}
		if cls.ConstructorParams[0].JavaType != "android.content.Context" {
			t.Errorf("ConstructorParams[0].JavaType = %q, want %q",
				cls.ConstructorParams[0].JavaType, "android.content.Context")
		}
	})

	t.Run("abstract class does not get obtain=constructor", func(t *testing.T) {
		jc := &JavapClass{
			FullName:   "android.app.AbstractThing",
			IsAbstract: true,
			Constructors: []JavapConstructor{
				{Params: nil},
			},
		}
		cls := classFromJavap(jc, "app")
		if cls.Obtain != "" {
			t.Errorf("Obtain = %q, want empty for abstract class", cls.Obtain)
		}
	})

	t.Run("interface does not get obtain=constructor", func(t *testing.T) {
		jc := &JavapClass{
			FullName:    "android.app.SomeInterface",
			IsInterface: true,
		}
		cls := classFromJavap(jc, "app")
		if cls.Obtain != "" {
			t.Errorf("Obtain = %q, want empty for interface", cls.Obtain)
		}
	})

	t.Run("system service class keeps obtain=system_service", func(t *testing.T) {
		AndroidServiceName = map[string]string{
			"android.app.AlarmManager": "alarm",
		}
		jc := &JavapClass{
			FullName: "android.app.AlarmManager",
			Constructors: []JavapConstructor{
				{Params: nil},
			},
		}
		cls := classFromJavap(jc, "alarm")
		if cls.Obtain != "system_service" {
			t.Errorf("Obtain = %q, want %q", cls.Obtain, "system_service")
		}
		if cls.ServiceName != "alarm" {
			t.Errorf("ServiceName = %q, want %q", cls.ServiceName, "alarm")
		}
	})
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
