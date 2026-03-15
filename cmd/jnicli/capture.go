package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	pb "github.com/xaionaro-go/jni/proto/jni_raw"
)

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "High-level capture commands (camera, location)",
}

var captureCameraCmd = &cobra.Command{
	Use:   "camera",
	Short: "Take a photo and save as JPEG",
	Long:  "Captures a JPEG photo from the device camera using the CameraCapture helper class.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		cameraIndex, _ := cmd.Flags().GetInt("camera")
		output, _ := cmd.Flags().GetString("output")
		client := pb.NewJNIServiceClient(grpcConn)

		// Find CameraCapture class.
		cls, err := client.FindClass(ctx, &pb.FindClassRequest{
			Name: "center/dx/jni/jniservice/CameraCapture",
		})
		if err != nil {
			return fmt.Errorf("finding CameraCapture class (is the APK installed?): %w", err)
		}

		// Get takePicture method.
		method, err := client.GetStaticMethodID(ctx, &pb.GetStaticMethodIDRequest{
			ClassHandle: cls.GetClassHandle(),
			Name:        "takePicture",
			Sig:         "(Landroid/content/Context;I)[B",
		})
		if err != nil {
			return fmt.Errorf("getting takePicture method: %w", err)
		}

		// The app context is typically handle 1 (from initAndroidContext) or
		// handle 2 (from setAppContext in APK mode). Try handle 2 first.
		contextHandle := int64(2)

		// Call takePicture(context, cameraIndex).
		result, err := client.CallStaticMethod(ctx, &pb.CallStaticMethodRequest{
			ClassHandle: cls.GetClassHandle(),
			MethodId:    method.GetMethodId(),
			ReturnType:  pb.JType_OBJECT,
			Args: []*pb.JValue{
				{Value: &pb.JValue_L{L: contextHandle}},
				{Value: &pb.JValue_I{I: int32(cameraIndex)}},
			},
		})
		if err != nil {
			return fmt.Errorf("camera capture failed: %w", err)
		}

		arrayHandle := result.GetResult().GetL()
		if arrayHandle == 0 {
			return fmt.Errorf("camera returned null (check camera permission)")
		}

		// Get the JPEG bytes.
		data, err := client.GetByteArrayData(ctx, &pb.GetByteArrayDataRequest{
			ArrayHandle: arrayHandle,
		})
		if err != nil {
			return fmt.Errorf("reading image data: %w", err)
		}

		if err := os.WriteFile(output, data.GetData(), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", output, err)
		}
		fmt.Fprintf(os.Stderr, "Saved %d bytes to %s\n", len(data.GetData()), output)
		return nil
	},
}

