//go:build !android

package e2e_grpc_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// These tests require a running jniservice on an Android device or emulator.
// Set JNICTL_E2E_ADDR=<host:port> to enable them.
// Optionally set JNICTL_BIN to the path of a pre-built jnicli binary.
//
// Run with:
//   JNICTL_E2E_ADDR=localhost:50051 go test -v -count=1 -run TestE2E ./

func skipIfNoEmulator(t *testing.T) {
	t.Helper()
	if os.Getenv("JNICTL_E2E_ADDR") == "" {
		t.Skip("JNICTL_E2E_ADDR not set; skipping E2E tests")
	}
}

// runLiveJnicli runs jnicli against the live server and returns stdout.
// Fails the test on non-zero exit.
func runLiveJnicli(t *testing.T, args ...string) string {
	t.Helper()
	addr := os.Getenv("JNICTL_E2E_ADDR")
	fullArgs := append([]string{"--addr", addr, "--insecure"}, mtlsFlags()...)
	fullArgs = append(fullArgs, args...)

	cmd := jnicliCommand(fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("jnicli %s: %v\nstdout: %s\nstderr: %s",
			strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

// runLiveJnicliExpectError runs jnicli and expects a non-zero exit code.
func runLiveJnicliExpectError(t *testing.T, args ...string) string {
	t.Helper()
	addr := os.Getenv("JNICTL_E2E_ADDR")
	fullArgs := append([]string{"--addr", addr, "--insecure"}, mtlsFlags()...)
	fullArgs = append(fullArgs, args...)

	cmd := jnicliCommand(fullArgs...)

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error from jnicli %s but got success:\n%s",
			strings.Join(args, " "), out)
	}
	return string(out)
}

// mtlsFlags returns --cert/--key/--ca flags if TestMain registered
// a client certificate. Returns nil when no certs are available
// (e.g. when JNICTL_E2E_ADDR is not set and setup was skipped).
func mtlsFlags() []string {
	if testCertFile == "" {
		return nil
	}
	return []string{
		"--cert", testCertFile,
		"--key", testKeyFile,
		"--ca", testCAFile,
	}
}

// jnicliCommand returns an exec.Cmd for jnicli.
// Uses JNICTL_BIN if set, otherwise falls back to "go run".
func jnicliCommand(args ...string) *exec.Cmd {
	if bin := os.Getenv("JNICTL_BIN"); bin != "" {
		return exec.Command(bin, args...)
	}
	return exec.Command("go", append([]string{"run", jnicliBin}, args...)...)
}

// parseJSON parses a JSON string into a map.
func parseJSON(t *testing.T, out string) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("invalid JSON: %v\nraw output: %s", err, out)
	}
	return resp
}

// getInt64Field extracts an int64 field from a JSON response.
// Handles both JSON number (int32 fields) and JSON string (int64 fields from protojson).
func getInt64Field(t *testing.T, resp map[string]any, field string) int64 {
	t.Helper()
	v, ok := resp[field]
	if !ok {
		t.Fatalf("missing field %q in response: %v", field, resp)
	}
	switch val := v.(type) {
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			t.Fatalf("field %q: not a valid int64: %q", field, val)
		}
		return n
	case float64:
		return int64(val)
	default:
		t.Fatalf("field %q: unexpected type %T: %v", field, v, v)
		return 0
	}
}

// ---- Connectivity ----

func TestE2E_JNIGetVersion(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnicli(t, "jni", "get-version")
	resp := parseJSON(t, out)

	version, ok := resp["version"]
	if !ok {
		t.Fatal("missing 'version' field in response")
	}
	v, ok := version.(float64)
	if !ok {
		t.Fatalf("version is not a number: %T %v", version, version)
	}
	if v <= 0 {
		t.Errorf("expected version > 0, got %v", v)
	}
	t.Logf("JNI version: %v (0x%x)", v, int(v))
}

// ---- Class Operations ----

