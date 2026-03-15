package grpcgen

// ServerData holds all information needed to render a gRPC server file.
type ServerData struct {
	Package      string // Go package name, e.g. "location"
	GoModule     string // Go module path, e.g. "github.com/AndroidGoLab/jni"
	GoImport     string // Go import of the javagen package, e.g. "github.com/AndroidGoLab/jni/location"
	NeedsJNI     bool   // Whether the generated code needs to import the jni package
	NeedsHandles bool   // Whether the generated code needs to import the grpcserver handles package
	Services     []ServerService
	DataClasses  []ServerDataClass
}

// ServerDataClass describes a data class used for result conversion.
type ServerDataClass struct {
	GoType string
	Fields []ServerDataClassField
}

// ServerDataClassField describes a field in a data class.
type ServerDataClassField struct {
	GoName    string
	ProtoName string
	GoType    string
}

// ServerService describes a gRPC service backed by a javagen-generated class.
type ServerService struct {
	GoType       string // javagen type name, e.g. "Manager", "Adapter"
	ServiceName  string // proto service name, e.g. "ManagerService"
	Obtain       string // how the manager is obtained: "system_service", etc.
	Close        bool   // whether manager has Close()
	Methods      []ServerMethod
	NeedsHandles bool // whether any method uses object handles
}

// ServerMethod describes a single RPC method implementation.
type ServerMethod struct {
	GoName       string // Go method name on the gRPC service interface, e.g. "GetLastKnownLocation"
	SpecGoName   string // Go method name in the javagen code, e.g. "GetLastKnownLocation" or "getProvidersRaw"
	RequestType  string // Proto request message type, e.g. "GetLastKnownLocationRequest"
	ResponseType string // Proto response message type, e.g. "GetLastKnownLocationResponse"
	CallArgs     string // Pre-rendered Go arguments, e.g. "req.GetProvider()"
	ReturnKind   string // "void", "string", "bool", "object", "primitive", "data_class"
	DataClass    string // If returning a data class, its Go type, e.g. "Location"
	HasError     bool   // Whether the javagen method returns error
	HasResult    bool   // Whether the javagen method returns a non-void value
	GoReturnType string // Go type of the return value, e.g. "bool", "string", "*jni.Object"
	NeedsHandles bool   // Whether any param or return uses object handles
	// Pre-rendered conversion expression for result assignment in proto response.
	ResultExpr string // e.g. "result", "int32(result)"
	// Pre-rendered data class conversion for the response.
	DataClassConversion string // e.g. "&pb.Location{\n\tLatitude: result.Latitude,\n...}"
}
