#!/usr/bin/env bash
#
# transform_examples.sh – Migrates all example main.go files from the
# old c-shared/goRun/goGetOutput pattern to the new NativeActivity pattern.
#
# Three categories of files are handled:
#   A) Files with func run(vm *jni.VM) error  (42 files)
#   B) Files with func run() (no args)        (1 file: pdf)
#   C) Files with no run function at all       (11 files: logic in goRun body)
#
# Usage:  bash tools/transform_examples.sh
# The script does NOT format the files; run goimports or gofmt afterward.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EXAMPLES_DIR="$REPO_ROOT/examples"

# awk snippet to remove a function by name (given as "export_name").
# Removes the "//export <name>" comment and the entire function body
# by tracking brace depth. Prints all other lines.
remove_exported_func() {
    local file="$1"
    local export_name="$2"
    awk -v ename="$export_name" '
    $0 == "//export " ename { skip=1; next }
    skip && /^func / {
        # Start of the function line – count its braces
        in_func=1; brace=0
        for (i=1; i<=length($0); i++) {
            c = substr($0, i, 1)
            if (c == "{") brace++
            else if (c == "}") brace--
        }
        if (brace <= 0) { skip=0; in_func=0 }
        next
    }
    in_func {
        for (i=1; i<=length($0); i++) {
            c = substr($0, i, 1)
            if (c == "{") brace++
            else if (c == "}") brace--
        }
        if (brace <= 0) { skip=0; in_func=0 }
        next
    }
    skip { next }
    { print }
    ' "$file" > "$file.tmp" && mv "$file.tmp" "$file"
}

count=0

for mainfile in "$EXAMPLES_DIR"/*/main.go; do
    dir="$(basename "$(dirname "$mainfile")")"

    # ---------------------
    # Detect category
    # ---------------------
    has_run_vm=0
    has_run_noargs=0
    grep -q 'func run(vm \*jni\.VM) error' "$mainfile" && has_run_vm=1 || true
    grep -q 'func run()' "$mainfile" && has_run_noargs=1 || true

    # ========================================================================
    # STEP 1: Replace cgo preamble
    # ========================================================================
    # Old:  /*\n#include <jni.h>\n*/
    # New:  /*\n#include <android/native_activity.h>\nextern ...\nstatic ...\n*/
    sed -i '
