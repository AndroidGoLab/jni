package specgen

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// LoadServiceNames runs the Java svcgen program against the given android.jar
// classpath and returns a map of fully-qualified Java class name to service
// name string (e.g. "android.app.AlarmManager" -> "alarm").
func LoadServiceNames(classPath string) (map[string]string, error) {
	javaSource, err := findSvcgenSource()
	if err != nil {
		return nil, fmt.Errorf("find svcgen source: %w", err)
	}

	// Compile the Java source. javac writes the .class file next to the source.
	sourceDir := filepath.Dir(javaSource)
	compileCmd := exec.Command("javac", "-cp", classPath, javaSource)
	compileCmd.Dir = sourceDir
	if out, err := compileCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("compile svcgen: %s: %w", out, err)
	}

	// Run the compiled class with android.jar on the classpath.
	cp := classPath + string(os.PathListSeparator) + sourceDir
	runCmd := exec.Command("java", "-cp", cp, "Main")
	runCmd.Dir = sourceDir
	out, err := runCmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		return nil, fmt.Errorf("run svcgen: %s: %w", stderr, err)
	}

	var result map[string]string
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse svcgen output: %w", err)
	}
	return result, nil
}

// findSvcgenSource locates tools/cmd/svcgen/Main.java relative to
// this Go source file's location in the repository.
func findSvcgenSource() (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}

	// This file is at tools/pkg/specgen/service_name_loader.go.
	// The Java source is at tools/cmd/svcgen/Main.java.
	toolsDir := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))
	javaSource := filepath.Join(toolsDir, "cmd", "svcgen", "Main.java")
	if _, err := os.Stat(javaSource); err != nil {
		return "", fmt.Errorf("svcgen source not found at %s: %w", javaSource, err)
	}
	return javaSource, nil
}
