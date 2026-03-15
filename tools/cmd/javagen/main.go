package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/xaionaro-go/jni/tools/pkg/javagen"
)

func main() {
	specsDir := flag.String("specs", "spec/java", "directory containing per-package YAML specs")
	overlaysDir := flag.String("overlays", "spec/overlays/java", "directory containing per-package overlay YAMLs")
	templatesDir := flag.String("templates", "templates/java", "directory containing Go templates")
	outputDir := flag.String("output", ".", "base output directory")
	goModule := flag.String("go-module", "github.com/xaionaro-go/jni", "Go module path (used to derive output directories from go_import)")
	flag.Parse()

	specs, err := filepath.Glob(filepath.Join(*specsDir, "*.yaml"))
	if err != nil {
		log.Fatalf("glob specs: %v", err)
	}
	if len(specs) == 0 {
		log.Fatalf("no spec files found in %s", *specsDir)
	}

	for _, specPath := range specs {
		baseName := strings.TrimSuffix(filepath.Base(specPath), ".yaml")
		overlayPath := filepath.Join(*overlaysDir, baseName+".yaml")

		if err := javagen.Generate(specPath, overlayPath, *templatesDir, *outputDir, *goModule); err != nil {
			log.Fatalf("generate %s: %v", baseName, err)
		}
	}

	// Generate consolidated android/constants.go with all constants
	// from every spec, prefixed by package name. This file is CGO-free
	// so non-CGO code (e.g., jnicli) can import it.
	if err := javagen.GenerateConsolidatedConstants(specs, *overlaysDir, *outputDir); err != nil {
		log.Fatalf("generate consolidated constants: %v", err)
	}
}
