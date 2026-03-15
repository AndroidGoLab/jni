package protogen

import (
	"testing"

	"github.com/AndroidGoLab/jni/tools/pkg/javagen"
)

func TestDetectStreamingPattern_LocationListener(t *testing.T) {
	cb := &javagen.MergedCallback{
		JavaInterface: "android.location.LocationListener",
		GoType:        "locationListener",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onLocationChanged", GoField: "OnLocation"},
			{JavaMethod: "onProviderEnabled", GoField: "OnProviderEnabled"},
			{JavaMethod: "onProviderDisabled", GoField: "OnProviderDisabled"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != ServerStreaming {
		t.Errorf("expected ServerStreaming for LocationListener, got %d", pattern)
	}
}

func TestDetectStreamingPattern_BluetoothGattCallback(t *testing.T) {
	cb := &javagen.MergedCallback{
		JavaInterface: "android.bluetooth.BluetoothGattCallback",
		GoType:        "gattCallback",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onConnectionStateChange", GoField: "OnConnectionStateChange"},
			{JavaMethod: "onServicesDiscovered", GoField: "OnServicesDiscovered"},
			{JavaMethod: "onCharacteristicRead", GoField: "OnCharacteristicRead"},
			{JavaMethod: "onCharacteristicWrite", GoField: "OnCharacteristicWrite"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != BidiStreaming {
		t.Errorf("expected BidiStreaming for BluetoothGattCallback, got %d", pattern)
	}
}

func TestDetectStreamingPattern_BluetoothGattServerCallback(t *testing.T) {
	cb := &javagen.MergedCallback{
		JavaInterface: "android.bluetooth.BluetoothGattServerCallback",
		GoType:        "gattServerCallback",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onConnectionStateChange", GoField: "OnConnectionStateChange"},
			{JavaMethod: "onCharacteristicReadRequest", GoField: "OnCharacteristicReadRequest"},
			{JavaMethod: "onCharacteristicWriteRequest", GoField: "OnCharacteristicWriteRequest"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != BidiStreaming {
		t.Errorf("expected BidiStreaming for BluetoothGattServerCallback, got %d", pattern)
	}
}

func TestDetectStreamingPattern_ReadWriteHeuristic(t *testing.T) {
	// A callback with 2+ methods containing read/write should be detected as bidi
	// even if not in the known list.
	cb := &javagen.MergedCallback{
		JavaInterface: "com.example.CustomStreamCallback",
		GoType:        "customStreamCallback",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onDataRead", GoField: "OnDataRead"},
			{JavaMethod: "onDataWrite", GoField: "OnDataWrite"},
			{JavaMethod: "onError", GoField: "OnError"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != BidiStreaming {
		t.Errorf("expected BidiStreaming for callback with read+write methods, got %d", pattern)
	}
}

func TestDetectStreamingPattern_SingleReadNotBidi(t *testing.T) {
	// Only 1 read/write method should not trigger bidi.
	cb := &javagen.MergedCallback{
		JavaInterface: "com.example.ReaderCallback",
		GoType:        "readerCallback",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onDataRead", GoField: "OnDataRead"},
			{JavaMethod: "onError", GoField: "OnError"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != ServerStreaming {
		t.Errorf("expected ServerStreaming for callback with only 1 read method, got %d", pattern)
	}
}

func TestDetectStreamingPattern_CaseInsensitive(t *testing.T) {
	cb := &javagen.MergedCallback{
		JavaInterface: "com.example.IOCallback",
		GoType:        "ioCallback",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onREADComplete", GoField: "OnReadComplete"},
			{JavaMethod: "onWRITEComplete", GoField: "OnWriteComplete"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != BidiStreaming {
		t.Errorf("expected BidiStreaming for case-insensitive read/write match, got %d", pattern)
	}
}

func TestDetectStreamingPattern_ScanCallback(t *testing.T) {
	cb := &javagen.MergedCallback{
		JavaInterface: "android.bluetooth.le.ScanCallback",
		GoType:        "scanCallback",
		Methods: []javagen.MergedCallbackMethod{
			{JavaMethod: "onScanResult", GoField: "OnScanResult"},
			{JavaMethod: "onScanFailed", GoField: "OnScanFailed"},
		},
	}

	pattern := DetectStreamingPattern(cb)
	if pattern != ServerStreaming {
		t.Errorf("expected ServerStreaming for ScanCallback, got %d", pattern)
	}
}
