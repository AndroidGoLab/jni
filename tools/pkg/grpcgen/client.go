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

// ClientData holds all information needed to render a gRPC client file.
type ClientData struct {
	Package     string          // Go package name, e.g. "location"
	GoModule    string          // Go module path, e.g. "github.com/xaionaro-go/jni"
	Services    []ClientService // One per gRPC service in the package
	DataClasses []ClientDataClass
}

// ClientService describes a gRPC service client wrapper.
type ClientService struct {
	GoType      string // Short name, e.g. "Manager"
	ServiceName string // Proto service name, e.g. "ManagerService"
	Methods     []ClientMethod
}

// ClientMethod describes a single client RPC wrapper method.
type ClientMethod struct {
	GoName       string // Exported Go method name on the client, e.g. "GetLastKnownLocation"
	RequestType  string // Proto request message type
	ResponseType string // Proto response message type
	// Params to accept in the Go API.
	Params []ClientParam
	// Return type information.
	ReturnKind   string // "void", "string", "bool", "primitive", "data_class", "object"
	GoReturnType string // Go type of the result, e.g. "bool", "string", "int32"
	DataClass    string // If returning a data class, its Go type
	HasError     bool   // Whether the underlying method returns error
	HasResult    bool   // Whether the method returns a non-void value
	// Pre-rendered expression for extracting the result from the response.
	ResultExpr string // e.g. "resp.Result", "resp.GetResult()"
}

// ClientParam describes a parameter in the client's Go API.
type ClientParam struct {
	GoName    string // Go parameter name
	GoType    string // Go type
	ProtoName string // Corresponding proto field name (exported)
	IsObject  bool   // Whether this is an object handle (int64 in proto)
}

// ClientDataClass describes a data class used for result conversion in the client.
type ClientDataClass struct {
	GoType string
	Fields []ClientDataClassField
}

// ClientDataClassField describes a field in a data class for client-side conversion.
type ClientDataClassField struct {
	GoName    string // Field name in the Go struct
	ProtoName string // Field name in the proto message
	GoType    string // Go type
	ProtoType string // Proto Go type (may differ, e.g. int32 vs int16)
}

// GenerateClient loads a Java API spec and overlay, merges them, builds client
// data structures, and writes a gRPC client implementation file.
// protoDir is the base directory containing compiled proto Go stubs (for name resolution).
func GenerateClient(specPath, overlayPath, outputDir, goModule, protoDir string) error {
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

	// Build proto data to get the canonical RPC names (with collision renames).
	protoData := protogen.BuildProtoData(merged, goModule)

	// Scan compiled proto stubs for actual Go names (handles protoc naming quirks).
	goNames := protoscan.Scan(filepath.Join(protoDir, merged.Package))

	data := buildClientData(merged, goModule, protoData, goNames)
	if len(data.Services) == 0 {
		return nil
	}

	pkgDir := filepath.Join(outputDir, "grpc", "client", merged.Package)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", pkgDir, err)
	}

	outputPath := filepath.Join(pkgDir, "client.go")
	if err := renderClient(data, outputPath); err != nil {
		return fmt.Errorf("render client: %w", err)
	}

	return nil
}

// buildClientData converts a MergedSpec into client template data.
// protoData provides canonical RPC names (with collision renames).
// goNames provides actual Go names from compiled proto stubs.
func buildClientData(
	merged *javagen.MergedSpec,
	goModule string,
	protoData *protogen.ProtoData,
	goNames protoscan.GoNames,
) *ClientData {
	data := &ClientData{
		Package:  merged.Package,
		GoModule: goModule,
	}

	// Build data class field info for result conversion.
	javaClassToDataClass := make(map[string]string)
	dcFieldMap := make(map[string][]javagen.MergedField)
	for _, dc := range merged.DataClasses {
		if !isExported(dc.GoType) {
			continue
		}
		javaClassToDataClass[dc.JavaClass] = dc.GoType
		dcFieldMap[dc.GoType] = dc.Fields
	}

	// Build data classes for client-side proto->Go conversion.
	for _, dc := range merged.DataClasses {
		if !isExported(dc.GoType) {
			continue
		}
		cdc := ClientDataClass{GoType: dc.GoType}
		for _, f := range dc.Fields {
			pt := protoGoType(f.CallSuffix, f.GoType)
			cdc.Fields = append(cdc.Fields, ClientDataClassField{
				GoName:    f.GoName,
				ProtoName: protoGoFieldName(f.GoName),
				GoType:    f.GoType,
				ProtoType: pt,
			})
		}
		data.DataClasses = append(data.DataClasses, cdc)
	}

	// Build RPC name lookup from protogen data (handles collision renames).
	// Maps lowercase(original name) → final RPC name.
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

		rawServiceName := exportName(cls.GoType) + "Service"
		serviceName := goNames.ResolveService(rawServiceName)
		svc := ClientService{
			GoType:      cls.GoType,
			ServiceName: serviceName,
		}

		for _, m := range cls.Methods {
			if !isExported(m.GoName) {
				continue
			}
			cm := buildClientMethod(m, javaClassToDataClass, dcFieldMap, protoRPCNames, goNames)
			svc.Methods = append(svc.Methods, cm)
		}

		if len(svc.Methods) == 0 {
			continue
		}
		data.Services = append(data.Services, svc)
	}

	return data
}

