package grpcgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xaionaro-go/jni/tools/pkg/javagen"
	"github.com/xaionaro-go/jni/tools/pkg/protogen"
	"github.com/xaionaro-go/jni/tools/pkg/protoscan"
)

// GenerateServer loads a Java API spec and overlay, merges them, builds server
// data structures, and writes a gRPC server implementation file.
// It returns composite entries for each service generated (for use in the
// composite server registration file).
func GenerateServer(specPath, overlayPath, outputDir, goModule, protoDir string) ([]CompositeEntry, error) {
	spec, err := javagen.LoadSpec(specPath)
	if err != nil {
		return nil, fmt.Errorf("load spec: %w", err)
	}

	overlay, err := javagen.LoadOverlay(overlayPath)
	if err != nil {
		return nil, fmt.Errorf("load overlay: %w", err)
	}

	merged, err := javagen.Merge(spec, overlay)
	if err != nil {
		return nil, fmt.Errorf("merge: %w", err)
	}

	protoData := protogen.BuildProtoData(merged, goModule)
	goNames := protoscan.Scan(filepath.Join(protoDir, merged.Package))

	data := buildServerData(merged, goModule, protoData, goNames)
	if len(data.Services) == 0 {
		return nil, nil
	}

	pkgDir := filepath.Join(outputDir, "grpc", "server", merged.Package)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", pkgDir, err)
	}

	outputPath := filepath.Join(pkgDir, "server.go")
	if err := renderServer(data, outputPath); err != nil {
		return nil, fmt.Errorf("render server: %w", err)
	}

	var entries []CompositeEntry
	for _, svc := range data.Services {
		entries = append(entries, CompositeEntry{
			Package:      data.Package,
			GoType:       svc.GoType,
			ServiceName:  svc.ServiceName,
			NeedsHandles: svc.NeedsHandles,
		})
	}
	return entries, nil
}

// buildServerData converts a MergedSpec into server template data.
func buildServerData(
	merged *javagen.MergedSpec,
	goModule string,
	protoData *protogen.ProtoData,
	goNames protoscan.GoNames,
) *ServerData {
	data := &ServerData{
		Package:  merged.Package,
		GoModule: goModule,
		GoImport: merged.GoImport,
	}

	// Build maps from Java class name to data class info.
	// Only include exported data classes (those whose GoType starts with uppercase),
	// because the Extract function for unexported types is not accessible from
	// the grpcserver package.
	javaClassToDataClass := make(map[string]string)
	dataClassMap := make(map[string]bool)
	dcFieldMap := make(map[string][]javagen.MergedField)
	for _, dc := range merged.DataClasses {
		if !isExported(dc.GoType) {
			continue
		}
		javaClassToDataClass[dc.JavaClass] = dc.GoType
		dataClassMap[dc.GoType] = true
		dcFieldMap[dc.GoType] = dc.Fields
	}

	// Build data class info for result conversion.
	for _, dc := range merged.DataClasses {
		if !isExported(dc.GoType) {
			continue
		}
		sdc := ServerDataClass{GoType: dc.GoType}
		for _, f := range dc.Fields {
			sdc.Fields = append(sdc.Fields, ServerDataClassField{
				GoName:    f.GoName,
				ProtoName: f.GoName,
				GoType:    f.GoType,
			})
		}
		data.DataClasses = append(data.DataClasses, sdc)
	}

	// Build RPC name lookup from protogen data (handles collision renames).
	protoRPCNames := make(map[string]string)
	for _, ps := range protoData.Services {
		for _, rpc := range ps.RPCs {
			if rpc.OriginalName != "" {
				protoRPCNames[strings.ToLower(rpc.OriginalName)] = rpc.Name
			}
			protoRPCNames[strings.ToLower(rpc.Name)] = rpc.Name
		}
	}

	// Build services from classes that have a NewXxx(ctx *app.Context) constructor.
	for _, cls := range merged.Classes {
		if !hasContextConstructor(cls) {
			continue
		}
		if len(cls.Methods) == 0 {
			continue
		}

		rawServiceName := exportName(cls.GoType) + "Service"
		serviceName := goNames.ResolveService(rawServiceName)

		svc := ServerService{
			GoType:      cls.GoType,
			ServiceName: serviceName,
			Obtain:      cls.Obtain,
			Close:       cls.Close,
		}

		for _, m := range cls.Methods {
			if !isExported(m.GoName) {
				continue
			}
			sm := buildServerMethod(m, dataClassMap, javaClassToDataClass, dcFieldMap, protoRPCNames, goNames)
			if sm.ReturnKind == "data_class" || sm.ReturnKind == "object" {
				data.NeedsJNI = true
			}
			if sm.NeedsHandles {
				svc.NeedsHandles = true
				data.NeedsHandles = true
			}
			svc.Methods = append(svc.Methods, sm)
		}

		if len(svc.Methods) == 0 {
			continue
		}
		data.Services = append(data.Services, svc)
	}

	return data
}

