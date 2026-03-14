package protogen

import (
	"strings"

	"github.com/xaionaro-go/jni/tools/pkg/javagen"
)

// ProtoData holds all information needed to render a .proto file.
type ProtoData struct {
	Package   string
	GoPackage string
	Services  []ProtoService
	Messages  []ProtoMessage
	Enums     []ProtoEnum
}

// ProtoService describes a gRPC service generated from a Java class.
type ProtoService struct {
	Name string
	RPCs []ProtoRPC
}

// ProtoRPC describes a single RPC method in a gRPC service.
type ProtoRPC struct {
	Name           string
	OriginalName   string // Name before collision renaming (empty if unchanged).
	InputType      string
	OutputType     string
	ClientStreaming bool
	ServerStreaming bool
	Comment        string
}

// ProtoMessage describes a protobuf message type.
type ProtoMessage struct {
	Name   string
	Fields []ProtoField
}

// ProtoField describes a single field in a protobuf message.
type ProtoField struct {
	Type     string
	Name     string
	Number   int
	Repeated bool
	Optional bool
	Comment  string
}

// ProtoEnum describes a protobuf enum type.
type ProtoEnum struct {
	Name   string
	Values []ProtoEnumValue
}

// ProtoEnumValue describes a single value in a protobuf enum.
type ProtoEnumValue struct {
	Name   string
	Number int
}

// BuildProtoData converts a MergedSpec into proto data structures.
func BuildProtoData(merged *javagen.MergedSpec, goModule string) *ProtoData {
	data := &ProtoData{
		Package:   merged.Package,
		GoPackage: goModule + "/proto/" + merged.Package,
	}

	// Build a map from Java class name to data class proto message name so we
	// can reference data class messages in RPC responses. Proto convention
	// requires PascalCase, so we capitalize the first letter of GoType.
	javaClassToDataMsg := make(map[string]string)
	dataClassNames := make(map[string]bool)
	for _, dc := range merged.DataClasses {
		protoName := capitalizeFirst(dc.GoType)
		dataClassNames[protoName] = true
		javaClassToDataMsg[dc.JavaClass] = protoName
	}

	// 1. Messages from data classes.
	for _, dc := range merged.DataClasses {
		msg := buildDataClassMessage(dc, dataClassNames)
		data.Messages = append(data.Messages, msg)
	}

	// Track seen message names to avoid duplicates when multiple services
	// share the same method names (e.g., connect, close).
	seenMessages := make(map[string]bool)
	for _, m := range data.Messages {
		seenMessages[m.Name] = true
	}
	appendUniqueMessages := func(msgs []ProtoMessage) {
		for _, m := range msgs {
			if seenMessages[m.Name] {
				continue
			}
			seenMessages[m.Name] = true
			data.Messages = append(data.Messages, m)
		}
	}

	// 2. Services from classes that have methods.
	for _, cls := range merged.Classes {
		if len(cls.Methods) == 0 {
			continue
		}
		svc, msgs := buildServiceFromClass(cls, dataClassNames, javaClassToDataMsg)
		data.Services = append(data.Services, svc)
		appendUniqueMessages(msgs)
	}

	// 3. Streaming RPCs from callbacks.
	for _, cb := range merged.Callbacks {
		rpcs, msgs := buildStreamingFromCallback(cb, dataClassNames)
		if len(rpcs) == 0 {
			continue
		}
		// Add streaming RPCs to the first service, or create a new one.
		svcName := cb.GoType + "Service"
		// Capitalize the first letter for the service name.
		if len(svcName) > 0 {
			svcName = strings.ToUpper(svcName[:1]) + svcName[1:]
		}
		svcIdx := -1
		for i := range data.Services {
			if data.Services[i].Name == svcName {
				svcIdx = i
				break
			}
		}
		if svcIdx >= 0 {
			data.Services[svcIdx].RPCs = append(data.Services[svcIdx].RPCs, rpcs...)
		} else {
			data.Services = append(data.Services, ProtoService{
				Name: svcName,
				RPCs: rpcs,
			})
		}
		appendUniqueMessages(msgs)
	}

	// 4. Fix RPC names that collide with message names. Proto3 shares
	// the namespace between services, messages, and enums within a
	// package. Append "_" suffix to the colliding RPC name.
	allNames := make(map[string]bool, len(data.Messages))
	for _, m := range data.Messages {
		allNames[m.Name] = true
	}
	for _, e := range data.Enums {
		allNames[e.Name] = true
	}
	for si := range data.Services {
		for ri := range data.Services[si].RPCs {
			rpc := &data.Services[si].RPCs[ri]
			if !allNames[rpc.Name] {
				continue
			}

			// Find the original request/response messages to copy their fields.
			var origReqFields, origRespFields []ProtoField
			for _, m := range data.Messages {
				switch m.Name {
				case rpc.InputType:
					origReqFields = m.Fields
				case rpc.OutputType:
					origRespFields = m.Fields
				}
			}

			rpc.OriginalName = rpc.Name
			newName := rpc.Name + "Op"
			newReq := newName + "Request"
			newResp := newName + "Response"

			// Create renamed request/response messages with copied fields.
			data.Messages = append(data.Messages,
				ProtoMessage{Name: newReq, Fields: origReqFields},
				ProtoMessage{Name: newResp, Fields: origRespFields},
			)

			rpc.Name = newName
			rpc.InputType = newReq
			rpc.OutputType = newResp
		}
	}

	return data
}

