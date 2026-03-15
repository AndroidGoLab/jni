#!/usr/bin/env bash
# Shell-based E2E test runner for jnicli against a running jniservice.
#
# Prerequisites:
#   - JNICTL_E2E_ADDR set to the server address (e.g. localhost:50051)
#   - JNICTL_BIN set to the path of the jnicli binary (or jnicli in PATH)
#   - jq installed for JSON parsing
#
# Usage:
#   JNICTL_E2E_ADDR=localhost:50051 JNICTL_BIN=./build/jnicli-host ./run_e2e.sh

set -euo pipefail

: "${JNICTL_E2E_ADDR:?JNICTL_E2E_ADDR must be set (e.g. localhost:50051)}"
JNICTL="${JNICTL_BIN:-jnicli}"
PASS=0
FAIL=0
TOTAL=0

jnicli() {
    "$JNICTL" --addr "$JNICTL_E2E_ADDR" --insecure "$@"
}

run_test() {
    local name="$1"
    shift
    TOTAL=$((TOTAL + 1))
    if "$@"; then
        PASS=$((PASS + 1))
        echo "PASS: $name"
    else
        FAIL=$((FAIL + 1))
        echo "FAIL: $name"
    fi
}

# ---- Test functions ----

test_get_version() {
    local out
    out=$(jnicli jni get-version)
    local version
    version=$(echo "$out" | jq -r '.version')
    [ "$version" -gt 0 ]
}

test_find_class_string() {
    local out
    out=$(jnicli jni class find --name java/lang/String)
    local handle
    handle=$(echo "$out" | jq -r '.classHandle')
    [ -n "$handle" ] && [ "$handle" != "null" ] && [ "$handle" != "0" ]
}

test_find_class_system() {
    local out
    out=$(jnicli jni class find --name java/lang/System)
    local handle
    handle=$(echo "$out" | jq -r '.classHandle')
    [ -n "$handle" ] && [ "$handle" != "null" ] && [ "$handle" != "0" ]
}

test_find_class_not_found() {
    # Should fail (non-zero exit).
    if jnicli jni class find --name does/not/Exist 2>/dev/null; then
        return 1
    fi
    return 0
}

test_get_static_method_id() {
    local class_out method_out
    class_out=$(jnicli jni class find --name java/lang/System)
    local class_handle
    class_handle=$(echo "$class_out" | jq -r '.classHandle')

    method_out=$(jnicli jni method get-static-id \
        --class "$class_handle" --name currentTimeMillis --sig '()J')
    local method_id
    method_id=$(echo "$method_out" | jq -r '.methodId')
    [ -n "$method_id" ] && [ "$method_id" != "null" ] && [ "$method_id" != "0" ]
}

test_call_current_time_millis() {
    local class_out method_out call_out
    class_out=$(jnicli jni class find --name java/lang/System)
    local class_handle
    class_handle=$(echo "$class_out" | jq -r '.classHandle')

    method_out=$(jnicli jni method get-static-id \
        --class "$class_handle" --name currentTimeMillis --sig '()J')
    local method_id
    method_id=$(echo "$method_out" | jq -r '.methodId')

    call_out=$(jnicli jni method call-static \
        --class "$class_handle" --method "$method_id" --return-type long)
    local time_val
    time_val=$(echo "$call_out" | jq -r '.result.j')
    [ -n "$time_val" ] && [ "$time_val" != "null" ] && [ "$time_val" -gt 0 ]
}

test_string_new_and_get() {
    local new_out get_out
    new_out=$(jnicli jni string new --value "hello world")
    local handle
    handle=$(echo "$new_out" | jq -r '.stringHandle')

    get_out=$(jnicli jni string get --handle "$handle")
    local value
    value=$(echo "$get_out" | jq -r '.value')
    [ "$value" = "hello world" ]
}

test_string_length() {
    local new_out length_out
    new_out=$(jnicli jni string new --value "hello")
    local handle
    handle=$(echo "$new_out" | jq -r '.stringHandle')

    length_out=$(jnicli jni string length --handle "$handle")
    local length
    length=$(echo "$length_out" | jq -r '.length')
    [ "$length" = "5" ]
}

test_object_is_same() {
    local out
    out=$(jnicli jni class find --name java/lang/String)
    local handle
    handle=$(echo "$out" | jq -r '.classHandle')

    local same_out
    same_out=$(jnicli jni object is-same --object1 "$handle" --object2 "$handle")
    local result
    result=$(echo "$same_out" | jq -r '.result')
    [ "$result" = "true" ]
}

test_object_get_class() {
    local new_out
    new_out=$(jnicli jni string new --value "test")
    local str_handle
    str_handle=$(echo "$new_out" | jq -r '.stringHandle')

    local class_out
    class_out=$(jnicli jni object get-class --object "$str_handle")
    local class_handle
    class_handle=$(echo "$class_out" | jq -r '.classHandle')
    [ -n "$class_handle" ] && [ "$class_handle" != "null" ] && [ "$class_handle" != "0" ]
}

test_handle_release() {
    local new_out
    new_out=$(jnicli jni string new --value "disposable")
    local handle
    handle=$(echo "$new_out" | jq -r '.stringHandle')

    # Release should succeed.
    jnicli handle release --handle "$handle" >/dev/null

    # Second release should fail (handle no longer exists).
    if jnicli handle release --handle "$handle" 2>/dev/null; then
        return 1
    fi
    return 0
}

test_exception_check() {
    local out
    out=$(jnicli jni exception check)
    # hasException should be false (protojson omits false, so missing = ok).
    local has_exc
    has_exc=$(echo "$out" | jq -r '.hasException // false')
    [ "$has_exc" = "false" ]
}

test_get_superclass() {
    local out
    out=$(jnicli jni class find --name java/lang/String)
    local class_handle
    class_handle=$(echo "$out" | jq -r '.classHandle')

    local super_out
    super_out=$(jnicli jni class get-superclass --class "$class_handle")
    local super_handle
    super_handle=$(echo "$super_out" | jq -r '.classHandle')
    [ -n "$super_handle" ] && [ "$super_handle" != "null" ] && [ "$super_handle" != "0" ]
}

test_is_assignable_from() {
    local out
    out=$(jnicli jni class find --name java/lang/String)
    local handle
    handle=$(echo "$out" | jq -r '.classHandle')

    local result_out
    result_out=$(jnicli jni class is-assignable-from --class1 "$handle" --class2 "$handle")
    local result
    result=$(echo "$result_out" | jq -r '.result')
    [ "$result" = "true" ]
}

# ---- Run all tests ----

echo "=== E2E Tests against $JNICTL_E2E_ADDR ==="
echo "Binary: $JNICTL"
echo

run_test "GetVersion"                test_get_version
run_test "FindClass_String"          test_find_class_string
run_test "FindClass_System"          test_find_class_system
run_test "FindClass_NotFound"        test_find_class_not_found
run_test "GetStaticMethodID"         test_get_static_method_id
run_test "CallCurrentTimeMillis"     test_call_current_time_millis
run_test "StringNewAndGet"           test_string_new_and_get
run_test "StringLength"              test_string_length
run_test "ObjectIsSame"              test_object_is_same
run_test "ObjectGetClass"            test_object_get_class
run_test "HandleRelease"             test_handle_release
run_test "ExceptionCheck"            test_exception_check
run_test "GetSuperclass"             test_get_superclass
run_test "IsAssignableFrom"          test_is_assignable_from

echo
echo "=== Results: $PASS passed, $FAIL failed, $TOTAL total ==="

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