func TestE2E_FindClass(t *testing.T) {
	skipIfNoEmulator(t)

	t.Run("java/lang/String", func(t *testing.T) {
		out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
		resp := parseJSON(t, out)
		handle := getInt64Field(t, resp, "classHandle")
		if handle <= 0 {
			t.Errorf("expected classHandle > 0, got %d", handle)
		}
		t.Logf("String class handle: %d", handle)
	})

	t.Run("java/lang/System", func(t *testing.T) {
		out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/System")
		resp := parseJSON(t, out)
		handle := getInt64Field(t, resp, "classHandle")
		if handle <= 0 {
			t.Errorf("expected classHandle > 0, got %d", handle)
		}
		t.Logf("System class handle: %d", handle)
	})

	t.Run("java/lang/Integer", func(t *testing.T) {
		out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/Integer")
		resp := parseJSON(t, out)
		handle := getInt64Field(t, resp, "classHandle")
		if handle <= 0 {
			t.Errorf("expected classHandle > 0, got %d", handle)
		}
		t.Logf("Integer class handle: %d", handle)
	})
}

func TestE2E_FindClassNotFound(t *testing.T) {
	skipIfNoEmulator(t)
	out := runLiveJnicliExpectError(t, "jni", "class", "find", "--name", "does/not/Exist")
	t.Logf("expected error output: %s", out)
}

// ---- Method Lookup ----

func TestE2E_GetStaticMethodID(t *testing.T) {
	skipIfNoEmulator(t)

	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/System")
	classHandle := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "method", "get-static-id",
		"--class", strconv.FormatInt(classHandle, 10),
		"--name", "currentTimeMillis",
		"--sig", "()J")
	resp := parseJSON(t, out)
	methodID := getInt64Field(t, resp, "methodId")
	if methodID == 0 {
		t.Error("expected methodId != 0")
	}
	t.Logf("currentTimeMillis method ID: %d", methodID)
}

// ---- Method Calls ----

func TestE2E_CallStaticMethod_CurrentTimeMillis(t *testing.T) {
	skipIfNoEmulator(t)

	// Find java/lang/System.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/System")
	classHandle := getInt64Field(t, parseJSON(t, out), "classHandle")

	// Get currentTimeMillis method ID.
	out = runLiveJnicli(t, "jni", "method", "get-static-id",
		"--class", strconv.FormatInt(classHandle, 10),
		"--name", "currentTimeMillis",
		"--sig", "()J")
	methodID := getInt64Field(t, parseJSON(t, out), "methodId")

	// Call it.
	out = runLiveJnicli(t, "jni", "method", "call-static",
		"--class", strconv.FormatInt(classHandle, 10),
		"--method", strconv.FormatInt(methodID, 10),
		"--return-type", "long")
	resp := parseJSON(t, out)

	// Response has result.j (long value).
	result, ok := resp["result"]
	if !ok {
		t.Fatal("missing 'result' field in response")
	}
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result is not an object: %T %v", result, result)
	}
	timeMillis := getInt64Field(t, resultMap, "j")
	if timeMillis <= 0 {
		t.Errorf("expected currentTimeMillis > 0, got %d", timeMillis)
	}
	t.Logf("currentTimeMillis: %d", timeMillis)
}

// ---- String Operations ----

func TestE2E_StringNewAndGet(t *testing.T) {
	skipIfNoEmulator(t)

	const testString = "hello world"

	// Create a new Java string.
	out := runLiveJnicli(t, "jni", "string", "new", "--value", testString)
	resp := parseJSON(t, out)
	handle := getInt64Field(t, resp, "stringHandle")
	if handle <= 0 {
		t.Fatalf("expected stringHandle > 0, got %d", handle)
	}

	// Read it back.
	out = runLiveJnicli(t, "jni", "string", "get",
		"--handle", strconv.FormatInt(handle, 10))
	resp = parseJSON(t, out)
	value, ok := resp["value"]
	if !ok {
		t.Fatal("missing 'value' field")
	}
	if value != testString {
		t.Errorf("expected value %q, got %q", testString, value)
	}
	t.Logf("string round-trip: %q", value)
}

