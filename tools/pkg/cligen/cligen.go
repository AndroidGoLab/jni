// Package cligen generates cobra CLI commands from Java API YAML specs.
// It produces Go source files for cmd/jnictl that call proto-generated
// gRPC stubs directly, covering the full Android API surface.
package cligen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xaionaro-go/jni/tools/pkg/javagen"
	"github.com/xaionaro-go/jni/tools/pkg/protogen"
	"github.com/xaionaro-go/jni/tools/pkg/protoscan"
)

// Generate loads a Java API spec and overlay, builds proto data, converts
// it to CLI data structures, and writes a cobra command file.
// protoDir is the base directory containing compiled proto Go stubs.
func Generate(
	specPath string,
	overlayPath string,
	outputDir string,
	goModule string,
	protoDir string,
) error {
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

	protoData := protogen.BuildProtoData(merged, goModule)
	if len(protoData.Services) == 0 {
		return nil
	}

	// Resolve proto names to actual Go names by scanning compiled proto stubs.
	goNames := protoscan.Scan(filepath.Join(protoDir, merged.Package))

	// Build a combined map for name resolution (services + RPCs + messages).
	goClientNames := make(map[string]string)
	for k, v := range goNames.ServiceClients {
		goClientNames[k] = v
	}
	for k, v := range goNames.RPCMethods {
		goClientNames[k] = v
	}
	for k, v := range goNames.MessageTypes {
		goClientNames[k] = v
	}

	cliPkg := buildCLIPackage(protoData, goModule, goClientNames)
	if cliPkg == nil {
		return nil
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}

	outputPath := filepath.Join(outputDir, merged.Package+".go")
	if err := renderPackage(cliPkg, outputPath); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	return nil
}

// Removed: scanGoClientNames replaced by protoscan.Scan
