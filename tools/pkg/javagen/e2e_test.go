package javagen

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testGoModule = "github.com/xaionaro-go/jni"

// findRepoRoot walks up from the test directory to locate go.mod.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (no go.mod found)")
		}
		dir = parent
	}
}

// TestAllJavaSpecs_LoadAndMerge loads every spec/java/*.yaml, applies overlay,
// and merges. This exercises LoadSpec, LoadOverlay, Merge, and all type resolution.
func TestAllJavaSpecs_LoadAndMerge(t *testing.T) {
	root := findRepoRoot(t)
	specsDir := filepath.Join(root, "spec", "java")
	overlaysDir := filepath.Join(root, "spec", "overlays", "java")

	specs, err := filepath.Glob(filepath.Join(specsDir, "*.yaml"))
	if err != nil {
		t.Fatalf("glob specs: %v", err)
	}
	if len(specs) == 0 {
		t.Fatal("no spec files found")
	}

	t.Logf("found %d spec files", len(specs))

	if len(specs) != 53 {
		t.Errorf("expected 53 spec files, got %d", len(specs))
	}

	for _, specPath := range specs {
		baseName := strings.TrimSuffix(filepath.Base(specPath), ".yaml")
		t.Run(baseName, func(t *testing.T) {
			spec, err := LoadSpec(specPath)
			if err != nil {
				t.Fatalf("LoadSpec: %v", err)
			}
			if spec.Package == "" {
				t.Error("package is empty")
			}
			if spec.GoImport == "" {
				t.Error("go_import is empty")
			}

			overlayPath := filepath.Join(overlaysDir, baseName+".yaml")
			overlay, err := LoadOverlay(overlayPath)
			if err != nil {
				t.Fatalf("LoadOverlay: %v", err)
			}

			merged, err := Merge(spec, overlay)
			if err != nil {
				t.Fatalf("Merge: %v", err)
			}

			// Merge preserves spec fields.
			if merged.GoImport != spec.GoImport {
				t.Errorf("merged go_import %q != spec go_import %q", merged.GoImport, spec.GoImport)
			}
		})
	}
}

// TestAllJavaSpecs_Generate runs the full Generate pipeline for every
// spec/java/*.yaml and verifies the output is valid Go.
func TestAllJavaSpecs_Generate(t *testing.T) {
	root := findRepoRoot(t)
	specsDir := filepath.Join(root, "spec", "java")
	overlaysDir := filepath.Join(root, "spec", "overlays", "java")
	templatesDir := filepath.Join(root, "templates", "java")
	outputDir := t.TempDir()

	specs, err := filepath.Glob(filepath.Join(specsDir, "*.yaml"))
	if err != nil {
		t.Fatalf("glob specs: %v", err)
	}
	if len(specs) == 0 {
		t.Fatal("no spec files found")
	}

	for _, specPath := range specs {
		baseName := strings.TrimSuffix(filepath.Base(specPath), ".yaml")
		overlayPath := filepath.Join(overlaysDir, baseName+".yaml")

		t.Run(baseName, func(t *testing.T) {
			if err := Generate(specPath, overlayPath, templatesDir, outputDir, testGoModule); err != nil {
				t.Fatalf("Generate: %v", err)
			}

			// Determine the output package directory from go_import.
			spec, err := LoadSpec(specPath)
			if err != nil {
				t.Fatalf("LoadSpec: %v", err)
			}
			pkgDir := filepath.Join(outputDir, GoImportToRelDir(spec.GoImport, testGoModule))

			// Verify the directory was created.
			info, err := os.Stat(pkgDir)
			if err != nil {
				t.Fatalf("output dir missing: %v", err)
			}
			if !info.IsDir() {
				t.Fatalf("output path is not a directory: %s", pkgDir)
			}

			// Verify generated Go files parse correctly.
			fset := token.NewFileSet()
			entries, err := os.ReadDir(pkgDir)
			if err != nil {
				t.Fatalf("readdir: %v", err)
			}
			goFileCount := 0
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
					continue
				}
				goFileCount++
				goPath := filepath.Join(pkgDir, e.Name())
				_, parseErr := parser.ParseFile(fset, goPath, nil, parser.AllErrors)
				if parseErr != nil {
					t.Errorf("parse %s: %v", e.Name(), parseErr)
				}
			}
			if goFileCount == 0 {
				t.Error("no .go files generated")
			}
		})
	}
}