func TestE2E_StringLength(t *testing.T) {
	skipIfNoEmulator(t)

	const testString = "hello"

	out := runLiveJnicli(t, "jni", "string", "new", "--value", testString)
	handle := getInt64Field(t, parseJSON(t, out), "stringHandle")

	out = runLiveJnicli(t, "jni", "string", "length",
		"--handle", strconv.FormatInt(handle, 10))
	resp := parseJSON(t, out)

	length, ok := resp["length"]
	if !ok {
		t.Fatal("missing 'length' field")
	}
	l, ok := length.(float64)
	if !ok {
		t.Fatalf("length is not a number: %T %v", length, length)
	}
	if int(l) != len(testString) {
		t.Errorf("expected length %d, got %v", len(testString), l)
	}
	t.Logf("string length: %v", l)
}

// ---- Object Operations ----

func TestE2E_ObjectIsSame(t *testing.T) {
	skipIfNoEmulator(t)

	// Get a handle via FindClass.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	handle := getInt64Field(t, parseJSON(t, out), "classHandle")

	// Same handle must be same object.
	out = runLiveJnicli(t, "jni", "object", "is-same",
		"--object1", strconv.FormatInt(handle, 10),
		"--object2", strconv.FormatInt(handle, 10))
	resp := parseJSON(t, out)

	result, ok := resp["result"]
	if !ok {
		// protojson omits false booleans; missing = false which is wrong here.
		t.Fatal("missing 'result' field — expected true")
	}
	if result != true {
		t.Errorf("expected is-same to be true, got %v", result)
	}
}

func TestE2E_ObjectIsSame_DifferentObjects(t *testing.T) {
	skipIfNoEmulator(t)

	// Two different classes should not be the same object.
	out1 := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	h1 := getInt64Field(t, parseJSON(t, out1), "classHandle")

	out2 := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/Integer")
	h2 := getInt64Field(t, parseJSON(t, out2), "classHandle")

	out := runLiveJnicli(t, "jni", "object", "is-same",
		"--object1", strconv.FormatInt(h1, 10),
		"--object2", strconv.FormatInt(h2, 10))
	resp := parseJSON(t, out)

	// protojson omits false booleans, so missing "result" means false.
	if result, ok := resp["result"]; ok && result == true {
		t.Error("expected is-same to be false for String vs Integer classes")
	}
}

func TestE2E_ObjectGetClass(t *testing.T) {
	skipIfNoEmulator(t)

	// Create a string object and get its class.
	out := runLiveJnicli(t, "jni", "string", "new", "--value", "test")
	strHandle := getInt64Field(t, parseJSON(t, out), "stringHandle")

	out = runLiveJnicli(t, "jni", "object", "get-class",
		"--object", strconv.FormatInt(strHandle, 10))
	resp := parseJSON(t, out)
	classHandle := getInt64Field(t, resp, "classHandle")
	if classHandle <= 0 {
		t.Errorf("expected classHandle > 0, got %d", classHandle)
	}
	t.Logf("class of string object: %d", classHandle)
}

// ---- Handle Management ----

func TestE2E_HandleRelease(t *testing.T) {
	skipIfNoEmulator(t)

	// Create a disposable string.
	out := runLiveJnicli(t, "jni", "string", "new", "--value", "disposable")
	handle := getInt64Field(t, parseJSON(t, out), "stringHandle")

	// Release it — should succeed.
	out = runLiveJnicli(t, "handle", "release",
		"--handle", strconv.FormatInt(handle, 10))
	t.Logf("release response: %s", strings.TrimSpace(out))

	// Releasing again should fail (handle no longer exists).
	runLiveJnicliExpectError(t, "handle", "release",
		"--handle", strconv.FormatInt(handle, 10))
}

// ---- Exception Handling ----

func TestE2E_ExceptionCheck(t *testing.T) {
	skipIfNoEmulator(t)

	out := runLiveJnicli(t, "jni", "exception", "check")
	resp := parseJSON(t, out)

	// No exception should be pending after normal operations.
	// protojson omits false booleans, so missing "hasException" means false.
	if resp["hasException"] == true {
		t.Error("unexpected exception pending")
	}
}

// ---- Error Cases ----

