package cligen

import (
	"strings"
	"unicode"

	"github.com/xaionaro-go/jni/tools/pkg/protogen"
)

// CLIPackage describes a single proto package's CLI commands.
type CLIPackage struct {
	GoModule    string
	PackageName string       // e.g., "alarm"
	VarPrefix   string       // e.g., "alarm" (for Go var names)
	Services    []CLIService // non-streaming services
}

// CLIService describes a gRPC service's CLI subcommands.
type CLIService struct {
	ProtoServiceName string       // e.g., "ManagerService"
	CobraName        string       // e.g., "manager"
	VarName          string       // e.g., "Manager" (for Go var names)
	Short            string       // cobra short description
	Commands         []CLICommand // leaf commands
}

// CLICommand describes a single RPC as a cobra leaf command.
type CLICommand struct {
	RPCName         string    // e.g., "Cancel"
	CobraName       string    // e.g., "cancel"
	VarName         string    // e.g., "Cancel" (for Go var names)
	Short           string    // cobra short description
	RequestType     string    // e.g., "CancelRequest"
	Flags           []CLIFlag // request message fields → flags
	ServerStreaming bool
	ClientStreaming bool
}

// CLIFlag describes a cobra flag derived from a proto request field.
type CLIFlag struct {
	CobraName  string // e.g., "carrier-frequency" (kebab-case)
	ProtoField string // e.g., "CarrierFrequency" (PascalCase, for req.Field = v)
	CobraType  string // e.g., "Int64" (for Flags().Get<Type> / Flags().<Type>)
	GoType     string // e.g., "int64"
	Default    string // e.g., "0", `""`, "false"
	ProtoType  string // e.g., "int64" (proto field Go type)
}

// buildCLIPackage converts ProtoData into a CLIPackage.
// Returns nil if there are no usable RPCs.
func buildCLIPackage(pd *protogen.ProtoData, goModule string) *CLIPackage {
	msgFields := buildMessageFieldMap(pd)

	pkg := &CLIPackage{
		GoModule:    goModule,
		PackageName: pd.Package,
		VarPrefix:   sanitizeVarName(pd.Package),
	}

	for _, svc := range pd.Services {
		cliSvc := buildCLIService(svc, msgFields)
		if len(cliSvc.Commands) == 0 {
			continue
		}
		pkg.Services = append(pkg.Services, cliSvc)
	}

	if len(pkg.Services) == 0 {
		return nil
	}
	return pkg
}

// buildMessageFieldMap creates a lookup from message name to its fields.
func buildMessageFieldMap(pd *protogen.ProtoData) map[string][]protogen.ProtoField {
	m := make(map[string][]protogen.ProtoField, len(pd.Messages))
	for _, msg := range pd.Messages {
		m[msg.Name] = msg.Fields
	}
	return m
}

// buildCLIService converts a ProtoService into a CLIService.
func buildCLIService(
	svc protogen.ProtoService,
	msgFields map[string][]protogen.ProtoField,
) CLIService {
	svcName := strings.TrimSuffix(svc.Name, "Service")
	cs := CLIService{
		ProtoServiceName: svc.Name,
		CobraName:        toKebabCase(svcName),
		VarName:          svcName,
		Short:            svc.Name + " operations",
	}

	for _, rpc := range svc.RPCs {
		// Skip bidi-streaming RPCs (require stdin interaction).
		if rpc.ClientStreaming && rpc.ServerStreaming {
			continue
		}

		cmd := buildCLICommand(rpc, msgFields)
		cs.Commands = append(cs.Commands, cmd)
	}

	return cs
}

// buildCLICommand converts a ProtoRPC into a CLICommand.
func buildCLICommand(
	rpc protogen.ProtoRPC,
	msgFields map[string][]protogen.ProtoField,
) CLICommand {
	cmd := CLICommand{
		RPCName:         rpc.Name,
		CobraName:       toKebabCase(rpc.Name),
		VarName:         rpc.Name,
		Short:           rpc.Name + " RPC",
		RequestType:     rpc.InputType,
		ServerStreaming: rpc.ServerStreaming,
		ClientStreaming: rpc.ClientStreaming,
	}

	fields := msgFields[rpc.InputType]
	for _, f := range fields {
		flag := buildCLIFlag(f)
		cmd.Flags = append(cmd.Flags, flag)
	}

	return cmd
}

// buildCLIFlag converts a ProtoField into a CLIFlag.
func buildCLIFlag(f protogen.ProtoField) CLIFlag {
	cobraType, goType, defaultVal := mapProtoTypeToFlag(f.Type)
	return CLIFlag{
		CobraName:  toKebabCase(snakeToCamel(f.Name)),
		ProtoField: snakeToPascal(f.Name),
		CobraType:  cobraType,
		GoType:     goType,
		Default:    defaultVal,
		ProtoType:  goType,
	}
}

// mapProtoTypeToFlag returns (cobraFlagType, goType, defaultValue)
// for a proto field type.
func mapProtoTypeToFlag(protoType string) (string, string, string) {
	switch protoType {
	case "string":
		return "String", "string", `""`
	case "bool":
		return "Bool", "bool", "false"
	case "int32":
		return "Int32", "int32", "0"
	case "int64":
		return "Int64", "int64", "0"
	case "uint32":
		return "Uint32", "uint32", "0"
	case "uint64":
		return "Uint64", "uint64", "0"
	case "float":
		return "Float32", "float32", "0"
	case "double":
		return "Float64", "float64", "0"
	default:
		// Complex/message types and handles → int64.
		return "Int64", "int64", "0"
	}
}

// toKebabCase converts PascalCase or camelCase to kebab-case.
func toKebabCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := rune(s[i-1])
				switch {
				case unicode.IsLower(prev):
					b.WriteByte('-')
				case unicode.IsUpper(prev) && i+1 < len(s) && unicode.IsLower(rune(s[i+1])):
					b.WriteByte('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// snakeToCamel converts snake_case to camelCase.
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// snakeToPascal converts snake_case to PascalCase.
// Trailing underscores are preserved (protoc adds them for Go keywords).
func snakeToPascal(s string) string {
	trailing := ""
	if strings.HasSuffix(s, "_") {
		trailing = "_"
		s = strings.TrimRight(s, "_")
	}
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "") + trailing
}

// sanitizeVarName converts a package name (possibly with underscores/slashes)
// into a valid Go identifier prefix.
func sanitizeVarName(s string) string {
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}
