//go:build !android

package e2e_grpc_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mp4FtypSignature is the "ftyp" box marker that appears at byte offset 4
// in all ISO Base Media (MP4) files.
var mp4FtypSignature = []byte("ftyp")

// ---- Video Recording ----

func TestE2E_CameraRecord(t *testing.T) {
	skipIfNoEmulator(t)

	outFile := filepath.Join(t.TempDir(), "video.mp4")

	runLiveJnicli(t,
		"--timeout", "60s",
		"camera", "record",
		"-d", "1s",
		"-o", outFile,
	)

	info, err := os.Stat(outFile)
	require.NoError(t, err, "recorded file must exist")
	require.Greater(t, info.Size(), int64(10000),
		"recorded MP4 must be larger than 10 KB")

	header := readFileHeader(t, outFile, 8)
	assert.Contains(t, string(header), string(mp4FtypSignature),
		"file header must contain the MP4 'ftyp' signature")

	t.Logf("recorded %s (%d bytes)", outFile, info.Size())
}

// ---- Audio Recording ----

func TestE2E_MicrophoneRecord(t *testing.T) {
	skipIfNoEmulator(t)

	outFile := filepath.Join(t.TempDir(), "audio.m4a")

	runLiveJnicli(t,
		"--timeout", "60s",
		"microphone", "record",
		"-d", "1s",
		"-o", outFile,
	)

	info, err := os.Stat(outFile)
	require.NoError(t, err, "recorded file must exist")
	require.Greater(t, info.Size(), int64(1000),
		"recorded M4A must be larger than 1 KB")

	t.Logf("recorded %s (%d bytes)", outFile, info.Size())
}

// ---- GPS Location ----

func TestE2E_LocationGet(t *testing.T) {
	skipIfNoEmulator(t)

	out := runLiveJnicli(t, "location", "get")
	resp := parseJSON(t, out)

	for _, field := range []string{"provider", "latitude", "longitude", "altitude", "accuracy"} {
		_, ok := resp[field]
		assert.True(t, ok, "response must contain field %q", field)
	}

	lat := getFloat64Field(t, resp, "latitude")
	lon := getFloat64Field(t, resp, "longitude")
	assert.NotZero(t, lat, "latitude must not be zero")
	assert.NotZero(t, lon, "longitude must not be zero")

	t.Logf("location: lat=%.6f lon=%.6f", lat, lon)
}

// ---- Battery ----

func TestE2E_BatteryTemperature(t *testing.T) {
	skipIfNoEmulator(t)

	// Per the Android source (BatteryManager.java), getIntProperty constants:
	//   BATTERY_PROPERTY_CAPACITY = 4  (battery level 0-100%)
	//   BATTERY_PROPERTY_STATUS   = 6  (charging status)
	// There is no BATTERY_PROPERTY_TEMPERATURE; temperature is only
	// available via Intent extras, not via getIntProperty.
	// We test BATTERY_PROPERTY_CAPACITY (4) which reliably returns 0-100.
	out := runLiveJnicli(t, "battery", "battery-manager", "get-int-property", "--arg0", "4")
	resp := parseJSON(t, out)

	capacity := getInt64Field(t, resp, "result")
	assert.Greater(t, capacity, int64(0),
		"battery capacity must be positive")
	assert.LessOrEqual(t, capacity, int64(100),
		"battery capacity must be at most 100")

	t.Logf("battery capacity: %d%%", capacity)
}

// ---- Helpers ----

// readFileHeader reads the first n bytes of a file.
func readFileHeader(t *testing.T, path string, n int) []byte {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	buf := make([]byte, n)
	nRead, err := f.Read(buf)
	require.NoError(t, err)

	return buf[:nRead]
}

// getFloat64Field extracts a float64 field from a JSON response.
func getFloat64Field(t *testing.T, resp map[string]any, field string) float64 {
	t.Helper()

	v, ok := resp[field]
	if !ok {
		t.Fatalf("missing field %q in response: %v", field, resp)
	}

	switch val := v.(type) {
	case float64:
		return val
	default:
		t.Fatalf("field %q: expected float64, got %T: %v", field, v, v)
		return 0
	}
}
