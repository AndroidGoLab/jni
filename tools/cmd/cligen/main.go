package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/xaionaro-go/jni/tools/pkg/cligen"
)

func main() {
	specsDir := flag.String("specs", "spec/java", "directory containing per-package YAML specs")
	overlaysDir := flag.String("overlays", "spec/overlays/java", "directory containing overlays")
	outputDir := flag.String("output", "cmd/jnictl", "output directory for generated CLI files")
	goModule := flag.String("go-module", "github.com/xaionaro-go/jni", "Go module path")
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

		if err := cligen.Generate(specPath, overlayPath, *outputDir, *goModule); err != nil {
			log.Fatalf("generate %s: %v", baseName, err)
		}
	}
}
