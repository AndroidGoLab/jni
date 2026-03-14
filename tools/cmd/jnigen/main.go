package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xaionaro-go/jni/tools/pkg/jnigen"
)

func main() {
	specPath := flag.String("spec", "spec/jni.yaml", "path to JNI spec YAML")
	overlayPath := flag.String("overlay", "spec/overlays/jni.yaml", "path to overlay YAML")
	templatesDir := flag.String("templates", "templates/jni/", "path to template directory")
	outputDir := flag.String("output", ".", "output root directory")
	flag.Parse()

	if err := run(*specPath, *overlayPath, *templatesDir, *outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "jnigen: %v\n", err)
		os.Exit(1)
	}
}

func run(specPath, overlayPath, templatesDir, outputDir string) error {
	spec, err := jnigen.LoadSpec(specPath)
	if err != nil {
		return fmt.Errorf("loading spec: %w", err)
	}

	overlay, err := jnigen.LoadOverlay(overlayPath)
	if err != nil {
		return fmt.Errorf("loading overlay: %w", err)
	}

	merged, err := jnigen.Merge(spec, overlay)
	if err != nil {
		return fmt.Errorf("merging spec and overlay: %w", err)
	}

	return jnigen.Render(merged, templatesDir, outputDir)
}
