package main

import (
	"fmt"

	"github.com/spf13/cobra"
	appconsts "github.com/AndroidGoLab/jni/app/consts"
	locationconsts "github.com/AndroidGoLab/jni/location/consts"
	pb "github.com/AndroidGoLab/jni/proto/jni_raw"
)

var locationGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get last known GPS coordinates",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := pb.NewJNIServiceClient(grpcConn)
		appCtx, err := client.GetAppContext(ctx, &pb.GetAppContextRequest{})
		if err != nil {
			return fmt.Errorf("getting app context: %w", err)
		}
		contextHandle := appCtx.GetContextHandle()

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

		locStr, err := client.NewStringUTF(ctx, &pb.NewStringUTFRequest{Value: appconsts.LocationService})
		if err != nil {
			return err
		}

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

		locCls, err := client.FindClass(ctx, &pb.FindClassRequest{Name: "android/location/Location"})
		if err != nil {
			return err
		}
		getLatMID, _ := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getLatitude", Sig: "()D",
		})
		getLngMID, _ := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getLongitude", Sig: "()D",
		})
		getAccMID, _ := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getAccuracy", Sig: "()F",
		})
		getAltMID, _ := client.GetMethodID(ctx, &pb.GetMethodIDRequest{
			ClassHandle: locCls.GetClassHandle(), Name: "getAltitude", Sig: "()D",
		})

		providers := []string{
			locationconsts.GpsProvider,
			locationconsts.NetworkProvider,
			locationconsts.FusedProvider,
			locationconsts.PassiveProvider,
		}
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

func init() {
	locationCmd.AddCommand(locationGetCmd)
}
