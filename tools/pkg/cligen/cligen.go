// Package cligen generates cobra CLI commands from Java API YAML specs.
// It produces Go source files for cmd/jnictl that call proto-generated
// gRPC stubs directly, covering the full Android API surface.
package cligen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xaionaro-go/jni/tools/pkg/javagen"
	"github.com/xaionaro-go/jni/tools/pkg/protogen"
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

	// Resolve proto service names to actual Go client constructor names
	// by scanning the compiled _grpc.pb.go file.
	goClientNames := scanGoClientNames(filepath.Join(protoDir, merged.Package))

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

var newClientRe = regexp.MustCompile(`^func New(\w+Client)\(`)

// scanGoClientNames reads a _grpc.pb.go file and returns a map from
// proto service name to the actual Go client constructor suffix.
// E.g. if the file has "func NewP2PConfigServiceClient(", it maps
// "P2pConfigService" → "P2PConfigService".
func scanGoClientNames(protoPackageDir string) map[string]string {
	result := make(map[string]string)

	matches, _ := filepath.Glob(filepath.Join(protoPackageDir, "*_grpc.pb.go"))
	for _, path := range matches {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			m := newClientRe.FindStringSubmatch(scanner.Text())
			if m == nil {
				continue
			}
			// m[1] is e.g. "P2PConfigServiceClient"
			goName := strings.TrimSuffix(m[1], "Client")
			// Map from lowercase version to actual Go name.
			result[strings.ToLower(goName)] = goName
		}
		f.Close()
	}
	return result
}