func TestE2E_ErrorCases(t *testing.T) {
	skipIfNoEmulator(t)

	t.Run("FindClass_NonexistentClass", func(t *testing.T) {
		runLiveJnicliExpectError(t, "jni", "class", "find", "--name", "com/fake/DoesNotExist")
	})

	t.Run("GetMethodID_InvalidClass", func(t *testing.T) {
		// Handle 999999 doesn't exist — should return a gRPC error.
		runLiveJnicliExpectError(t, "jni", "method", "get-static-id",
			"--class", "999999",
			"--name", "foo",
			"--sig", "()V")
	})

	t.Run("GetMethodID_NonexistentMethod", func(t *testing.T) {
		// Valid class, but method doesn't exist.
		out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/System")
		classHandle := getInt64Field(t, parseJSON(t, out), "classHandle")

		runLiveJnicliExpectError(t, "jni", "method", "get-static-id",
			"--class", strconv.FormatInt(classHandle, 10),
			"--name", "nonExistentMethod",
			"--sig", "()V")
	})

	t.Run("GetString_InvalidHandle", func(t *testing.T) {
		runLiveJnicliExpectError(t, "jni", "string", "get", "--handle", "999999")
	})
}

// ---- Full Workflow ----

func TestE2E_FullWorkflow_CallAnyJavaMethod(t *testing.T) {
	skipIfNoEmulator(t)

	// 1. Find java/lang/System.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/System")
	classHandle := getInt64Field(t, parseJSON(t, out), "classHandle")
	t.Logf("step 1: System class handle=%d", classHandle)

	// 2. Get static method ID for currentTimeMillis.
	out = runLiveJnicli(t, "jni", "method", "get-static-id",
		"--class", strconv.FormatInt(classHandle, 10),
		"--name", "currentTimeMillis",
		"--sig", "()J")
	methodID := getInt64Field(t, parseJSON(t, out), "methodId")
	t.Logf("step 2: method ID=%d", methodID)

	// 3. Call currentTimeMillis.
	out = runLiveJnicli(t, "jni", "method", "call-static",
		"--class", strconv.FormatInt(classHandle, 10),
		"--method", strconv.FormatInt(methodID, 10),
		"--return-type", "long")
	resp := parseJSON(t, out)
	resultMap, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing or invalid 'result' field: %v", resp)
	}
	time1 := getInt64Field(t, resultMap, "j")
	t.Logf("step 3: currentTimeMillis=%d", time1)

	if time1 <= 0 {
		t.Fatalf("expected currentTimeMillis > 0, got %d", time1)
	}

	// 4. Call again — value should be >= previous.
	out = runLiveJnicli(t, "jni", "method", "call-static",
		"--class", strconv.FormatInt(classHandle, 10),
		"--method", strconv.FormatInt(methodID, 10),
		"--return-type", "long")
	resp = parseJSON(t, out)
	resultMap, ok = resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing or invalid 'result' field: %v", resp)
	}
	time2 := getInt64Field(t, resultMap, "j")
	t.Logf("step 4: currentTimeMillis=%d (delta=%dms)", time2, time2-time1)

	if time2 < time1 {
		t.Errorf("expected time2 >= time1, got time1=%d time2=%d", time1, time2)
	}

	// 5. Verify no exception pending.
	out = runLiveJnicli(t, "jni", "exception", "check")
	excResp := parseJSON(t, out)
	if excResp["hasException"] == true {
		t.Error("unexpected exception after calls")
	}
	t.Log("step 5: no exceptions pending")

	// 6. Create a string, read it back, verify.
	const msg = "e2e-workflow-test"
	out = runLiveJnicli(t, "jni", "string", "new", "--value", msg)
	strHandle := getInt64Field(t, parseJSON(t, out), "stringHandle")

	out = runLiveJnicli(t, "jni", "string", "get",
		"--handle", strconv.FormatInt(strHandle, 10))
	strResp := parseJSON(t, out)
	if strResp["value"] != msg {
		t.Errorf("string round-trip: expected %q, got %q", msg, strResp["value"])
	}
	t.Logf("step 6: string round-trip OK: %q", msg)

	// 7. Release handles.
	runLiveJnicli(t, "handle", "release",
		"--handle", strconv.FormatInt(classHandle, 10))
	runLiveJnicli(t, "handle", "release",
		"--handle", strconv.FormatInt(strHandle, 10))
	t.Log("step 7: handles released")
}