// buildClientMethod converts a MergedMethod to a ClientMethod.
func buildClientMethod(
	m javagen.MergedMethod,
	javaClassToDataMsg map[string]string,
	dcFieldMap map[string][]javagen.MergedField,
	protoRPCNames map[string]string,
	goNames protoscan.GoNames,
) ClientMethod {
	rawName := exportName(m.GoName)
	// Resolve through protogen (collision renames) then protoc (naming quirks).
	goName := rawName
	if resolved, ok := protoRPCNames[strings.ToLower(rawName)]; ok {
		goName = resolved
	}
	goName = goNames.ResolveRPC(goName)
	reqType := goName + "Request"
	respType := goName + "Response"

	cm := ClientMethod{
		GoName:       goName,
		RequestType:  reqType,
		ResponseType: respType,
		HasError:     m.Error,
		HasResult:    m.ReturnKind != javagen.ReturnVoid,
		GoReturnType: m.GoReturn,
	}

	// Build params for the client API.
	for _, p := range m.Params {
		cp := ClientParam{
			GoName:    p.GoName,
			ProtoName: exportName(p.GoName),
			IsObject:  p.IsObject && !p.IsString,
		}
		if cp.IsObject {
			// Object handles are int64 in proto.
			cp.GoType = "int64"
		} else {
			cp.GoType = p.GoType
		}
		cm.Params = append(cm.Params, cp)
	}

	// Determine return kind and result expression.
	switch m.ReturnKind {
	case javagen.ReturnVoid:
		cm.ReturnKind = "void"
	case javagen.ReturnString:
		cm.ReturnKind = "string"
		cm.ResultExpr = "resp.GetResult()"
		cm.GoReturnType = "string"
	case javagen.ReturnBool:
		cm.ReturnKind = "bool"
		cm.ResultExpr = "resp.GetResult()"
		cm.GoReturnType = "bool"
	case javagen.ReturnPrimitive:
		cm.ReturnKind = "primitive"
		cm.ResultExpr = clientPrimitiveResultExpr(m.GoReturn)
		cm.GoReturnType = m.GoReturn
	case javagen.ReturnObject:
		if dcName, ok := javaClassToDataMsg[m.Returns]; ok {
			cm.ReturnKind = "data_class"
			cm.DataClass = dcName
			cm.GoReturnType = "*" + dcName
		} else {
			cm.ReturnKind = "object"
			cm.ResultExpr = "resp.GetResult()"
			cm.GoReturnType = "int64"
		}
	}

	return cm
}

// clientPrimitiveResultExpr returns the expression to extract and convert
// a primitive result from the proto response.
func clientPrimitiveResultExpr(goType string) string {
	switch goType {
	case "int32", "int64", "float32", "float64", "bool", "string":
		return "resp.GetResult()"
	case "int16":
		return "int16(resp.GetResult())"
	case "uint16":
		return "uint16(resp.GetResult())"
	case "byte":
		return "byte(resp.GetResult())"
	default:
		return "resp.GetResult()"
	}
}

// buildClientParamAssign generates the expression to assign a Go parameter
// to a proto request field.
func buildClientParamAssign(p ClientParam) string {
	switch p.GoType {
	case "int16":
		return fmt.Sprintf("int32(%s)", p.GoName)
	case "uint16":
		return fmt.Sprintf("int32(%s)", p.GoName)
	case "byte":
		return fmt.Sprintf("int32(%s)", p.GoName)
	default:
		return p.GoName
	}
}

// clientGoZero returns the zero value for a Go type.
func clientGoZero(goType string) string {
	switch goType {
	case "string":
		return `""`
	case "bool":
		return "false"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "byte":
		return "0"
	default:
		if strings.HasPrefix(goType, "*") {
			return "nil"
		}
		return goType + "{}"
	}
}
