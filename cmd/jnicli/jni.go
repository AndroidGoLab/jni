package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	pb "github.com/AndroidGoLab/jni/proto/jni_raw"
)

var jniCmd = &cobra.Command{
	Use:   "jni",
	Short: "Raw JNI Env operations",
	Long:  "Direct access to the JNI Env surface: FindClass, GetMethodID, Call*Method, Get/SetField, string/array ops, etc.",
}

// ---- Version ----

var jniGetVersionCmd = &cobra.Command{
	Use:   "get-version",
	Short: "Get the JNI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetVersion(ctx, &pb.GetVersionRequest{})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Class ----

var jniClassCmd = &cobra.Command{
	Use:   "class",
	Short: "Class operations (find-class, get-superclass, is-assignable-from)",
}

var jniFindClassCmd = &cobra.Command{
	Use:   "find",
	Short: "Find a Java class by name (e.g. java/lang/String)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		name, _ := cmd.Flags().GetString("name")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.FindClass(ctx, &pb.FindClassRequest{Name: name})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetSuperclassCmd = &cobra.Command{
	Use:   "get-superclass",
	Short: "Get the superclass of a class",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetSuperclass(ctx, &pb.GetSuperclassRequest{ClassHandle: cls})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniIsAssignableFromCmd = &cobra.Command{
	Use:   "is-assignable-from",
	Short: "Check if class1 is assignable from class2",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		c1, _ := cmd.Flags().GetInt64("class1")
		c2, _ := cmd.Flags().GetInt64("class2")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.IsAssignableFrom(ctx, &pb.IsAssignableFromRequest{Class1: c1, Class2: c2})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Object ----

var jniObjectCmd = &cobra.Command{
	Use:   "object",
	Short: "Object operations (alloc, new, get-class, is-instance-of, is-same, ref-type)",
}

var jniAllocObjectCmd = &cobra.Command{
	Use:   "alloc",
	Short: "Allocate an object without calling a constructor",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.AllocObject(ctx, &pb.AllocObjectRequest{ClassHandle: cls})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniNewObjectCmd = &cobra.Command{
	Use:   "new",
	Short: "Create an object by calling a constructor",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		mid, _ := cmd.Flags().GetInt64("method")
		argStrs, _ := cmd.Flags().GetStringSlice("arg")
		jargs, err := parseJValues(argStrs)
		if err != nil {
			return err
		}
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.NewObject(ctx, &pb.NewObjectRequest{
			ClassHandle: cls,
			MethodId:    mid,
			Args:        jargs,
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetObjectClassCmd = &cobra.Command{
	Use:   "get-class",
	Short: "Get the class of an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetObjectClass(ctx, &pb.GetObjectClassRequest{ObjectHandle: obj})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniIsInstanceOfCmd = &cobra.Command{
	Use:   "is-instance-of",
	Short: "Check if an object is an instance of a class",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		cls, _ := cmd.Flags().GetInt64("class")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.IsInstanceOf(ctx, &pb.IsInstanceOfRequest{ObjectHandle: obj, ClassHandle: cls})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniIsSameObjectCmd = &cobra.Command{
	Use:   "is-same",
	Short: "Check if two handles refer to the same object",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		o1, _ := cmd.Flags().GetInt64("object1")
		o2, _ := cmd.Flags().GetInt64("object2")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.IsSameObject(ctx, &pb.IsSameObjectRequest{Object1: o1, Object2: o2})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetObjectRefTypeCmd = &cobra.Command{
	Use:   "ref-type",
	Short: "Get the reference type of an object (local/global/weak)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetObjectRefType(ctx, &pb.GetObjectRefTypeRequest{ObjectHandle: obj})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Method/Field ID lookup ----

var jniMethodCmd = &cobra.Command{
	Use:   "method",
	Short: "Method operations (get-id, get-static-id, call, call-static, call-nonvirtual)",
}

var jniGetMethodIDCmd = &cobra.Command{
	Use:   "get-id",
	Short: "Get a method ID by name and signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		name, _ := cmd.Flags().GetString("name")
		sig, _ := cmd.Flags().GetString("sig")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{ClassHandle: cls, Name: name, Sig: sig})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetStaticMethodIDCmd = &cobra.Command{
	Use:   "get-static-id",
	Short: "Get a static method ID by name and signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		name, _ := cmd.Flags().GetString("name")
		sig, _ := cmd.Flags().GetString("sig")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetStaticMethodID(ctx, &pb.GetStaticMethodIDRequest{ClassHandle: cls, Name: name, Sig: sig})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniCallMethodCmd = &cobra.Command{
	Use:   "call",
	Short: "Call an instance method (specify --return-type: void,boolean,byte,char,short,int,long,float,double,object)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		mid, _ := cmd.Flags().GetInt64("method")
		retTypeStr, _ := cmd.Flags().GetString("return-type")
		argStrs, _ := cmd.Flags().GetStringSlice("arg")
		retType := parseJType(retTypeStr)
		jargs, err := parseJValues(argStrs)
		if err != nil {
			return err
		}
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.CallMethod(ctx, &pb.CallMethodRequest{
			ObjectHandle: obj,
			MethodId:     mid,
			ReturnType:   retType,
			Args:         jargs,
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniCallStaticMethodCmd = &cobra.Command{
	Use:   "call-static",
	Short: "Call a static method",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		mid, _ := cmd.Flags().GetInt64("method")
		retTypeStr, _ := cmd.Flags().GetString("return-type")
		argStrs, _ := cmd.Flags().GetStringSlice("arg")
		retType := parseJType(retTypeStr)
		jargs, err := parseJValues(argStrs)
		if err != nil {
			return err
		}
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.CallStaticMethod(ctx, &pb.CallStaticMethodRequest{
			ClassHandle: cls,
			MethodId:    mid,
			ReturnType:  retType,
			Args:        jargs,
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Field ----

var jniFieldCmd = &cobra.Command{
	Use:   "field",
	Short: "Field operations (get-id, get-static-id, get, set, get-static, set-static)",
}

var jniGetFieldIDCmd = &cobra.Command{
	Use:   "get-id",
	Short: "Get a field ID by name and signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		name, _ := cmd.Flags().GetString("name")
		sig, _ := cmd.Flags().GetString("sig")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetFieldID(ctx, &pb.GetFieldIDRequest{ClassHandle: cls, Name: name, Sig: sig})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetStaticFieldIDCmd = &cobra.Command{
	Use:   "get-static-id",
	Short: "Get a static field ID by name and signature",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		name, _ := cmd.Flags().GetString("name")
		sig, _ := cmd.Flags().GetString("sig")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetStaticFieldID(ctx, &pb.GetStaticFieldIDRequest{ClassHandle: cls, Name: name, Sig: sig})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetFieldCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an instance field value",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		fid, _ := cmd.Flags().GetInt64("field")
		ftStr, _ := cmd.Flags().GetString("type")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetField(ctx, &pb.GetFieldValueRequest{
			ObjectHandle: obj,
			FieldId:      fid,
			FieldType:    parseJType(ftStr),
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniSetFieldCmd = &cobra.Command{
	Use:   "set",
	Short: "Set an instance field value",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		fid, _ := cmd.Flags().GetInt64("field")
		valStr, _ := cmd.Flags().GetString("value")
		jval, err := parseSingleJValue(valStr)
		if err != nil {
			return err
		}
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.SetField(ctx, &pb.SetFieldValueRequest{
			ObjectHandle: obj,
			FieldId:      fid,
			Value:        jval,
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetStaticFieldCmd = &cobra.Command{
	Use:   "get-static",
	Short: "Get a static field value",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		cls, _ := cmd.Flags().GetInt64("class")
		fid, _ := cmd.Flags().GetInt64("field")
		ftStr, _ := cmd.Flags().GetString("type")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetStaticField(ctx, &pb.GetStaticFieldValueRequest{
			ClassHandle: cls,
			FieldId:     fid,
			FieldType:   parseJType(ftStr),
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- String ----

var jniStringCmd = &cobra.Command{
	Use:   "string",
	Short: "String operations (new, get, length)",
}

var jniNewStringCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new Java string from UTF-8",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		val, _ := cmd.Flags().GetString("value")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.NewStringUTF(ctx, &pb.NewStringUTFRequest{Value: val})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetStringCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a Java string as UTF-8",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		handle, _ := cmd.Flags().GetInt64("handle")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetStringUTFChars(ctx, &pb.GetStringUTFCharsRequest{StringHandle: handle})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetStringLengthCmd = &cobra.Command{
	Use:   "length",
	Short: "Get the length of a Java string",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		handle, _ := cmd.Flags().GetInt64("handle")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetStringLength(ctx, &pb.GetStringLengthRequest{StringHandle: handle})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Array ----

var jniArrayCmd = &cobra.Command{
	Use:   "array",
	Short: "Array operations (new-primitive, new-object, length, get-element, set-element, get-region, set-region)",
}

var jniNewPrimitiveArrayCmd = &cobra.Command{
	Use:   "new-primitive",
	Short: "Create a new primitive array",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		etStr, _ := cmd.Flags().GetString("element-type")
		length, _ := cmd.Flags().GetInt32("length")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.NewPrimitiveArray(ctx, &pb.NewPrimitiveArrayRequest{
			ElementType: parseJType(etStr),
			Length:      length,
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniNewObjectArrayCmd = &cobra.Command{
	Use:   "new-object",
	Short: "Create a new object array",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		length, _ := cmd.Flags().GetInt32("length")
		cls, _ := cmd.Flags().GetInt64("class")
		init, _ := cmd.Flags().GetInt64("init")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.NewObjectArray(ctx, &pb.NewObjectArrayRequest{
			Length:      length,
			ClassHandle: cls,
			InitElement: init,
		})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniGetArrayLengthCmd = &cobra.Command{
	Use:   "length",
	Short: "Get the length of an array",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		handle, _ := cmd.Flags().GetInt64("handle")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.GetArrayLength(ctx, &pb.GetArrayLengthRequest{ArrayHandle: handle})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Exception ----

var jniExceptionCmd = &cobra.Command{
	Use:   "exception",
	Short: "Exception operations (check, clear, describe, occurred, throw, throw-new)",
}

var jniExceptionCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a JNI exception is pending",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.ExceptionCheck(ctx, &pb.ExceptionCheckRequest{})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniExceptionClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear a pending JNI exception",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.ExceptionClear(ctx, &pb.ExceptionClearRequest{})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Reference ----

var jniRefCmd = &cobra.Command{
	Use:   "ref",
	Short: "Reference management (new-global, delete-global, new-local, delete-local, new-weak, delete-weak)",
}

var jniNewGlobalRefCmd = &cobra.Command{
	Use:   "new-global",
	Short: "Create a global reference to an object",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		obj, _ := cmd.Flags().GetInt64("object")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.NewGlobalRef(ctx, &pb.NewGlobalRefRequest{ObjectHandle: obj})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

var jniDeleteGlobalRefCmd = &cobra.Command{
	Use:   "delete-global",
	Short: "Delete a global reference",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		ref, _ := cmd.Flags().GetInt64("ref")
		client := pb.NewJNIServiceClient(grpcConn)
		resp, err := client.DeleteGlobalRef(ctx, &pb.DeleteGlobalRefRequest{RefHandle: ref})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

// ---- Helpers ----

func parseJType(s string) pb.JType {
	switch strings.ToLower(s) {
	case "void":
		return pb.JType_VOID
	case "boolean", "bool", "z":
		return pb.JType_BOOLEAN
	case "byte", "b":
		return pb.JType_BYTE
	case "char", "c":
		return pb.JType_CHAR
	case "short", "s":
		return pb.JType_SHORT
	case "int", "i":
		return pb.JType_INT
	case "long", "j":
		return pb.JType_LONG
	case "float", "f":
		return pb.JType_FLOAT
	case "double", "d":
		return pb.JType_DOUBLE
	case "object", "l":
		return pb.JType_OBJECT
	default:
		return pb.JType_VOID
	}
}

// parseJValues parses CLI arg strings in "type:value" format.
// Examples: "int:42", "long:100", "bool:true", "object:5", "string:hello"
func parseJValues(args []string) ([]*pb.JValue, error) {
	result := make([]*pb.JValue, 0, len(args))
	for _, a := range args {
		v, err := parseSingleJValue(a)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func parseSingleJValue(s string) (*pb.JValue, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid jvalue %q: expected type:value (e.g. int:42)", s)
	}
	typ, val := parts[0], parts[1]
	switch strings.ToLower(typ) {
	case "boolean", "bool", "z":
		b, err := strconv.ParseBool(val)
		if err != nil {
			return nil, fmt.Errorf("parse boolean %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_Z{Z: b}}, nil
	case "byte", "b":
		n, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("parse byte %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_B{B: int32(n)}}, nil
	case "char", "c":
		n, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("parse char %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_C{C: uint32(n)}}, nil
	case "short", "s":
		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("parse short %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_S{S: int32(n)}}, nil
	case "int", "i":
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parse int %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_I{I: int32(n)}}, nil
	case "long", "j":
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse long %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_J{J: n}}, nil
	case "float", "f":
		n, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return nil, fmt.Errorf("parse float %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_F{F: float32(n)}}, nil
	case "double", "d":
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("parse double %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_D{D: n}}, nil
	case "object", "l":
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse object handle %q: %w", val, err)
		}
		return &pb.JValue{Value: &pb.JValue_L{L: n}}, nil
	default:
		return nil, fmt.Errorf("unknown jvalue type %q", typ)
	}
}

func init() {
	// Class
	jniFindClassCmd.Flags().String("name", "", "class name (e.g. java/lang/String)")
	_ = jniFindClassCmd.MarkFlagRequired("name")
	jniGetSuperclassCmd.Flags().Int64("class", 0, "class handle")
	jniIsAssignableFromCmd.Flags().Int64("class1", 0, "first class handle")
	jniIsAssignableFromCmd.Flags().Int64("class2", 0, "second class handle")
	jniClassCmd.AddCommand(jniFindClassCmd, jniGetSuperclassCmd, jniIsAssignableFromCmd)

	// Object
	jniAllocObjectCmd.Flags().Int64("class", 0, "class handle")
	jniNewObjectCmd.Flags().Int64("class", 0, "class handle")
	jniNewObjectCmd.Flags().Int64("method", 0, "constructor method ID")
	jniNewObjectCmd.Flags().StringSlice("arg", nil, "arguments as type:value (e.g. int:42)")
	jniGetObjectClassCmd.Flags().Int64("object", 0, "object handle")
	jniIsInstanceOfCmd.Flags().Int64("object", 0, "object handle")
	jniIsInstanceOfCmd.Flags().Int64("class", 0, "class handle")
	jniIsSameObjectCmd.Flags().Int64("object1", 0, "first object handle")
	jniIsSameObjectCmd.Flags().Int64("object2", 0, "second object handle")
	jniGetObjectRefTypeCmd.Flags().Int64("object", 0, "object handle")
	jniObjectCmd.AddCommand(jniAllocObjectCmd, jniNewObjectCmd, jniGetObjectClassCmd, jniIsInstanceOfCmd, jniIsSameObjectCmd, jniGetObjectRefTypeCmd)

	// Method
	jniGetMethodIDCmd.Flags().Int64("class", 0, "class handle")
	jniGetMethodIDCmd.Flags().String("name", "", "method name")
	jniGetMethodIDCmd.Flags().String("sig", "", "JNI method signature (e.g. (I)V)")
	jniGetStaticMethodIDCmd.Flags().Int64("class", 0, "class handle")
	jniGetStaticMethodIDCmd.Flags().String("name", "", "method name")
	jniGetStaticMethodIDCmd.Flags().String("sig", "", "JNI method signature")
	jniCallMethodCmd.Flags().Int64("object", 0, "object handle")
	jniCallMethodCmd.Flags().Int64("method", 0, "method ID")
	jniCallMethodCmd.Flags().String("return-type", "void", "return type (void,boolean,byte,char,short,int,long,float,double,object)")
	jniCallMethodCmd.Flags().StringSlice("arg", nil, "arguments as type:value")
	jniCallStaticMethodCmd.Flags().Int64("class", 0, "class handle")
	jniCallStaticMethodCmd.Flags().Int64("method", 0, "method ID")
	jniCallStaticMethodCmd.Flags().String("return-type", "void", "return type")
	jniCallStaticMethodCmd.Flags().StringSlice("arg", nil, "arguments as type:value")
	jniMethodCmd.AddCommand(jniGetMethodIDCmd, jniGetStaticMethodIDCmd, jniCallMethodCmd, jniCallStaticMethodCmd)

	// Field
	jniGetFieldIDCmd.Flags().Int64("class", 0, "class handle")
	jniGetFieldIDCmd.Flags().String("name", "", "field name")
	jniGetFieldIDCmd.Flags().String("sig", "", "JNI field signature (e.g. I, Ljava/lang/String;)")
	jniGetStaticFieldIDCmd.Flags().Int64("class", 0, "class handle")
	jniGetStaticFieldIDCmd.Flags().String("name", "", "field name")
	jniGetStaticFieldIDCmd.Flags().String("sig", "", "JNI field signature")
	jniGetFieldCmd.Flags().Int64("object", 0, "object handle")
	jniGetFieldCmd.Flags().Int64("field", 0, "field ID")
	jniGetFieldCmd.Flags().String("type", "int", "field type")
	jniSetFieldCmd.Flags().Int64("object", 0, "object handle")
	jniSetFieldCmd.Flags().Int64("field", 0, "field ID")
	jniSetFieldCmd.Flags().String("value", "", "value as type:value (e.g. int:42)")
	jniGetStaticFieldCmd.Flags().Int64("class", 0, "class handle")
	jniGetStaticFieldCmd.Flags().Int64("field", 0, "field ID")
	jniGetStaticFieldCmd.Flags().String("type", "int", "field type")
	jniFieldCmd.AddCommand(jniGetFieldIDCmd, jniGetStaticFieldIDCmd, jniGetFieldCmd, jniSetFieldCmd, jniGetStaticFieldCmd)

	// String
	jniNewStringCmd.Flags().String("value", "", "UTF-8 string value")
	jniGetStringCmd.Flags().Int64("handle", 0, "string handle")
	jniGetStringLengthCmd.Flags().Int64("handle", 0, "string handle")
	jniStringCmd.AddCommand(jniNewStringCmd, jniGetStringCmd, jniGetStringLengthCmd)

	// Array
	jniNewPrimitiveArrayCmd.Flags().String("element-type", "int", "element type")
	jniNewPrimitiveArrayCmd.Flags().Int32("length", 0, "array length")
	jniNewObjectArrayCmd.Flags().Int32("length", 0, "array length")
	jniNewObjectArrayCmd.Flags().Int64("class", 0, "element class handle")
	jniNewObjectArrayCmd.Flags().Int64("init", 0, "initial element handle (0 for null)")
	jniGetArrayLengthCmd.Flags().Int64("handle", 0, "array handle")
	jniArrayCmd.AddCommand(jniNewPrimitiveArrayCmd, jniNewObjectArrayCmd, jniGetArrayLengthCmd)

	// Exception
	jniExceptionCmd.AddCommand(jniExceptionCheckCmd, jniExceptionClearCmd)

	// Reference
	jniNewGlobalRefCmd.Flags().Int64("object", 0, "object handle")
	jniDeleteGlobalRefCmd.Flags().Int64("ref", 0, "reference handle")
	jniRefCmd.AddCommand(jniNewGlobalRefCmd, jniDeleteGlobalRefCmd)

	// Wire up top-level
	jniCmd.AddCommand(jniGetVersionCmd, jniClassCmd, jniObjectCmd, jniMethodCmd, jniFieldCmd, jniStringCmd, jniArrayCmd, jniExceptionCmd, jniRefCmd)
	rootCmd.AddCommand(jniCmd)
}