// ---- Superclass / IsAssignableFrom ----

func TestE2E_GetSuperclass(t *testing.T) {
	skipIfNoEmulator(t)

	// String's superclass should be Object.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	strClassHandle := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "class", "get-superclass",
		"--class", strconv.FormatInt(strClassHandle, 10))
	resp := parseJSON(t, out)
	superHandle := getInt64Field(t, resp, "classHandle")
	if superHandle <= 0 {
		t.Errorf("expected superclass handle > 0, got %d", superHandle)
	}
	t.Logf("String superclass handle: %d", superHandle)
}

func TestE2E_IsAssignableFrom(t *testing.T) {
	skipIfNoEmulator(t)

	// String is assignable from String.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	strClass := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "class", "is-assignable-from",
		"--class1", strconv.FormatInt(strClass, 10),
		"--class2", strconv.FormatInt(strClass, 10))
	resp := parseJSON(t, out)

	result, ok := resp["result"]
	if !ok {
		t.Fatal("missing 'result' field — expected true")
	}
	if result != true {
		t.Errorf("expected String assignable from String, got %v", result)
	}
}

// ---- IsInstanceOf ----

func TestE2E_IsInstanceOf(t *testing.T) {
	skipIfNoEmulator(t)

	// Create a string, check it's an instance of String class.
	out := runLiveJnicli(t, "jni", "string", "new", "--value", "test")
	strHandle := getInt64Field(t, parseJSON(t, out), "stringHandle")

	out = runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	strClass := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "object", "is-instance-of",
		"--object", strconv.FormatInt(strHandle, 10),
		"--class", strconv.FormatInt(strClass, 10))
	resp := parseJSON(t, out)

	result, ok := resp["result"]
	if !ok {
		t.Fatal("missing 'result' field — expected true")
	}
	if result != true {
		t.Errorf("expected string to be instance of String, got %v", result)
	}
}

// ---- GetMethodID (instance method) ----

func TestE2E_GetMethodID(t *testing.T) {
	skipIfNoEmulator(t)

	// String.length() instance method.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	strClass := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "method", "get-id",
		"--class", strconv.FormatInt(strClass, 10),
		"--name", "length",
		"--sig", "()I")
	resp := parseJSON(t, out)
	methodID := getInt64Field(t, resp, "methodId")
	if methodID == 0 {
		t.Error("expected methodId != 0 for String.length()")
	}
	t.Logf("String.length() method ID: %d", methodID)
}

// ---- CallMethod (instance) ----

func TestE2E_CallMethod_StringLength(t *testing.T) {
	skipIfNoEmulator(t)

	const testStr = "abcde"

	// Create a String object.
	out := runLiveJnicli(t, "jni", "string", "new", "--value", testStr)
	strHandle := getInt64Field(t, parseJSON(t, out), "stringHandle")

	// Find String class.
	out = runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/String")
	strClass := getInt64Field(t, parseJSON(t, out), "classHandle")

	// Get String.length() method ID.
	out = runLiveJnicli(t, "jni", "method", "get-id",
		"--class", strconv.FormatInt(strClass, 10),
		"--name", "length",
		"--sig", "()I")
	methodID := getInt64Field(t, parseJSON(t, out), "methodId")

	// Call String.length() on our string.
	out = runLiveJnicli(t, "jni", "method", "call",
		"--object", strconv.FormatInt(strHandle, 10),
		"--method", strconv.FormatInt(methodID, 10),
		"--return-type", "int")
	resp := parseJSON(t, out)

	resultMap, ok := resp["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing or invalid 'result' field: %v", resp)
	}
	length := getInt64Field(t, resultMap, "i")
	if int(length) != len(testStr) {
		t.Errorf("expected String.length() = %d, got %d", len(testStr), length)
	}
	t.Logf("String(\"%s\").length() = %d", testStr, length)
}

// ---- GetFieldID ----

