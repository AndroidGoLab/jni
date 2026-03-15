package protogen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AndroidGoLab/jni/tools/pkg/javagen"
)

// TestGenerate_AllRealSpecs is an E2E integration test that loads all spec/java/*.yaml
// files and verifies protogen generates valid .proto files for each one.
func TestGenerate_AllRealSpecs(t *testing.T) {
	root := findRepoRoot(t)
	specsDir := filepath.Join(root, "spec", "java")
	overlaysDir := filepath.Join(root, "spec", "overlays", "java")
	outputDir := t.TempDir()
	goModule := "github.com/AndroidGoLab/jni"

	specFiles, err := filepath.Glob(filepath.Join(specsDir, "*.yaml"))
	if err != nil {
		t.Fatalf("glob specs: %v", err)
	}
	if len(specFiles) < 30 {
		t.Fatalf("expected at least 30 spec files, found %d", len(specFiles))
	}

	var failed []string
	for _, specPath := range specFiles {
		baseName := strings.TrimSuffix(filepath.Base(specPath), ".yaml")
		overlayPath := filepath.Join(overlaysDir, baseName+".yaml")

		// Load the spec to determine the actual package name, which may differ
		// from the YAML filename (e.g. connectivity.yaml has package "net").
		spec, err := javagen.LoadSpec(specPath)
		if err != nil {
			t.Errorf("%s: load spec: %v", baseName, err)
			failed = append(failed, baseName)
			continue
		}
		pkgName := spec.Package

		if err := Generate(specPath, overlayPath, outputDir, goModule); err != nil {
			t.Errorf("Generate %s: %v", baseName, err)
			failed = append(failed, baseName)
			continue
		}

		protoPath := filepath.Join(outputDir, pkgName, pkgName+".proto")
		data, err := os.ReadFile(protoPath)
		if err != nil {
			t.Errorf("%s (pkg=%s): proto file not created at %s: %v", baseName, pkgName, protoPath, err)
			failed = append(failed, baseName)
			continue
		}

		content := string(data)
		if !strings.Contains(content, `syntax = "proto3";`) {
			t.Errorf("%s: missing proto3 syntax declaration", baseName)
			failed = append(failed, baseName)
		}
	}

	t.Logf("processed %d spec files, %d failures", len(specFiles), len(failed))
	if len(failed) > 0 {
		t.Errorf("failed specs: %s", strings.Join(failed, ", "))
	}
}
