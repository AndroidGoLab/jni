package main

import (
	"strconv"

	"github.com/spf13/cobra"
	pb "github.com/AndroidGoLab/jni/proto/jni_raw"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Device information",
}

var deviceInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get device model, manufacturer, SDK version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := pb.NewJNIServiceClient(grpcConn)

		buildCls, err := client.FindClass(ctx, &pb.FindClassRequest{Name: "android/os/Build"})
		if err != nil {
			return err
		}

		getStringField := func(name string) string {
			fid, err := client.GetStaticFieldID(ctx, &pb.GetStaticFieldIDRequest{
				ClassHandle: buildCls.GetClassHandle(), Name: name, Sig: "Ljava/lang/String;",
			})
			if err != nil {
				return ""
			}
			val, err := client.GetStaticField(ctx, &pb.GetStaticFieldValueRequest{
				ClassHandle: buildCls.GetClassHandle(),
				FieldId:     fid.GetFieldId(),
				FieldType:   pb.JType_OBJECT,
			})
			if err != nil {
				return ""
			}
			h := val.GetResult().GetL()
			if h == 0 {
				return ""
			}
			str, err := client.GetStringUTFChars(ctx, &pb.GetStringUTFCharsRequest{StringHandle: h})
			if err != nil {
				return ""
			}
			return str.GetValue()
		}

		verCls, err := client.FindClass(ctx, &pb.FindClassRequest{Name: "android/os/Build$VERSION"})
		if err != nil {
			return err
		}
		sdkFID, err := client.GetStaticFieldID(ctx, &pb.GetStaticFieldIDRequest{
			ClassHandle: verCls.GetClassHandle(), Name: "SDK_INT", Sig: "I",
		})
		if err != nil {
			return err
		}
		sdkVal, err := client.GetStaticField(ctx, &pb.GetStaticFieldValueRequest{
			ClassHandle: verCls.GetClassHandle(),
			FieldId:     sdkFID.GetFieldId(),
			FieldType:   pb.JType_INT,
		})
		if err != nil {
			return err
		}

		releaseFID, _ := client.GetStaticFieldID(ctx, &pb.GetStaticFieldIDRequest{
			ClassHandle: verCls.GetClassHandle(), Name: "RELEASE", Sig: "Ljava/lang/String;",
		})
		releaseVal, _ := client.GetStaticField(ctx, &pb.GetStaticFieldValueRequest{
			ClassHandle: verCls.GetClassHandle(), FieldId: releaseFID.GetFieldId(), FieldType: pb.JType_OBJECT,
		})
		releaseStr, _ := client.GetStringUTFChars(ctx, &pb.GetStringUTFCharsRequest{
			StringHandle: releaseVal.GetResult().GetL(),
		})

		return printResult(map[string]any{
			"manufacturer": getStringField("MANUFACTURER"),
			"model":        getStringField("MODEL"),
			"brand":        getStringField("BRAND"),
			"device":       getStringField("DEVICE"),
			"product":      getStringField("PRODUCT"),
			"sdk_int":      strconv.Itoa(int(sdkVal.GetResult().GetI())),
			"release":      releaseStr.GetValue(),
		})
	},
}

func init() {
	deviceCmd.AddCommand(deviceInfoCmd)
	rootCmd.AddCommand(deviceCmd)
}