// buildDataClassMessage converts a MergedDataClass into a ProtoMessage.
func buildDataClassMessage(dc javagen.MergedDataClass, dataClassNames map[string]bool) ProtoMessage {
	msg := ProtoMessage{Name: capitalizeFirst(dc.GoType)}
	for i, f := range dc.Fields {
		fieldType := protoTypeFromCallSuffix(f.CallSuffix, f.GoType)
		// If the field's GoType matches a known data class, use the message name.
		capitalizedType := capitalizeFirst(f.GoType)
		if f.CallSuffix == "Object" && dataClassNames[capitalizedType] {
			fieldType = capitalizedType
		}
		msg.Fields = append(msg.Fields, ProtoField{
			Type:   fieldType,
			Name:   toSnakeCase(f.GoName),
			Number: i + 1,
		})
	}
	return msg
}

// isNonServiceKind reports whether a class kind should be skipped during
// service generation (e.g. iterable_data, builder).
func isNonServiceKind(kind string) bool {
	switch kind {
	case "iterable_data", "builder":
		return true
	}
	return false
}

// buildServiceFromClass converts a MergedClass into a ProtoService and its
// associated request/response messages.
func buildServiceFromClass(cls javagen.MergedClass, dataClassNames map[string]bool, javaClassToDataMsg map[string]string) (ProtoService, []ProtoMessage) {
	svc := ProtoService{Name: capitalizeFirst(cls.GoType) + "Service"}
	var msgs []ProtoMessage

	for _, m := range cls.Methods {
		rpc, reqMsg, respMsg := buildRPCFromMethod(m, dataClassNames, javaClassToDataMsg)
		svc.RPCs = append(svc.RPCs, rpc)
		msgs = append(msgs, reqMsg)
		if respMsg != nil {
			msgs = append(msgs, *respMsg)
		}
	}

	return svc, msgs
}

// buildRPCFromMethod converts a MergedMethod into a ProtoRPC and its
// request/response messages.
func buildRPCFromMethod(m javagen.MergedMethod, dataClassNames map[string]bool, javaClassToDataMsg map[string]string) (ProtoRPC, ProtoMessage, *ProtoMessage) {
	goName := capitalizeFirst(m.GoName)
	reqName := goName + "Request"
	respName := goName + "Response"

	// Build request message.
	reqMsg := ProtoMessage{Name: reqName}
	for i, p := range m.Params {
		reqMsg.Fields = append(reqMsg.Fields, ProtoField{
			Type:   protoTypeFromParam(p),
			Name:   toSnakeCase(p.GoName),
			Number: i + 1,
		})
	}

	// Build response message.
	var respMsg *ProtoMessage
	if m.ReturnKind != javagen.ReturnVoid {
		rm := ProtoMessage{Name: respName}
		resultType := protoTypeFromReturn(m, dataClassNames, javaClassToDataMsg)
		rm.Fields = append(rm.Fields, ProtoField{
			Type:   resultType,
			Name:   "result",
			Number: 1,
		})
		respMsg = &rm
	} else {
		// For void methods, create an empty response message.
		respMsg = &ProtoMessage{Name: respName}
	}

	rpc := ProtoRPC{
		Name:       goName,
		InputType:  reqName,
		OutputType: respName,
	}

	return rpc, reqMsg, respMsg
}

