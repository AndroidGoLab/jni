#!/bin/bash
# Build, install, and test all example APKs on the emulator.
# Usage: ./test_all_apks.sh [example_name ...]
# If no arguments, tests all examples.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/.."

export ANDROID_SDK="${ANDROID_SDK:-$HOME/android-sdk}"
ADB="$ANDROID_SDK/platform-tools/adb"
RESULTS_DIR="$SCRIPT_DIR/test_results"
SCREENSHOTS_DIR="$RESULTS_DIR/screenshots"
mkdir -p "$RESULTS_DIR" "$SCREENSHOTS_DIR"

# Wait for device
"$ADB" wait-for-device

# Set up mock location provider on emulator (requires root)
"$ADB" root >/dev/null 2>&1 && sleep 2
"$ADB" shell appops set com.android.shell android:mock_location allow 2>/dev/null || true
"$ADB" shell cmd location providers add-test-provider gps \
    --requiresNetwork false --requiresSatellite false --requiresCell false \
    --hasMonetaryCost false --supportsAltitude true --supportsSpeed true \
    --supportsBearing true --powerRequirement 1 --accuracy 1 2>/dev/null || true
"$ADB" shell cmd location providers set-test-provider-enabled gps true 2>/dev/null || true
"$ADB" shell cmd location providers set-test-provider-location gps \
    --location "37.421998,-122.084000" 2>/dev/null || true
sleep 1

PASS=0
FAIL=0
SKIP=0
ERRORS=""

# Get list of examples to test
if [ $# -gt 0 ]; then
    EXAMPLES=("$@")
else
    EXAMPLES=()
    for d in "$SCRIPT_DIR"/*/; do
        name=$(basename "$d")
        [ "$name" = "apk" ] && continue
        [ "$name" = "test_results" ] && continue
        [ -f "$d/Makefile" ] || continue
        EXAMPLES+=("$name")
    done
fi

TOTAL=${#EXAMPLES[@]}
echo "=== Testing $TOTAL APKs ==="
echo ""

for i in "${!EXAMPLES[@]}"; do
    name="${EXAMPLES[$i]}"
    n=$((i + 1))
    echo "[$n/$TOTAL] $name"

    dir="$SCRIPT_DIR/$name"
    pkg="com.github.xaionaro_go.jni.examples.$name"
    result_file="$RESULTS_DIR/$name.log"

    # Build
    echo "  Building..."
    if ! (cd "$dir" && ANDROID_SDK="$ANDROID_SDK" make clean build) > "$result_file" 2>&1; then
        echo "  BUILD FAILED"
        FAIL=$((FAIL + 1))
        ERRORS="$ERRORS\n  BUILD_FAIL: $name"
        cat "$result_file" | tail -5 | sed 's/^/    /'
        continue
    fi

    # Uninstall old version (ignore errors)
    "$ADB" uninstall "$pkg" > /dev/null 2>&1 || true

    # Install
    echo "  Installing..."
    if ! "$ADB" install -r "$dir/build/$name.apk" >> "$result_file" 2>&1; then
        echo "  INSTALL FAILED"
        FAIL=$((FAIL + 1))
        ERRORS="$ERRORS\n  INSTALL_FAIL: $name"
        continue
    fi

    # Grant permissions (read from Makefile)
    perms=$(grep 'EXAMPLE_PERMISSIONS' "$dir/Makefile" 2>/dev/null | sed 's/.*:= *//' || true)
    for perm in $perms; do
        "$ADB" shell pm grant "$pkg" "$perm" 2>/dev/null || true
    done

    # Clear logcat
    "$ADB" logcat -c 2>/dev/null || true

    # Launch
    echo "  Running..."
    "$ADB" shell am start -n "$pkg/com.github.xaionaro_go.jni.example.ExampleActivity" >> "$result_file" 2>&1

    # Wait for output (the app writes to logcat with tag GoJNI)
    # Poll for up to 30 seconds
    got_output=false
    for attempt in $(seq 1 30); do
        sleep 1
        output=$("$ADB" logcat -d -s GoJNI 2>/dev/null | grep -v "^-" || true)
        if [ -n "$output" ]; then
            got_output=true
            break
        fi
    done

    # Capture screenshot
    "$ADB" exec-out screencap -p > "$SCREENSHOTS_DIR/$name.png" 2>/dev/null || true

    if ! $got_output; then
        # Check for crashes
        crash=$("$ADB" logcat -d | grep -iE "FATAL|AndroidRuntime|signal.*SIGSEGV|panic" | head -5 || true)
        if [ -n "$crash" ]; then
            echo "  CRASHED"
            echo "$crash" >> "$result_file"
            echo "$crash" | sed 's/^/    /'
            FAIL=$((FAIL + 1))
            ERRORS="$ERRORS\n  CRASH: $name"
        else
            echo "  TIMEOUT (no output in 30s)"
            FAIL=$((FAIL + 1))
            ERRORS="$ERRORS\n  TIMEOUT: $name"
        fi
    else
        # Check for ERROR in output
        echo "$output" >> "$result_file"
        if echo "$output" | grep -q "ERROR:"; then
            echo "  ERROR in output:"
            echo "$output" | grep "ERROR:" | sed 's/^/    /'
            FAIL=$((FAIL + 1))
            ERRORS="$ERRORS\n  ERROR: $name"
        else
            echo "  PASS"
            PASS=$((PASS + 1))
        fi
    fi

    # Force stop the app
    "$ADB" shell am force-stop "$pkg" 2>/dev/null || true

    # Uninstall to keep device clean
    "$ADB" uninstall "$pkg" > /dev/null 2>&1 || true
done

echo ""
echo "=== Results ==="
echo "  PASS: $PASS"
echo "  FAIL: $FAIL"
echo "  TOTAL: $TOTAL"
if [ -n "$ERRORS" ]; then
    echo ""
    echo "Failures:"
    echo -e "$ERRORS"
fi

# Exit with failure if any tests failed
[ $FAIL -eq 0 ]