func TestE2E_GetFieldID(t *testing.T) {
	skipIfNoEmulator(t)

	// Integer has a "value" field (private, but JNI doesn't enforce access).
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/Integer")
	cls := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "field", "get-id",
		"--class", strconv.FormatInt(cls, 10),
		"--name", "value",
		"--sig", "I")
	resp := parseJSON(t, out)
	fieldID := getInt64Field(t, resp, "fieldId")
	if fieldID == 0 {
		t.Error("expected fieldId != 0 for Integer.value")
	}
	t.Logf("Integer.value field ID: %d", fieldID)
}

// ---- GetStaticFieldID ----

func TestE2E_GetStaticFieldID(t *testing.T) {
	skipIfNoEmulator(t)

	// Integer.MAX_VALUE static field.
	out := runLiveJnicli(t, "jni", "class", "find", "--name", "java/lang/Integer")
	cls := getInt64Field(t, parseJSON(t, out), "classHandle")

	out = runLiveJnicli(t, "jni", "field", "get-static-id",
		"--class", strconv.FormatInt(cls, 10),
		"--name", "MAX_VALUE",
		"--sig", "I")
	resp := parseJSON(t, out)
	fieldID := getInt64Field(t, resp, "fieldId")
	if fieldID == 0 {
		t.Error("expected fieldId != 0 for Integer.MAX_VALUE")
	}
	t.Logf("Integer.MAX_VALUE field ID: %d", fieldID)

	// Verify the static field value using call: Integer.MAX_VALUE = 2147483647.
	// This uses GetStaticField which may not be implemented yet.
	// If it is, verify the value.
	out2 := runLiveJnicli(t, "jni", "field", "get-static",
		"--class", strconv.FormatInt(cls, 10),
		"--field", strconv.FormatInt(fieldID, 10),
		"--type", "int")
	resp2 := parseJSON(t, out2)
	resultMap, ok := resp2["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing 'result' in GetStaticField response: %v", resp2)
	}
	val := getInt64Field(t, resultMap, "i")
	const expectedMaxValue = 2147483647
	if val != expectedMaxValue {
		t.Errorf("expected Integer.MAX_VALUE = %d, got %d", expectedMaxValue, val)
	}
	t.Logf("Integer.MAX_VALUE = %d", val)
}

// ---- Describe all test names for documentation ----

func TestE2E_SummaryOfTests(t *testing.T) {
	skipIfNoEmulator(t)
	tests := []string{
		"TestE2E_JNIGetVersion",
		"TestE2E_FindClass",
		"TestE2E_FindClassNotFound",
		"TestE2E_GetStaticMethodID",
		"TestE2E_CallStaticMethod_CurrentTimeMillis",
		"TestE2E_StringNewAndGet",
		"TestE2E_StringLength",
		"TestE2E_ObjectIsSame",
		"TestE2E_ObjectIsSame_DifferentObjects",
		"TestE2E_ObjectGetClass",
		"TestE2E_HandleRelease",
		"TestE2E_ExceptionCheck",
		"TestE2E_ErrorCases",
		"TestE2E_FullWorkflow_CallAnyJavaMethod",
		"TestE2E_GetSuperclass",
		"TestE2E_IsAssignableFrom",
		"TestE2E_IsInstanceOf",
		"TestE2E_GetMethodID",
		"TestE2E_CallMethod_StringLength",
		"TestE2E_GetFieldID",
		"TestE2E_GetStaticFieldID",
	}
	t.Logf("E2E test suite: %d tests", len(tests))
	for _, name := range tests {
		t.Logf("  %s", name)
	}

	// Verify we're running against a real server.
	addr := os.Getenv("JNICTL_E2E_ADDR")
	bin := os.Getenv("JNICTL_BIN")
	t.Logf("Server: %s", addr)
	t.Logf("Binary: %s", bin)

	// Quick smoke test: get version to confirm connectivity.
	out := runLiveJnicli(t, "jni", "get-version")
	t.Logf("Smoke test (get-version): %s", strings.TrimSpace(out))
	fmt.Fprintf(os.Stderr, "E2E: connected to %s, %d tests available\n", addr, len(tests))
}