// buildServerMethod converts a MergedMethod to a ServerMethod with
// pre-rendered template expressions.
func buildServerMethod(
	m javagen.MergedMethod,
	dataClassNames map[string]bool,
	javaClassToDataMsg map[string]string,
	dcFieldMap map[string][]javagen.MergedField,
	protoRPCNames map[string]string,
	goNames protoscan.GoNames,
) ServerMethod {
	rawName := exportName(m.GoName)
	goName := rawName
	if resolved, ok := protoRPCNames[strings.ToLower(rawName)]; ok {
		goName = resolved
	}
	goName = goNames.ResolveRPC(goName)
	reqType := goName + "Request"
	respType := goName + "Response"

	sm := ServerMethod{
		GoName:       goName,
		SpecGoName:   m.GoName,
		RequestType:  reqType,
		ResponseType: respType,
		HasError:     m.Error,
		HasResult:    m.ReturnKind != javagen.ReturnVoid,
		GoReturnType: m.GoReturn,
	}

	// Build call arguments expression.
	callArgs, argsNeedHandles := buildCallArgs(m.Params)
	sm.CallArgs = callArgs

	// Determine return kind and result expression.
	switch m.ReturnKind {
	case javagen.ReturnVoid:
		sm.ReturnKind = "void"
	case javagen.ReturnString:
		sm.ReturnKind = "string"
		sm.ResultExpr = "result"
	case javagen.ReturnBool:
		sm.ReturnKind = "bool"
		sm.ResultExpr = "result"
	case javagen.ReturnPrimitive:
		sm.ReturnKind = "primitive"
		sm.ResultExpr = convertPrimitiveExpr(m.GoReturn, "result")
	case javagen.ReturnObject:
		if dcName, ok := javaClassToDataMsg[m.Returns]; ok {
			sm.ReturnKind = "data_class"
			sm.DataClass = dcName
			// Use "extracted" variable name since the template extracts
			// from the JNI object into a Go struct via VM.Do.
			sm.DataClassConversion = buildDataClassConversion(dcName, dcFieldMap[dcName], "extracted")
		} else {
			sm.ReturnKind = "object"
			argsNeedHandles = true
		}
	}

	sm.NeedsHandles = argsNeedHandles

	return sm
}

// buildCallArgs generates the Go expression for the call arguments and
// reports whether any argument requires the handle store.
func buildCallArgs(params []javagen.MergedParam) (string, bool) {
	if len(params) == 0 {
		return "", false
	}
	needsHandles := false
	var parts []string
	for _, p := range params {
		if p.IsObject && !p.IsString {
			// Object handles are looked up in the HandleStore.
			protoGetter := fmt.Sprintf("req.Get%s()", exportName(p.GoName))
			parts = append(parts, fmt.Sprintf("s.Handles.Get(%s)", protoGetter))
			needsHandles = true
		} else {
			// Strings, bools, and primitives use the proto getter.
			protoGetter := fmt.Sprintf("req.Get%s()", exportName(p.GoName))
			parts = append(parts, convertProtoToGo(p.GoType, protoGetter))
		}
	}
	return strings.Join(parts, ", "), needsHandles
}

// convertProtoToGo wraps a proto getter expression with a type cast if needed.
func convertProtoToGo(goType, expr string) string {
	switch goType {
	case "string", "bool", "int32", "int64", "float32", "float64":
		return expr
	case "int16":
		return fmt.Sprintf("int16(%s)", expr)
	case "uint16":
		return fmt.Sprintf("uint16(%s)", expr)
	case "byte":
		return fmt.Sprintf("byte(%s)", expr)
	default:
		return expr
	}
}

