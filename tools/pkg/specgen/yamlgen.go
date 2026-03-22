package specgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	dirPerm  = 0o755
	filePerm = 0o644
)

// AndroidServiceName maps known manager class names to their Android
// Context.getSystemService() constant names. Populated at runtime by
// LoadServiceNames, which reflects on android.jar via the svcgen Java tool.
var AndroidServiceName map[string]string

// loadMappingsFromSpecs reads all existing YAML spec files from dir and
// builds a map from Java class name → PackageMapping. This replaces the
// manual knownMappings table: when a class was already generated into a
// spec, its mapping is preserved automatically on re-generation.
func loadMappingsFromSpecs(dir string) (map[string]PackageMapping, error) {
	result := make(map[string]PackageMapping)
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return result, err
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var spec SpecFile
		if err := yaml.Unmarshal(data, &spec); err != nil {
			continue
		}
		if spec.Package == "" || spec.GoImport == "" {
			continue
		}
		for _, cls := range spec.Classes {
			if cls.JavaClass == "" {
				continue
			}
			result[cls.JavaClass] = PackageMapping{
				Package:  spec.Package,
				GoImport: spec.GoImport,
			}
		}
	}
	return result, nil
}

// GenerateSpec generates a YAML spec from .class files in a directory
// by running javap on each class.
func GenerateSpec(
	classPath string,
	className string,
	pkgMapping PackageMapping,
	goModule string,
) (*SpecFile, error) {
	jc, err := RunJavap(classPath, className)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", className, err)
	}

	spec := &SpecFile{
		Package:  pkgMapping.Package,
		GoImport: pkgMapping.GoImport,
	}

	cls := classFromJavap(jc, pkgMapping.Package)
	spec.Classes = append(spec.Classes, cls)

	return spec, nil
}