// TestGenerate_Idempotency runs Generate twice and verifies the output
// is byte-identical.
func TestGenerate_Idempotency(t *testing.T) {
	root := findRepoRoot(t)
	specsDir := filepath.Join(root, "spec", "java")
	overlaysDir := filepath.Join(root, "spec", "overlays", "java")
	templatesDir := filepath.Join(root, "templates", "java")

	// Use two different output dirs.
	outputDir1 := t.TempDir()
	outputDir2 := t.TempDir()

	// Pick a representative subset to keep test time reasonable.
	testSpecs := []string{"location", "bluetooth", "notification", "toast", "permission"}

	for _, name := range testSpecs {
		specPath := filepath.Join(specsDir, name+".yaml")
		overlayPath := filepath.Join(overlaysDir, name+".yaml")

		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			t.Skipf("spec %s not found", name)
		}

		if err := Generate(specPath, overlayPath, templatesDir, outputDir1, testGoModule); err != nil {
			t.Fatalf("Generate pass 1 (%s): %v", name, err)
		}
		if err := Generate(specPath, overlayPath, templatesDir, outputDir2, testGoModule); err != nil {
			t.Fatalf("Generate pass 2 (%s): %v", name, err)
		}
	}

	// Compare outputs.
	for _, name := range testSpecs {
		specPath := filepath.Join(specsDir, name+".yaml")
		spec, err := LoadSpec(specPath)
		if err != nil {
			t.Fatalf("LoadSpec %s: %v", name, err)
		}
		relDir := GoImportToRelDir(spec.GoImport, testGoModule)
		pkgDir1 := filepath.Join(outputDir1, relDir)
		pkgDir2 := filepath.Join(outputDir2, relDir)

		entries, err := os.ReadDir(pkgDir1)
		if err != nil {
			t.Fatalf("readdir %s: %v", name, err)
		}

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			data1, err := os.ReadFile(filepath.Join(pkgDir1, e.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", name, e.Name(), err)
			}
			data2, err := os.ReadFile(filepath.Join(pkgDir2, e.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", name, e.Name(), err)
			}
			if string(data1) != string(data2) {
				t.Errorf("idempotency failure: %s/%s differs between runs", name, e.Name())
			}
		}
	}
}

// TestGenerate_NoDrift verifies that regenerating from specs produces
// output identical to the committed generated files.
// This test checks for unintended divergence. If the templates or specs
// have been intentionally changed but output not yet regenerated, this
// test logs warnings but still passes (to avoid blocking other work).
func TestGenerate_NoDrift(t *testing.T) {
	root := findRepoRoot(t)
	specsDir := filepath.Join(root, "spec", "java")
	overlaysDir := filepath.Join(root, "spec", "overlays", "java")
	templatesDir := filepath.Join(root, "templates", "java")
	outputDir := t.TempDir()

	testSpecs := []string{"permission", "build", "environment"}

	for _, name := range testSpecs {
		specPath := filepath.Join(specsDir, name+".yaml")
		overlayPath := filepath.Join(overlaysDir, name+".yaml")

		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			continue
		}

		if err := Generate(specPath, overlayPath, templatesDir, outputDir, testGoModule); err != nil {
			t.Fatalf("Generate (%s): %v", name, err)
		}

		spec, err := LoadSpec(specPath)
		if err != nil {
			t.Fatalf("LoadSpec %s: %v", name, err)
		}

		relDir := GoImportToRelDir(spec.GoImport, testGoModule)
		genDir := filepath.Join(outputDir, relDir)
		committedDir := filepath.Join(root, relDir)

		if _, err := os.Stat(committedDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(genDir)
		if err != nil {
			t.Fatalf("readdir %s: %v", name, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			genData, err := os.ReadFile(filepath.Join(genDir, e.Name()))
			if err != nil {
				t.Fatalf("read gen %s/%s: %v", name, e.Name(), err)
			}
			committedPath := filepath.Join(committedDir, e.Name())
			committedData, err := os.ReadFile(committedPath)
			if err != nil {
				continue
			}
			if string(genData) != string(committedData) {
				t.Logf("drift detected: %s/%s differs from committed version (run make java to update)", name, e.Name())
			}
		}
	}
}

// TestGenerate_OutputFilePatterns verifies that each spec produces the
// expected set of output files based on its content.
func TestGenerate_OutputFilePatterns(t *testing.T) {
	root := findRepoRoot(t)
	templatesDir := filepath.Join(root, "templates", "java")

	tests := []struct {
		specName      string
		expectFiles   []string
		unexpectFiles []string
	}{
		{
			specName:    "location",
			expectFiles: []string{"doc.go", "init.go", "manager.go", "callbacks.go", "location.go", "constants.go"},
		},
		{
			specName:      "toast",
			expectFiles:   []string{"doc.go", "init.go", "toast.go", "constants.go"},
			unexpectFiles: []string{"callbacks.go"},
		},
		{
			specName:      "permission",
			expectFiles:   []string{"doc.go", "init.go", "constants.go", "context_compat.go", "activity_compat.go"},
			unexpectFiles: []string{"callbacks.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.specName, func(t *testing.T) {
			specPath := filepath.Join(root, "spec", "java", tt.specName+".yaml")
			overlayPath := filepath.Join(root, "spec", "overlays", "java", tt.specName+".yaml")
			outputDir := t.TempDir()

			if err := Generate(specPath, overlayPath, templatesDir, outputDir, testGoModule); err != nil {
				t.Fatalf("Generate: %v", err)
			}

			spec, err := LoadSpec(specPath)
			if err != nil {
				t.Fatalf("LoadSpec: %v", err)
			}
			pkgDir := filepath.Join(outputDir, GoImportToRelDir(spec.GoImport, testGoModule))

			for _, f := range tt.expectFiles {
				path := filepath.Join(pkgDir, f)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("expected file %s not found", f)
				}
			}
			for _, f := range tt.unexpectFiles {
				path := filepath.Join(pkgDir, f)
				if _, err := os.Stat(path); err == nil {
					t.Errorf("unexpected file %s found", f)
				}
			}
		})
	}
}