// convertPrimitiveExpr returns the expression to convert a Go primitive to the
// proto field type.
func convertPrimitiveExpr(goType, varName string) string {
	switch goType {
	case "int32", "int64", "float32", "float64", "bool", "string":
		return varName
	case "int16", "uint16":
		return fmt.Sprintf("int32(%s)", varName)
	case "byte":
		return fmt.Sprintf("int32(%s)", varName)
	default:
		return varName
	}
}

// buildDataClassConversion renders the Go expression that converts a javagen
// data class struct to a proto message pointer.
func buildDataClassConversion(dcName string, fields []javagen.MergedField, varName string) string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "&pb.%s{\n", dcName)
	for _, f := range fields {
		protoFieldName := protoGoFieldName(f.GoName)
		protoType := protoGoType(f.CallSuffix, f.GoType)
		valueExpr := dataClassFieldExpr(f.GoType, protoType, varName+"."+f.GoName)
		fmt.Fprintf(&buf, "\t\t%s: %s,\n", protoFieldName, valueExpr)
	}
	fmt.Fprintf(&buf, "\t}")
	return buf.String()
}

// protoGoType returns the Go type of a proto field given the JNI call suffix
// and the javagen Go type.
func protoGoType(callSuffix, goType string) string {
	switch callSuffix {
	case "Boolean":
		return "bool"
	case "Byte":
		return "uint32"
	case "Short", "Int":
		return "int32"
	case "Long":
		return "int64"
	case "Float":
		return "float32"
	case "Double":
		return "float64"
	case "Object":
		if goType == "string" {
			return "string"
		}
		return "int64" // object handle
	default:
		return "int32"
	}
}

// dataClassFieldExpr returns the Go expression to assign a javagen field value
// to a proto message field, with type casting if needed.
func dataClassFieldExpr(goType, protoType, expr string) string {
	if goType == protoType {
		return expr
	}
	// Need type cast.
	return fmt.Sprintf("%s(%s)", protoType, expr)
}

// protoGoFieldName converts a javagen Go field name to the corresponding
// protobuf-generated Go struct field name. The path is:
//
//	GoName -> toSnakeCase -> protobuf CamelCase
//
// Examples: "SSID" -> "ssid" -> "Ssid", "IPAddress" -> "ip_address" -> "IpAddress"
func protoGoFieldName(goName string) string {
	snake := toSnakeCase(goName)
	return protoCamelCase(snake)
}

// toSnakeCase converts a PascalCase or camelCase string to snake_case.
func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := rune(s[i-1])
				switch {
				case prev >= 'a' && prev <= 'z':
					b.WriteByte('_')
				case prev >= 'A' && prev <= 'Z' && i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z':
					b.WriteByte('_')
				}
			}
			b.WriteRune(r - 'A' + 'a')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// protoCamelCase converts a snake_case string to the CamelCase used by
// protoc-gen-go. Each underscore-separated word gets its first letter
// uppercased; the rest stay lowercase.
func protoCamelCase(s string) string {
	var buf strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper && r >= 'a' && r <= 'z' {
			buf.WriteRune(r - 'a' + 'A')
		} else {
			buf.WriteRune(r)
		}
		upper = false
	}
	return buf.String()
}

// hasContextConstructor reports whether a class has a NewXxx(ctx *app.Context)
// constructor, which is required for the per-RPC manager creation pattern.
// Only system_service classes generate this constructor; context_method classes
// do not.
func hasContextConstructor(cls javagen.MergedClass) bool {
	switch cls.Kind {
	case "data_class", "iterable_data", "builder":
		return false
	}
	return cls.Obtain == "system_service"
}

// isExported reports whether a Go name is exported (starts with uppercase).
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

// exportName capitalizes the first letter of a Go name (as protoc-gen-go does).
func exportName(s string) string {
	if s == "" {
		return ""
	}
	first := s[0]
	if first >= 'a' && first <= 'z' {
		return string(first-'a'+'A') + s[1:]
	}
	return s
}