/^\/\*$/{
    N
    /\n#include <jni\.h>$/{
        N
        /\n\*\/$/c\
/*\
#include <android/native_activity.h>\
extern void goOnResume(ANativeActivity*);\
static void _onResume(ANativeActivity* a) { goOnResume(a); }\
*/
    }
}
' "$mainfile"

    # ========================================================================
    # STEP 2: Add new imports (capi, exampleui) and ensure unsafe + jni present
    # ========================================================================

    # Add "unsafe" if not already imported
    if ! grep -q '"unsafe"' "$mainfile"; then
        sed -i '/^import ($/a\\t"unsafe"' "$mainfile"
    fi

    # Add "github.com/AndroidGoLab/jni" if not already imported.
    if ! grep -q '"github\.com/AndroidGoLab/jni"$' "$mainfile"; then
        awk '
        /^import \($/,/^\)$/ {
            if ($0 == ")") {
                print "\t\"github.com/AndroidGoLab/jni\""
            }
        }
        { print }
        ' "$mainfile" > "$mainfile.tmp" && mv "$mainfile.tmp" "$mainfile"
    fi

    # Add "github.com/AndroidGoLab/jni/capi" right after the jni import
    if ! grep -q '"github.com/AndroidGoLab/jni/capi"' "$mainfile"; then
        sed -i '/"github\.com\/AndroidGoLab\/jni"$/a\\t"github.com/AndroidGoLab/jni/capi"' "$mainfile"
    fi

    # Add "github.com/AndroidGoLab/jni/examples/common/ui" right after the capi import
    if ! grep -q '"github.com/AndroidGoLab/jni/examples/common/ui"' "$mainfile"; then
        sed -i '/"github\.com\/AndroidGoLab\/jni\/capi"$/a\\t"github.com/AndroidGoLab/jni/examples/common/ui"' "$mainfile"
    fi

    # ========================================================================
    # STEP 3: Remove "var output bytes.Buffer"
    # ========================================================================
    sed -i '/^var output bytes\.Buffer$/d' "$mainfile"

    # ========================================================================
    # STEP 4: Handle each category differently for goRun/goGetOutput/run
    # ========================================================================

    if [ "$has_run_vm" -eq 1 ]; then
        # ---- Category A: has func run(vm *jni.VM) error ----

        # Remove goRun and goGetOutput exported functions
        remove_exported_func "$mainfile" "goRun"
        remove_exported_func "$mainfile" "goGetOutput"

        # Change run signature to accept output parameter
        sed -i 's/^func run(vm \*jni\.VM) error {$/func run(vm *jni.VM, output *bytes.Buffer) error {/' "$mainfile"

        # Replace &output with output everywhere in the file
        sed -i 's/\&output/output/g' "$mainfile"

    elif [ "$has_run_noargs" -eq 1 ]; then
        # ---- Category B: has func run() (e.g. pdf) ----

        # Remove goRun and goGetOutput exported functions
        remove_exported_func "$mainfile" "goRun"
        remove_exported_func "$mainfile" "goGetOutput"

        # Change run() to run(vm *jni.VM, output *bytes.Buffer) error
        sed -i 's/^func run() {$/func run(vm *jni.VM, output *bytes.Buffer) error {/' "$mainfile"

        # Add "return nil" before the closing } of the run function
        awk '
        /^func run\(vm \*jni\.VM, output \*bytes\.Buffer\) error \{$/ {
            in_run=1; brace=1  # opening { is on this line
            print
            next
        }
        in_run {
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace++
                else if (c == "}") brace--
            }
            if (brace <= 0) {
                # This is the closing brace of run
                print "\treturn nil"
                print
                in_run=0
                next
            }
            print
            next
        }
        { print }
        ' "$mainfile" > "$mainfile.tmp" && mv "$mainfile.tmp" "$mainfile"

        # Replace &output with output
        sed -i 's/\&output/output/g' "$mainfile"

    else
        # ---- Category C: no run function, logic in goRun body ----

        # Two-pass approach:
        # Pass 1: Extract goRun body lines into a temp file
        awk '
        /^\/\/export goRun$/ { skip_export=1; next }
        skip_export && /^func goRun\(/ {
            in_gorun=1; brace=0
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace++
                else if (c == "}") brace--
            }
            next
        }
        in_gorun {
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace++
                else if (c == "}") brace--
            }
            if (brace <= 0) {
                in_gorun=0; skip_export=0
                next
            }
            print
            next
        }
        ' "$mainfile" > "$mainfile.gorun_body"

        # Pass 2: Remove goRun + goGetOutput; append new run() at end of file
        awk -v bodyfile="$mainfile.gorun_body" '
        BEGIN {
            body = ""
            while ((getline line < bodyfile) > 0) {
                if (body != "") body = body "\n"
                body = body line
            }
            close(bodyfile)
        }
        /^\/\/export goRun$/ { skip_export=1; next }
        skip_export && /^func goRun\(/ {
            in_gorun=1; brace=0
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace++
                else if (c == "}") brace--
            }
            if (brace <= 0) { in_gorun=0; skip_export=0 }
            next
        }
        in_gorun {
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace++
                else if (c == "}") brace--
            }
            if (brace <= 0) { in_gorun=0; skip_export=0 }
            next
        }
        skip_export { next }

        /^\/\/export goGetOutput$/ { skip_get=1; next }
        skip_get && /^func goGetOutput\(/ {
            in_getout=1; brace_g=0
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace_g++
                else if (c == "}") brace_g--
            }
            if (brace_g <= 0) { skip_get=0; in_getout=0 }
            next
        }
        in_getout {
            for (i=1; i<=length($0); i++) {
                c = substr($0, i, 1)
                if (c == "{") brace_g++
                else if (c == "}") brace_g--
            }
            if (brace_g <= 0) { skip_get=0; in_getout=0 }
            next
        }
        skip_get { next }

        { print }

        END {
            print ""
            print "func run(vm *jni.VM, output *bytes.Buffer) error {"
            if (body != "") print body
            print "\treturn nil"
            print "}"
        }
        ' "$mainfile" > "$mainfile.tmp" && mv "$mainfile.tmp" "$mainfile"
        rm -f "$mainfile.gorun_body"

        # Replace &output with output
        sed -i 's/\&output/output/g' "$mainfile"
    fi

    # ========================================================================
    # STEP 5: Add init() and NativeActivity boilerplate after "func main() {}"
    # ========================================================================
    awk '
    { print }
    /^func main\(\) \{\}$/ {
        print ""
        print "func init() { ui.Register(run) }"
        print ""
        print "//export ANativeActivity_onCreate"
        print "func ANativeActivity_onCreate(activity *C.ANativeActivity, savedState unsafe.Pointer, savedStateSize C.size_t) {"
        print "\tui.OnCreate("
        print "\t\tjni.VMFromPtr(unsafe.Pointer(activity.vm)),"
        print "\t\tjni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),"
        print "\t)"
        print "\tactivity.callbacks.onResume = C._onResume"
        print "}"
        print ""
        print "//export goOnResume"
        print "func goOnResume(activity *C.ANativeActivity) {"
        print "\tui.OnResume("
        print "\t\tjni.ObjectFromRef(capi.Object(uintptr(unsafe.Pointer(activity.clazz)))),"
        print "\t)"
        print "}"
    }
    ' "$mainfile" > "$mainfile.tmp" && mv "$mainfile.tmp" "$mainfile"

    # ========================================================================
    # STEP 6: Clean up consecutive blank lines (collapse to max 1)
    # ========================================================================
    awk '
    /^$/ { blank++; if (blank <= 1) print; next }
    { blank=0; print }
    ' "$mainfile" > "$mainfile.tmp" && mv "$mainfile.tmp" "$mainfile"

    count=$((count + 1))
    echo "[$count] transformed: examples/$dir/main.go"
done

echo ""
echo "Done. Transformed $count files."
echo "Run 'goimports -w examples/*/main.go' to fix import ordering."
