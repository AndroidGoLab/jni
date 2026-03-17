package location

import (
	"fmt"
	"unsafe"

	"github.com/AndroidGoLab/jni"
)

// ExtractedLocation holds extracted fields from android.location.Location.
type ExtractedLocation struct {
	Latitude  float64
	Longitude float64
	Provider  string
}

// ExtractLocation extracts latitude, longitude, and provider from a Location JNI object.
func ExtractLocation(env *jni.Env, obj *jni.Object) (*ExtractedLocation, error) {
	if err := ensureInit(env); err != nil {
		return nil, err
	}

	cls, err := env.FindClass("android/location/Location")
	if err != nil {
		return nil, fmt.Errorf("find Location: %w", err)
	}

	getLatMid, err := env.GetMethodID(cls, "getLatitude", "()D")
	if err != nil {
		return nil, fmt.Errorf("get getLatitude: %w", err)
	}
	getLonMid, err := env.GetMethodID(cls, "getLongitude", "()D")
	if err != nil {
		return nil, fmt.Errorf("get getLongitude: %w", err)
	}
	getProvMid, err := env.GetMethodID(cls, "getProvider", "()Ljava/lang/String;")
	if err != nil {
		return nil, fmt.Errorf("get getProvider: %w", err)
	}

	lat, err := env.CallDoubleMethod(obj, getLatMid)
	if err != nil {
		return nil, fmt.Errorf("call getLatitude: %w", err)
	}
	lon, err := env.CallDoubleMethod(obj, getLonMid)
	if err != nil {
		return nil, fmt.Errorf("call getLongitude: %w", err)
	}
	provObj, err := env.CallObjectMethod(obj, getProvMid)
	if err != nil {
		return nil, fmt.Errorf("call getProvider: %w", err)
	}
	provider := env.GoString((*jni.String)(unsafe.Pointer(provObj)))

	return &ExtractedLocation{
		Latitude:  lat,
		Longitude: lon,
		Provider:  provider,
	}, nil
}
