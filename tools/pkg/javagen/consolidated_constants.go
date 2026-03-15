package javagen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PackageConstants holds all constants from a single spec package.
type PackageConstants struct {
	// PackageName is the short package name (e.g., "app", "location").
	PackageName string
	Constants   []Constant
}

// GenerateConsolidatedConstants collects constants from all spec files
// and writes a single CGO-free Go file at <outputDir>/android/constants.go.
//
// Each constant is prefixed with the TitleCased package name to avoid
// collisions (e.g., app.LocationService becomes AppLocationService,
// location.GpsProvider becomes LocationGpsProvider).
func GenerateConsolidatedConstants(
	specPaths []string,
	overlayDir string,
	outputDir string,
) error {
	var packages []PackageConstants
	for _, specPath := range specPaths {
		spec, err := LoadSpec(specPath)
		if err != nil {
			return fmt.Errorf("load spec %s: %w", specPath, err)
		}
		if len(spec.Constants) == 0 {
			continue
		}
		packages = append(packages, PackageConstants{
			PackageName: spec.Package,
			Constants:   spec.Constants,
		})
	}

	// Sort packages alphabetically for deterministic output.
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].PackageName < packages[j].PackageName
	})

	src := buildConsolidatedSource(packages)

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("gofmt consolidated constants: %w", err)
	}

	dir := filepath.Join(outputDir, "android")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	return os.WriteFile(filepath.Join(dir, "constants.go"), formatted, 0o644)
}

// buildConsolidatedSource renders all package constants into a single
// Go source string.
func buildConsolidatedSource(packages []PackageConstants) string {
	var b strings.Builder
	b.WriteString(generatedHeader)
	b.WriteString("// Package android provides Android platform constants that can be\n")
	b.WriteString("// imported without CGO dependencies.\n")
	b.WriteString("package android\n\n")

	for _, pkg := range packages {
		prefix := packagePrefix(pkg.PackageName)

		// Group constants by Go type, preserving YAML order within each group.
		groups := mergeConstants(pkg.Constants)
		for _, grp := range groups {
			if grp.GoType != "" {
				// Prefix the named type to avoid collisions.
				prefixedType := prefix + grp.GoType
				fmt.Fprintf(&b, "type %s %s\n\n", prefixedType, grp.BaseType)
			}

			b.WriteString("const (\n")
			for _, c := range grp.Values {
				if grp.GoType != "" {
					fmt.Fprintf(&b, "\t%s%s %s%s = %s\n", prefix, c.GoName, prefix, grp.GoType, c.Value)
				} else {
					fmt.Fprintf(&b, "\t%s%s = %s\n", prefix, c.GoName, c.Value)
				}
			}
			b.WriteString(")\n\n")
		}
	}

	return b.String()
}

// packagePrefix converts a spec package name (e.g., "wifi_p2p") to a
// TitleCase prefix suitable for Go identifiers (e.g., "WifiP2p").
func packagePrefix(pkg string) string {
	parts := strings.Split(pkg, "_")
	var b strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}