// buildStreamingFromCallback creates streaming RPCs and event messages from a
// callback interface.
func buildStreamingFromCallback(cb javagen.MergedCallback, dataClassNames map[string]bool) ([]ProtoRPC, []ProtoMessage) {
	pattern := DetectStreamingPattern(&cb)

	goType := cb.GoType
	// Capitalize first letter.
	if len(goType) > 0 {
		goType = strings.ToUpper(goType[:1]) + goType[1:]
	}

	var rpcs []ProtoRPC
	var msgs []ProtoMessage

	// Create event messages for each callback method.
	for _, m := range cb.Methods {
		eventName := goType + m.GoField + "Event"
		eventMsg := ProtoMessage{Name: eventName}
		for i, p := range m.Params {
			eventMsg.Fields = append(eventMsg.Fields, ProtoField{
				Type:   protoTypeFromParam(p),
				Name:   toSnakeCase(p.GoName),
				Number: i + 1,
			})
		}
		msgs = append(msgs, eventMsg)
	}

	// Create a unified event message that wraps all callback methods using oneof.
	// For simplicity, we use a single response message with all fields.
	streamEventName := goType + "Event"
	streamEventMsg := ProtoMessage{Name: streamEventName}
	fieldNum := 1
	for _, m := range cb.Methods {
		// Add a field referencing each specific event message.
		streamEventMsg.Fields = append(streamEventMsg.Fields, ProtoField{
			Type:     goType + m.GoField + "Event",
			Name:     toSnakeCase(m.GoField),
			Number:   fieldNum,
			Optional: true,
		})
		fieldNum++
	}
	msgs = append(msgs, streamEventMsg)

	// Create the subscribe request message.
	subscribeReqName := "Subscribe" + goType + "Request"
	msgs = append(msgs, ProtoMessage{Name: subscribeReqName})

	switch pattern {
	case ServerStreaming:
		rpcs = append(rpcs, ProtoRPC{
			Name:           "Subscribe" + goType,
			InputType:      subscribeReqName,
			OutputType:     streamEventName,
			ServerStreaming: true,
			Comment:        "Server-streaming events from " + cb.JavaInterface,
		})
	case BidiStreaming:
		// For bidi, create a client message too.
		clientMsgName := goType + "Command"
		msgs = append(msgs, ProtoMessage{Name: clientMsgName})
		rpcs = append(rpcs, ProtoRPC{
			Name:           goType + "Stream",
			InputType:      clientMsgName,
			OutputType:     streamEventName,
			ClientStreaming: true,
			ServerStreaming: true,
			Comment:        "Bidirectional streaming for " + cb.JavaInterface,
		})
	}

	return rpcs, msgs
}

// protoTypeFromCallSuffix maps a JNI CallSuffix and GoType to a proto type.
func protoTypeFromCallSuffix(callSuffix, goType string) string {
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
		return "float"
	case "Double":
		return "double"
	case "Object":
		if goType == "string" {
			return "string"
		}
		// Object handles are passed as int64.
		return "int64"
	default:
		return "int32"
	}
}

// protoTypeFromParam maps a MergedParam to a proto type.
func protoTypeFromParam(p javagen.MergedParam) string {
	if p.IsString {
		return "string"
	}
	if p.IsBool {
		return "bool"
	}
	if p.IsObject {
		// Object handles are passed as int64.
		return "int64"
	}
	// Infer from GoType for primitives.
	return protoTypeFromGoType(p.GoType)
}

// protoTypeFromGoType maps a Go type to a proto type.
func protoTypeFromGoType(goType string) string {
	switch goType {
	case "bool":
		return "bool"
	case "byte", "uint8":
		return "uint32"
	case "int16", "int32", "int":
		return "int32"
	case "int64":
		return "int64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "string":
		return "string"
	case "uint16":
		return "uint32"
	default:
		// Unknown types (e.g. *jni.Object) become handle references.
		return "int64"
	}
}

// protoTypeFromReturn maps a method's return info to a proto type, checking
// both the GoReturn name and the Java return type against known data classes.
func protoTypeFromReturn(m javagen.MergedMethod, dataClassNames map[string]bool, javaClassToDataMsg map[string]string) string {
	switch m.ReturnKind {
	case javagen.ReturnString:
		return "string"
	case javagen.ReturnBool:
		return "bool"
	case javagen.ReturnObject:
		// Check if GoReturn matches a data class name directly.
		capitalizedReturn := capitalizeFirst(m.GoReturn)
		if dataClassNames[capitalizedReturn] {
			return capitalizedReturn
		}
		// Check if the Java return type maps to a data class.
		if msgName, ok := javaClassToDataMsg[m.Returns]; ok {
			return msgName
		}
		return "int64"
	case javagen.ReturnPrimitive:
		return protoTypeFromCallSuffix(m.CallSuffix, m.GoReturn)
	default:
		return "int64"
	}
}

// capitalizeFirst capitalizes the first letter of a string.
// Proto convention requires PascalCase for service, message, and RPC names.
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	first := s[0]
	if first >= 'a' && first <= 'z' {
		return string(first-'a'+'A') + s[1:]
	}
	return s
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
