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
	Altitude  float64
	Bearing   float32
	Speed     float32
	Accuracy  float32
	Time      int64
	Provider  string
}

// ExtractLocation extracts fields from a Location JNI object.
func ExtractLocation(env *jni.Env, obj *jni.Object) (*ExtractedLocation, error) {
	if err := ensureInit(env); err != nil {
		return nil, err
	}

	cls, err := env.FindClass("android/location/Location")
	if err != nil {
		return nil, fmt.Errorf("find Location: %w", err)
	}
	defer env.DeleteLocalRef(&cls.Object)

	getLatMid, err := env.GetMethodID(cls, "getLatitude", "()D")
	if err != nil {
		return nil, fmt.Errorf("get getLatitude: %w", err)
	}
	getLonMid, err := env.GetMethodID(cls, "getLongitude", "()D")
	if err != nil {
		return nil, fmt.Errorf("get getLongitude: %w", err)
	}
	getAltMid, err := env.GetMethodID(cls, "getAltitude", "()D")
	if err != nil {
		return nil, fmt.Errorf("get getAltitude: %w", err)
	}
	getBearingMid, err := env.GetMethodID(cls, "getBearing", "()F")
	if err != nil {
		return nil, fmt.Errorf("get getBearing: %w", err)
	}
	getSpeedMid, err := env.GetMethodID(cls, "getSpeed", "()F")
	if err != nil {
		return nil, fmt.Errorf("get getSpeed: %w", err)
	}
	getAccuracyMid, err := env.GetMethodID(cls, "getAccuracy", "()F")
	if err != nil {
		return nil, fmt.Errorf("get getAccuracy: %w", err)
	}
	getTimeMid, err := env.GetMethodID(cls, "getTime", "()J")
	if err != nil {
		return nil, fmt.Errorf("get getTime: %w", err)
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
	alt, err := env.CallDoubleMethod(obj, getAltMid)
	if err != nil {
		return nil, fmt.Errorf("call getAltitude: %w", err)
	}
	bearing, err := env.CallFloatMethod(obj, getBearingMid)
	if err != nil {
		return nil, fmt.Errorf("call getBearing: %w", err)
	}
	speed, err := env.CallFloatMethod(obj, getSpeedMid)
	if err != nil {
		return nil, fmt.Errorf("call getSpeed: %w", err)
	}
	accuracy, err := env.CallFloatMethod(obj, getAccuracyMid)
	if err != nil {
		return nil, fmt.Errorf("call getAccuracy: %w", err)
	}
	locTime, err := env.CallLongMethod(obj, getTimeMid)
	if err != nil {
		return nil, fmt.Errorf("call getTime: %w", err)
	}
	provObj, err := env.CallObjectMethod(obj, getProvMid)
	if err != nil {
		return nil, fmt.Errorf("call getProvider: %w", err)
	}
	provider := env.GoString((*jni.String)(unsafe.Pointer(provObj)))
	env.DeleteLocalRef(provObj)

	return &ExtractedLocation{
		Latitude:  lat,
		Longitude: lon,
		Altitude:  alt,
		Bearing:   bearing,
		Speed:     speed,
		Accuracy:  accuracy,
		Time:      locTime,
		Provider:  provider,
	}, nil
}
