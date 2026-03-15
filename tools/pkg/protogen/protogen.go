package protogen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndroidGoLab/jni/tools/pkg/javagen"
)

// Generate loads a Java API spec and overlay, merges them, builds proto
// data structures, and writes a .proto file to the output directory.
func Generate(specPath, overlayPath, outputDir, goModule string) error {
	spec, err := javagen.LoadSpec(specPath)
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}

	overlay, err := javagen.LoadOverlay(overlayPath)
	if err != nil {
		return fmt.Errorf("load overlay: %w", err)
	}

	merged, err := javagen.Merge(spec, overlay)
	if err != nil {
		return fmt.Errorf("merge: %w", err)
	}

	data := BuildProtoData(merged, goModule)

	pkgDir := filepath.Join(outputDir, merged.Package)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", pkgDir, err)
	}

	outputPath := filepath.Join(pkgDir, merged.Package+".proto")
	if err := renderProto(data, outputPath); err != nil {
		return fmt.Errorf("render proto: %w", err)
	}

	return nil
}
