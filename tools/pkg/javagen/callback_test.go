package javagen

import (
	"strings"
	"testing"
)

func TestBuildCallbackType_WithParams(t *testing.T) {
	cb := &MergedCallback{
		GoType: "Listener",
		Methods: []MergedCallbackMethod{
			{GoField: "OnEvent", GoParams: "arg0 string, arg1 int32"},
			{GoField: "OnDone", GoParams: ""},
		},
	}

	code := BuildCallbackType(cb)
	if !strings.Contains(code, "type Listener struct") {
		t.Error("missing struct declaration")
	}
	if !strings.Contains(code, "OnEvent func(arg0 string, arg1 int32)") {
		t.Errorf("missing OnEvent field, got:\n%s", code)
	}
	if !strings.Contains(code, "OnDone func()") {
		t.Errorf("missing OnDone field, got:\n%s", code)
	}
}

func TestBuildCallbackDispatch_WithParams(t *testing.T) {
	cb := &MergedCallback{
		GoType: "Listener",
		Methods: []MergedCallbackMethod{
			{
				JavaMethod: "onLocationChanged",
				GoField:    "OnLocation",
				Params: []MergedParam{
					{JavaType: "android.location.Location", GoName: "arg0", GoType: "*jni.Object", IsObject: true},
				},
			},
			{
				JavaMethod: "onDone",
				GoField:    "OnDone",
				Params:     nil,
			},
		},
	}

	code := BuildCallbackDispatch(cb)
	if !strings.Contains(code, "switch methodName") {
		t.Error("missing switch statement")
	}
	if !strings.Contains(code, `case "onLocationChanged"`) {
		t.Error("missing onLocationChanged case")
	}
	if !strings.Contains(code, "cb.OnLocation") {
		t.Error("missing OnLocation dispatch")
	}
	if !strings.Contains(code, `case "onDone"`) {
		t.Error("missing onDone case")
	}
	if !strings.Contains(code, "cb.OnDone()") {
		t.Error("missing OnDone() call for no-param method")
	}
}

func TestBuildCallbackDispatch_StringParam(t *testing.T) {
	cb := &MergedCallback{
		GoType: "Listener",
		Methods: []MergedCallbackMethod{
			{
				JavaMethod: "onMessage",
				GoField:    "OnMessage",
				Params: []MergedParam{
					{JavaType: "String", GoName: "arg0", GoType: "string", IsString: true},
				},
			},
		},
	}

	code := BuildCallbackDispatch(cb)
	if !strings.Contains(code, "env.GoString") {
		t.Errorf("expected GoString for string param, got:\n%s", code)
	}
}