// GenerateFromRefDir scans ref/ for .class files and generates one YAML
// spec per top-level class (inner classes are grouped with their parent).
// extraClassPath is appended to the javap -cp argument.
func GenerateFromRefDir(
	refDir string,
	extraClassPath string,
	outputDir string,
	goModule string,
) error {
	// Load service name mappings from android.jar via the svcgen Java tool,
	// so that classFromJavap can detect system-service classes.
	if extraClassPath != "" {
		svcNames, err := LoadServiceNames(extraClassPath)
		if err != nil {
			return fmt.Errorf("load service names: %w", err)
		}
		AndroidServiceName = svcNames
	}

	var classFiles []string
	err := filepath.Walk(refDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".class") {
			classFiles = append(classFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", refDir, err)
	}

	// Separate top-level classes from inner classes.
	type classEntry struct {
		className string
		filePath  string
	}
	topLevel := make(map[string]classEntry)       // parent class → entry
	innerClasses := make(map[string][]classEntry) // parent class → inner entries

	for _, cf := range classFiles {
		rel, _ := filepath.Rel(refDir, cf)
		className := strings.TrimSuffix(rel, ".class")
		className = strings.ReplaceAll(className, "/", ".")

		entry := classEntry{className: className, filePath: cf}

		if strings.Contains(filepath.Base(cf), "$") {
			// Inner class — group with parent.
			parent := className[:strings.LastIndex(className, "$")]
			innerClasses[parent] = append(innerClasses[parent], entry)
		} else {
			topLevel[className] = entry
		}
	}

	if err := os.MkdirAll(outputDir, dirPerm); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}

	// Load mappings from existing spec files so that previously
	// generated class→package assignments are preserved automatically.
	existingMappings, err := loadMappingsFromSpecs(outputDir)
	if err != nil {
		return fmt.Errorf("load existing specs: %w", err)
	}

	cp := refDir
	if extraClassPath != "" {
		cp = refDir + ":" + extraClassPath
	}

	// Accumulate specs per Go package so that multiple Java classes
	// mapping to the same package are merged instead of overwritten.
	specs := make(map[string]*SpecFile) // key: Go package name

	for parentName, entry := range topLevel {
		mapping := inferClassMapping(parentName, goModule, existingMappings)

		spec, ok := specs[mapping.Package]
		if !ok {
			spec = &SpecFile{
				Package:  mapping.Package,
				GoImport: mapping.GoImport,
			}
			specs[mapping.Package] = spec
		}

		// Parse the top-level class.
		jc, err := RunJavap(cp, entry.className)
		if err != nil {
			return fmt.Errorf("javap %s: %w", entry.className, err)
		}
		cls := classFromJavap(jc, mapping.Package)
		spec.Classes = append(spec.Classes, cls)
		addConstants(spec, jc)

		// Parse inner classes.
		for _, inner := range innerClasses[parentName] {
			ijc, err := RunJavap(cp, inner.className)
			if err != nil {
				return fmt.Errorf("javap %s: %w", inner.className, err)
			}
			icls := classFromJavap(ijc, mapping.Package)
			spec.Classes = append(spec.Classes, icls)
			addConstants(spec, ijc)
		}
	}

	for pkgName, spec := range specs {
		spec.Constants = deduplicateConstants(spec.Constants)
		outPath := filepath.Join(outputDir, pkgName+".yaml")
		if err := writeSpecFile(spec, outPath); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

func addConstants(spec *SpecFile, jc *JavapClass) {
	for _, c := range jc.Constants {
		spec.Constants = append(spec.Constants, SpecConstant{
			GoName: javaConstantToGoName(c.Name),
			Value:  formatConstantValue(c),
			GoType: javaTypeToGoType(c.JavaType),
		})
	}
}

// formatConstantValue returns a YAML-ready representation of the constant's
// value. If javap provided a ConstantValue attribute, that is used; otherwise
// a type-appropriate placeholder is returned.
func formatConstantValue(c JavapConstant) string {
	if c.Value == "" {
		return inferConstantDefault(c.JavaType)
	}
	switch c.JavaType {
	case "java.lang.String":
		return strconv.Quote(c.Value)
	case "long":
		// javap outputs long values with a trailing "l" suffix (e.g. "86400000l")
		// which is not valid Go syntax — strip it.
		return strings.TrimSuffix(c.Value, "l")
	case "float":
		// javap outputs float values with a trailing "f" suffix (e.g. "-4.0f", "NaNf")
		// which is not valid Go syntax — strip it.
		return strings.TrimSuffix(c.Value, "f")
	default:
		return c.Value
	}
}

func classFromJavap(jc *JavapClass, goPkg string) SpecClass {
	cls := SpecClass{
		JavaClass: jc.FullName,
		GoType:    inferGoType(jc.FullName, goPkg),
	}

	// Determine obtain type.
	if svcName, ok := AndroidServiceName[jc.FullName]; ok && svcName != "" {
		cls.Obtain = "system_service"
		cls.ServiceName = svcName
		cls.Close = true
	}

	// Count method names to detect overloads.
	nameCounts := make(map[string]int)
	for _, m := range jc.Methods {
		if hasUnsupportedParams(m) {
			continue
		}
		nameCounts[m.Name]++
	}

	// Track per-name occurrence index for disambiguation.
	nameIndex := make(map[string]int)

	for _, m := range jc.Methods {
		if hasUnsupportedParams(m) {
			continue
		}

		sm := specMethodFromJavap(m)

		// Disambiguate overloaded methods by appending parameter count.
		if nameCounts[m.Name] > 1 {
			idx := nameIndex[m.Name]
			nameIndex[m.Name] = idx + 1
			suffix := fmt.Sprintf("%d", len(m.Params))
			if idx > 0 {
				suffix = fmt.Sprintf("%d_%d", len(m.Params), idx)
			}
			sm.GoName = sm.GoName + suffix
		}

		switch {
		case m.IsStatic:
			cls.StaticMethods = append(cls.StaticMethods, sm)
		default:
			cls.Methods = append(cls.Methods, sm)
		}
	}

	return cls
}

func specMethodFromJavap(m JavapMethod) SpecMethod {
	sm := SpecMethod{
		JavaMethod: m.Name,
		GoName:     javaMethodToGoName(m.Name),
		Returns:    javaTypeToSpecType(m.ReturnType),
		Error:      true,
	}

	for i, p := range m.Params {
		sm.Params = append(sm.Params, SpecParam{
			JavaType: javaTypeToSpecType(p.JavaType),
			GoName:   fmt.Sprintf("arg%d", i),
		})
	}

	return sm
}

// hasUnsupportedParams checks if a method has parameter types that can't
// be represented in the YAML spec (byte buffers, handlers, complex generics).
func hasUnsupportedParams(m JavapMethod) bool {
	for _, p := range m.Params {
		switch {
		case strings.Contains(p.JavaType, "ByteBuffer"):
			return true
		case strings.Contains(p.JavaType, "Handler"):
			return true
		case strings.Contains(p.JavaType, "[]"):
			// Array params are fine for primitives but complex for objects.
		}
	}
	return false
}

// inferClassMapping derives the Go package name from a single Java class name.
// It first checks existing spec files (loaded into existingMappings) so that
// previously generated class→package assignments are preserved. Falls back to
// prefix-based heuristics for new classes.
func inferClassMapping(
	className string,
	goModule string,
	existingMappings map[string]PackageMapping,
) PackageMapping {
	if m, ok := existingMappings[className]; ok {
		return m
	}
	return inferPackageMapping(className, goModule)
}

func inferPackageMapping(className string, goModule string) PackageMapping {
	// Map known Android package prefixes to Go packages.
	mappings := []struct {
		prefix string
		pkg    string
		goPath string
	}{
		{"android.app.admin.", "admin", "app/admin"},
		{"android.app.blob.", "blob", "app/blob"},
		{"android.app.role.", "role", "app/role"},
		{"android.app.job.", "job", "app/job"},
		{"android.app.usage.", "usage", "app/usage"},
		{"android.app.", "app", "app"},
		{"android.content.", "content", "content"},
		{"android.hardware.camera2.", "camera", "hardware/camera"},
		{"android.hardware.lights.", "lights", "hardware/lights"},
		{"android.hardware.", "hardware", "hardware"},
		{"android.location.altitude.", "altitude", "location/altitude"},
		{"android.location.", "location", "location"},
		{"android.media.session.", "session", "media/session"},
		{"android.media.", "media", "media"},
		{"android.net.wifi.p2p.", "wifi_p2p", "net/wifi/p2p"},
		{"android.net.wifi.rtt.", "wifi_rtt", "net/wifi/rtt"},
		{"android.net.wifi.", "wifi", "net/wifi"},
		{"android.net.", "net", "net"},
		{"android.nfc.", "nfc", "nfc"},
		{"android.os.storage.", "storage", "os/storage"},
		{"android.os.", "os", "os"},
		{"android.provider.", "provider", "provider"},
		{"android.se.omapi.", "omapi", "se/omapi"},
		{"android.telecom.", "telecom", "telecom"},
		{"android.telephony.", "telephony", "telephony"},
		{"android.view.inputmethod.", "inputmethod", "view/inputmethod"},
		{"android.view.", "display", "view/display"},
	}

	for _, m := range mappings {
		if strings.HasPrefix(className, m.prefix) {
			return PackageMapping{
				JavaPrefix: m.prefix,
				Package:    m.pkg,
				GoImport:   goModule + "/" + m.goPath,
			}
		}
	}

	// Fallback: use last segment of the Java package.
	parts := strings.Split(className, ".")
	pkg := parts[len(parts)-2]
	return PackageMapping{
		JavaPrefix: strings.Join(parts[:len(parts)-1], ".") + ".",
		Package:    pkg,
		GoImport:   goModule + "/" + pkg,
	}
}

// deduplicateConstants removes duplicate constants (by GoName) that
// arise when multiple Java classes in the same Go package export
// identically-named constants (e.g. CREATOR on Parcelable classes).
func deduplicateConstants(constants []SpecConstant) []SpecConstant {
	seen := make(map[string]struct{}, len(constants))
	result := make([]SpecConstant, 0, len(constants))
	for _, c := range constants {
		if _, ok := seen[c.GoName]; ok {
			continue
		}
		seen[c.GoName] = struct{}{}
		result = append(result, c)
	}
	return result
}

const generatedFileHeader = "# Code generated by specgen. DO NOT EDIT.\n" +
	"# To change this file, modify the generator at tools/cmd/specgen/\n" +
	"# or the ref/ class files, then run: make specs\n\n"

func writeSpecFile(spec *SpecFile, path string) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	var out []byte
	out = append(out, generatedFileHeader...)
	out = append(out, data...)
	return os.WriteFile(path, out, filePerm)
}

// ---- Name conversion helpers ----

// javaTypeToSpecType converts a fully-qualified Java type to the
// short form used in YAML specs.
func javaTypeToSpecType(jt string) string {
	switch jt {
	case "void":
		return "void"
	case "boolean":
		return "boolean"
	case "byte":
		return "byte"
	case "char":
		return "char"
	case "short":
		return "short"
	case "int":
		return "int"
	case "long":
		return "long"
	case "float":
		return "float"
	case "double":
		return "double"
	case "java.lang.String":
		return "String"
	case "java.lang.CharSequence":
		return "java.lang.CharSequence"
	case "byte[]":
		return "[B"
	case "int[]":
		return "[I"
	case "long[]":
		return "[J"
	default:
		return jt
	}
}

// javaTypeToGoType converts a Java type name to a Go type for constants.
func javaTypeToGoType(jt string) string {
	switch jt {
	case "int":
		return "int"
	case "long":
		return "int64"
	case "java.lang.String":
		return "string"
	case "boolean":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return "int"
	}
}

// javaMethodToGoName converts a Java method name (camelCase) to a Go
// exported name (PascalCase), with raw suffix for complex methods.
func javaMethodToGoName(name string) string {
	if len(name) == 0 {
		return name
	}
	goName := strings.ToUpper(name[:1]) + name[1:]

	// Append "Raw" suffix if the name starts with common patterns
	// indicating it returns a raw JNI object (convention in this project).
	return goName
}

// inferGoType determines the exported Go type name for a Java class.
// It strips the Go package name prefix when redundant (e.g.,
// "AlarmManager" in package "alarm" becomes "Manager").
func inferGoType(fullClass string, goPkg string) string {
	parts := strings.Split(fullClass, ".")
	name := parts[len(parts)-1]

	// Handle inner classes: Foo$Bar → FooBar (include parent for uniqueness).
	if idx := strings.LastIndex(name, "$"); idx >= 0 {
		parent := name[:idx]
		child := name[idx+1:]
		name = parent + child
	}

	// Strip Go package name prefix when redundant (e.g.,
	// "AlarmManager" in package "alarm" → "Manager").
	if len(goPkg) > 0 {
		prefix := strings.ToUpper(goPkg[:1]) + goPkg[1:]
		if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
			name = name[len(prefix):]
		}
	}

	return name
}

// javaConstantToGoName converts SCREAMING_SNAKE_CASE to PascalCase.
func javaConstantToGoName(name string) string {
	parts := strings.Split(strings.ToLower(name), "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// inferConstantDefault returns a placeholder default for a constant.
// The actual values come from the Android SDK; we use 0/"" as placeholders.
func inferConstantDefault(javaType string) string {
	switch javaType {
	case "java.lang.String":
		return `""`
	default:
		return "0"
	}
}