var captureLocationCmd = &cobra.Command{
	Use:   "location",
	Short: "Get current GPS coordinates",
	Long:  "Returns the last known location from the device's location providers.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := pb.NewJNIServiceClient(grpcConn)
		contextHandle := int64(2) // APK app context

		// Get Context class + getSystemService method.
		ctxCls, err := client.FindClass(ctx, &pb.FindClassRequest{Name: "android/content/Context"})
		if err != nil {
			return fmt.Errorf("finding Context class: %w", err)
		}
		gssMID, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: ctxCls.GetClassHandle(),
			Name:        "getSystemService",
			Sig:         "(Ljava/lang/String;)Ljava/lang/Object;",
		})
		if err != nil {
			return fmt.Errorf("getting getSystemService: %w", err)
		}

		// Create "location" string.
		locStr, err := client.NewStringUTF(ctx, &pb.NewStringUTFRequest{Value: "location"})
		if err != nil {
			return err
		}

		// Get LocationManager.
		lmResult, err := client.CallMethod(ctx, &pb.CallMethodRequest{
			ObjectHandle: contextHandle,
			MethodId:     gssMID.GetMethodId(),
			ReturnType:   pb.JType_OBJECT,
			Args:         []*pb.JValue{{Value: &pb.JValue_L{L: locStr.GetStringHandle()}}},
		})
		if err != nil {
			return fmt.Errorf("getSystemService(location): %w", err)
		}
		lmHandle := lmResult.GetResult().GetL()
		if lmHandle == 0 {
			return fmt.Errorf("LocationManager is null")
		}

		// Get LocationManager class + getLastKnownLocation.
		lmCls, err := client.FindClass(ctx, &pb.FindClassRequest{Name: "android/location/LocationManager"})
		if err != nil {
			return err
		}
		glklMID, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: lmCls.GetClassHandle(),
			Name:        "getLastKnownLocation",
			Sig:         "(Ljava/lang/String;)Landroid/location/Location;",
		})
		if err != nil {
			return err
		}

		// Location class methods.
		locCls, err := client.FindClass(ctx, &pb.FindClassRequest{Name: "android/location/Location"})
		if err != nil {
			return err
		}
		getLatMID, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getLatitude", Sig: "()D",
		})
		if err != nil {
			return err
		}
		getLngMID, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getLongitude", Sig: "()D",
		})
		if err != nil {
			return err
		}
		getAccMID, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getAccuracy", Sig: "()F",
		})
		if err != nil {
			return err
		}
		getAltMID, err := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getAltitude", Sig: "()D",
		})
		if err != nil {
			return err
		}

		// Try providers in order.
		providers := []string{"gps", "network", "fused", "passive"}
		for _, provider := range providers {
			provStr, err := client.NewStringUTF(ctx, &pb.NewStringUTFRequest{Value: provider})
			if err != nil {
				continue
			}
			locResult, err := client.CallMethod(ctx, &pb.CallMethodRequest{
				ObjectHandle: lmHandle,
				MethodId:     glklMID.GetMethodId(),
				ReturnType:   pb.JType_OBJECT,
				Args:         []*pb.JValue{{Value: &pb.JValue_L{L: provStr.GetStringHandle()}}},
			})
			if err != nil {
				continue
			}
			locHandle := locResult.GetResult().GetL()
			if locHandle == 0 {
				continue
			}

			lat, _ := client.CallMethod(ctx, &pb.CallMethodRequest{
				ObjectHandle: locHandle, MethodId: getLatMID.GetMethodId(), ReturnType: pb.JType_DOUBLE,
			})
			lng, _ := client.CallMethod(ctx, &pb.CallMethodRequest{
				ObjectHandle: locHandle, MethodId: getLngMID.GetMethodId(), ReturnType: pb.JType_DOUBLE,
			})
			acc, _ := client.CallMethod(ctx, &pb.CallMethodRequest{
				ObjectHandle: locHandle, MethodId: getAccMID.GetMethodId(), ReturnType: pb.JType_FLOAT,
			})
			alt, _ := client.CallMethod(ctx, &pb.CallMethodRequest{
				ObjectHandle: locHandle, MethodId: getAltMID.GetMethodId(), ReturnType: pb.JType_DOUBLE,
			})

			return printResult(map[string]any{
				"provider":  provider,
				"latitude":  lat.GetResult().GetD(),
				"longitude": lng.GetResult().GetD(),
				"altitude":  alt.GetResult().GetD(),
				"accuracy":  acc.GetResult().GetF(),
			})
		}
		return fmt.Errorf("no location available from any provider")
	},
}

var captureDeviceInfoCmd = &cobra.Command{
	Use:   "device-info",
	Short: "Get device model, manufacturer, SDK version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := pb.NewJNIServiceClient(grpcConn)

		// Build class has static String fields.
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

		// Build.VERSION has SDK_INT.
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

		return printResult(map[string]any{
			"manufacturer": getStringField("MANUFACTURER"),
			"model":        getStringField("MODEL"),
			"brand":        getStringField("BRAND"),
			"device":       getStringField("DEVICE"),
			"product":      getStringField("PRODUCT"),
			"sdk_int":      strconv.Itoa(int(sdkVal.GetResult().GetI())),
			"release":      func() string {
				fid, _ := client.GetStaticFieldID(ctx, &pb.GetStaticFieldIDRequest{
					ClassHandle: verCls.GetClassHandle(), Name: "RELEASE", Sig: "Ljava/lang/String;",
				})
				val, _ := client.GetStaticField(ctx, &pb.GetStaticFieldValueRequest{
					ClassHandle: verCls.GetClassHandle(), FieldId: fid.GetFieldId(), FieldType: pb.JType_OBJECT,
				})
				str, _ := client.GetStringUTFChars(ctx, &pb.GetStringUTFCharsRequest{StringHandle: val.GetResult().GetL()})
				return str.GetValue()
			}(),
		})
	},
}

func init() {
	captureCameraCmd.Flags().Int("camera", 0, "camera index (0=back, 1=front)")
	captureCameraCmd.Flags().StringP("output", "o", "capture.jpg", "output JPEG file path")

	captureCmd.AddCommand(captureCameraCmd, captureLocationCmd, captureDeviceInfoCmd)
	rootCmd.AddCommand(captureCmd)
}
