package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	pb "github.com/xaionaro-go/jni/proto/jni_raw"
)

var cameraPhotoCmd = &cobra.Command{
	Use:   "photo",
	Short: "Take a JPEG photo",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		cameraIndex, _ := cmd.Flags().GetInt("index")
		output, _ := cmd.Flags().GetString("output")
		client := pb.NewJNIServiceClient(grpcConn)

		cls, err := client.FindClass(ctx, &pb.FindClassRequest{
			Name: "center/dx/jni/jniservice/CameraCapture",
		})
		if err != nil {
			return fmt.Errorf("finding CameraCapture class (is the APK installed?): %w", err)
		}

		method, err := client.GetStaticMethodID(ctx, &pb.GetStaticMethodIDRequest{
			ClassHandle: cls.GetClassHandle(),
			Name:        "takePicture",
			Sig:         "(Landroid/content/Context;I)[B",
		})
		if err != nil {
			return fmt.Errorf("getting takePicture method: %w", err)
		}

		appCtx, err := client.GetAppContext(ctx, &pb.GetAppContextRequest{})
		if err != nil {
			return fmt.Errorf("getting app context: %w", err)
		}
		contextHandle := appCtx.GetContextHandle()
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

		data, err := client.GetByteArrayData(ctx, &pb.GetByteArrayDataRequest{
			ArrayHandle: arrayHandle,
		})
		if err != nil {
			return fmt.Errorf("reading image data: %w", err)
		}

		if output == "-" || output == "" {
			_, err := os.Stdout.Write(data.GetData())
			return err
		}
		if err := os.WriteFile(output, data.GetData(), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", output, err)
		}
		fmt.Fprintf(os.Stderr, "Saved %d bytes to %s\n", len(data.GetData()), output)
		return nil
	},
}

func init() {
	cameraPhotoCmd.Flags().Int("index", 0, "camera index (0=back, 1=front)")
	cameraPhotoCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
	cameraCmd.AddCommand(cameraPhotoCmd)
}