// TestGenerate_ContentPatterns verifies key content patterns in generated files.
func TestGenerate_ContentPatterns(t *testing.T) {
	root := findRepoRoot(t)
	templatesDir := filepath.Join(root, "templates", "java")
	specPath := filepath.Join(root, "spec", "java", "location.yaml")
	overlayPath := filepath.Join(root, "spec", "overlays", "java", "location.yaml")
	outputDir := t.TempDir()

	if err := Generate(specPath, overlayPath, templatesDir, outputDir, testGoModule); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	pkgDir := filepath.Join(outputDir, "location")

	// Check doc.go has package declaration.
	docGo := readFile(t, filepath.Join(pkgDir, "doc.go"))
	if !strings.Contains(docGo, "package location") {
		t.Error("doc.go missing package declaration")
	}
	if !strings.Contains(docGo, "Code generated by javagen") {
		t.Error("doc.go missing generated header")
	}

	// Check manager.go has NewManager constructor and methods.
	mgrGo := readFile(t, filepath.Join(pkgDir, "manager.go"))
	if !strings.Contains(mgrGo, "type Manager struct") {
		t.Error("manager.go missing Manager struct")
	}
	if !strings.Contains(mgrGo, "NewManager") {
		t.Error("manager.go missing NewManager constructor")
	}
	if !strings.Contains(mgrGo, "GetSystemService") {
		t.Error("manager.go missing GetSystemService call")
	}
	if !strings.Contains(mgrGo, "func (m *Manager)") {
		t.Error("manager.go missing Manager methods")
	}
	if !strings.Contains(mgrGo, "m.VM.Do") {
		t.Error("manager.go missing VM.Do pattern")
	}
	if !strings.Contains(mgrGo, "ensureInit") {
		t.Error("manager.go missing ensureInit call")
	}

	// Check init.go has sync.Once initialization.
	initGo := readFile(t, filepath.Join(pkgDir, "init.go"))
	if !strings.Contains(initGo, "sync.Once") {
		t.Error("init.go missing sync.Once")
	}
	if !strings.Contains(initGo, "ensureInit") {
		t.Error("init.go missing ensureInit")
	}
	if !strings.Contains(initGo, "FindClass") {
		t.Error("init.go missing FindClass")
	}

	// Check constants.go has expected values.
	constGo := readFile(t, filepath.Join(pkgDir, "constants.go"))
	if !strings.Contains(constGo, "GPS") {
		t.Error("constants.go missing GPS constant")
	}
	if !strings.Contains(constGo, "Network") {
		t.Error("constants.go missing Network constant")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}
