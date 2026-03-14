package specgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// PackageMapping defines how a Java package maps to a Go package.
type PackageMapping struct {
	JavaPrefix string // e.g. "android.app.admin"
	Package    string // e.g. "admin"
	GoImport   string // e.g. "github.com/xaionaro-go/jni/app/admin"
}

// AndroidServiceName maps known manager class names to their Android
// Context.getSystemService() constant names.
var AndroidServiceName = map[string]string{
	"android.app.AlarmManager":                        "alarm",
	"android.app.KeyguardManager":                     "keyguard",
	"android.app.NotificationManager":                 "notification",
	"android.app.admin.DevicePolicyManager":            "device_policy",
	"android.app.blob.BlobStoreManager":                "blob_store",
	"android.app.job.JobScheduler":                     "jobscheduler",
	"android.app.role.RoleManager":                     "role",
	"android.app.usage.UsageStatsManager":              "usagestats",
	"android.bluetooth.BluetoothManager":               "bluetooth",
	"android.companion.CompanionDeviceManager":         "companiondevice",
	"android.content.ClipboardManager":                 "clipboard",
	"android.hardware.ConsumerIrManager":               "consumer_ir",
	"android.hardware.camera2.CameraManager":           "camera",
	"android.hardware.lights.LightsManager":            "lights",
	"android.location.LocationManager":                 "location",
	"android.media.AudioManager":                       "audio",
	"android.media.RingtoneManager":                    "",
	"android.media.projection.MediaProjectionManager":  "media_projection",
	"android.media.session.MediaSessionManager":        "media_session",
	"android.net.ConnectivityManager":                  "connectivity",
	"android.net.wifi.WifiManager":                     "wifi",
	"android.net.wifi.p2p.WifiP2pManager":              "wifip2p",
	"android.net.wifi.rtt.WifiRttManager":              "wifirtt",
	"android.nfc.NfcManager":                           "nfc",
	"android.os.BatteryManager":                        "batterymanager",
	"android.os.PowerManager":                          "power",
	"android.os.Vibrator":                              "vibrator",
	"android.os.storage.StorageManager":                "storage",
	"android.print.PrintManager":                       "print",
	"android.se.omapi.SEService":                       "",
	"android.telecom.TelecomManager":                   "telecom",
	"android.telephony.TelephonyManager":               "phone",
	"android.view.WindowManager":                       "window",
	"android.view.inputmethod.InputMethodManager":      "input_method",
}

// GenerateSpec generates a YAML spec from .class files in a directory
// by running javap on each class.
func GenerateSpec(
	classPath string,
	className string,
	pkgMapping PackageMapping,
	goModule string,
) (*SpecFile, error) {
	jc, err := RunJavap(classPath, className)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", className, err)
	}

	spec := &SpecFile{
		Package:  pkgMapping.Package,
		GoImport: pkgMapping.GoImport,
	}

	cls := classFromJavap(jc)
	spec.Classes = append(spec.Classes, cls)

	return spec, nil
}

// GenerateFromRefDir scans ref/ for .class files and generates one YAML
// spec per top-level class (inner classes are grouped with their parent).
// extraClassPath is appended to the javap -cp argument.
func GenerateFromRefDir(
	refDir string,
	extraClassPath string,
	outputDir string,
	goModule string,
) error {
	var classFiles []string
	err := filepath.Walk(refDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".class") {
			classFiles = append(classFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", refDir, err)
	}

	// Separate top-level classes from inner classes.
	type classEntry struct {
		className string
		filePath  string
	}
	topLevel := make(map[string]classEntry)       // parent class → entry
	innerClasses := make(map[string][]classEntry) // parent class → inner entries

	for _, cf := range classFiles {
		rel, _ := filepath.Rel(refDir, cf)
		className := strings.TrimSuffix(rel, ".class")
		className = strings.ReplaceAll(className, "/", ".")

		entry := classEntry{className: className, filePath: cf}

		if strings.Contains(filepath.Base(cf), "$") {
			// Inner class — group with parent.
			parent := className[:strings.LastIndex(className, "$")]
			innerClasses[parent] = append(innerClasses[parent], entry)
		} else {
			topLevel[className] = entry
		}
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}

	cp := refDir
	if extraClassPath != "" {
		cp = refDir + ":" + extraClassPath
	}

	for parentName, entry := range topLevel {
		mapping := inferClassMapping(parentName, goModule)

		spec := &SpecFile{
			Package:  mapping.Package,
			GoImport: mapping.GoImport,
		}

		// Parse the top-level class.
		jc, err := RunJavap(cp, entry.className)
		if err != nil {
			return fmt.Errorf("javap %s: %w", entry.className, err)
		}
		cls := classFromJavap(jc)
		spec.Classes = append(spec.Classes, cls)
		addConstants(spec, jc)

		// Parse inner classes.
		for _, inner := range innerClasses[parentName] {
			ijc, err := RunJavap(cp, inner.className)
			if err != nil {
				return fmt.Errorf("javap %s: %w", inner.className, err)
			}
			icls := classFromJavap(ijc)
			spec.Classes = append(spec.Classes, icls)
			addConstants(spec, ijc)
		}

		outPath := filepath.Join(outputDir, mapping.Package+".yaml")
		if err := writeSpecFile(spec, outPath); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

func addConstants(spec *SpecFile, jc *JavapClass) {
	for _, c := range jc.Constants {
		spec.Constants = append(spec.Constants, SpecConstant{
			GoName: javaConstantToGoName(c.Name),
			Value:  inferConstantDefault(c.JavaType),
			GoType: javaTypeToGoType(c.JavaType),
		})
	}
}

func classFromJavap(jc *JavapClass) SpecClass {
	cls := SpecClass{
		JavaClass: jc.FullName,
		GoType:    inferGoType(jc.FullName),
	}

	// Determine obtain type.
	if svcName, ok := AndroidServiceName[jc.FullName]; ok && svcName != "" {
		cls.Obtain = "system_service"
		cls.ServiceName = svcName
		cls.Close = true
	}

	// Count method names to detect overloads.
	nameCounts := make(map[string]int)
	for _, m := range jc.Methods {
		if hasUnsupportedParams(m) {
			continue
		}
		nameCounts[m.Name]++
	}

	// Track per-name occurrence index for disambiguation.
	nameIndex := make(map[string]int)

	for _, m := range jc.Methods {
		if hasUnsupportedParams(m) {
			continue
		}

		sm := specMethodFromJavap(m)

		// Disambiguate overloaded methods by appending parameter count.
		if nameCounts[m.Name] > 1 {
			idx := nameIndex[m.Name]
			nameIndex[m.Name] = idx + 1
			suffix := fmt.Sprintf("%d", len(m.Params))
			if idx > 0 {
				suffix = fmt.Sprintf("%d_%d", len(m.Params), idx)
			}
			sm.GoName = sm.GoName + suffix
		}

		switch {
		case m.IsStatic:
			cls.StaticMethods = append(cls.StaticMethods, sm)
		default:
			cls.Methods = append(cls.Methods, sm)
		}
	}

	return cls
}

func specMethodFromJavap(m JavapMethod) SpecMethod {
	sm := SpecMethod{
		JavaMethod: m.Name,
		GoName:     javaMethodToGoName(m.Name),
		Returns:    javaTypeToSpecType(m.ReturnType),
		Error:      true,
	}

	for i, p := range m.Params {
		sm.Params = append(sm.Params, SpecParam{
			JavaType: javaTypeToSpecType(p.JavaType),
			GoName:   fmt.Sprintf("arg%d", i),
		})
	}

	return sm
}

// hasUnsupportedParams checks if a method has parameter types that can't
// be represented in the YAML spec (byte buffers, handlers, complex generics).
func hasUnsupportedParams(m JavapMethod) bool {
	for _, p := range m.Params {
		switch {
		case strings.Contains(p.JavaType, "ByteBuffer"):
			return true
		case strings.Contains(p.JavaType, "Handler"):
			return true
		case strings.Contains(p.JavaType, "[]"):
			// Array params are fine for primitives but complex for objects.
		}
	}
	return false
}

// inferClassMapping derives the Go package name from a single Java class name.
// E.g. "android.app.AlarmManager" → package "alarm", go_import ".../app/alarm".
func inferClassMapping(className string, goModule string) PackageMapping {
	// Known class → package mappings (primary service classes).
	knownMappings := map[string]struct{ pkg, goPath string }{
		"android.accounts.AccountManager":                   {"accounts", "accounts"},
		"android.app.Activity":                              {"app", "app"},
		"android.app.AlarmManager":                          {"alarm", "app/alarm"},
		"android.app.DownloadManager":                       {"download", "app/download"},
		"android.app.KeyguardManager":                       {"keyguard", "os/keyguard"},
		"android.app.NotificationChannel":                   {"notification", "app/notification"},
		"android.app.NotificationManager":                   {"notification", "app/notification"},
		"android.app.PendingIntent":                         {"app", "app"},
		"android.app.admin.DevicePolicyManager":             {"admin", "app/admin"},
		"android.app.blob.BlobStoreManager":                 {"blob", "app/blob"},
		"android.app.job.JobInfo":                           {"job", "app/job"},
		"android.app.job.JobScheduler":                      {"job", "app/job"},
		"android.app.role.RoleManager":                      {"role", "app/role"},
		"android.app.usage.UsageStats":                      {"usage", "app/usage"},
		"android.app.usage.UsageStatsManager":               {"usage", "app/usage"},
		"android.bluetooth.BluetoothAdapter":                {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothDevice":                 {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothGatt":                   {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothGattCharacteristic":     {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothGattDescriptor":         {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothGattServer":             {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothGattService":            {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothServerSocket":           {"bluetooth", "bluetooth"},
		"android.bluetooth.BluetoothSocket":                 {"bluetooth", "bluetooth"},
		"android.bluetooth.le.BluetoothLeAdvertiser":        {"bluetooth_le", "bluetooth/le"},
		"android.bluetooth.le.BluetoothLeScanner":           {"bluetooth_le", "bluetooth/le"},
		"android.bluetooth.le.ScanFilter":                   {"bluetooth_le", "bluetooth/le"},
		"android.bluetooth.le.ScanResult":                   {"bluetooth_le", "bluetooth/le"},
		"android.bluetooth.le.ScanSettings":                 {"bluetooth_le", "bluetooth/le"},
		"android.bluetooth.le.AdvertiseData":                {"bluetooth_le", "bluetooth/le"},
		"android.bluetooth.le.AdvertiseSettings":            {"bluetooth_le", "bluetooth/le"},
		"android.companion.AssociationRequest":              {"companion", "companion"},
		"android.companion.CompanionDeviceManager":          {"companion", "companion"},
		"android.content.BroadcastReceiver":                 {"content", "content"},
		"android.content.ClipData":                          {"clipboard", "content/clipboard"},
		"android.content.ClipboardManager":                  {"clipboard", "content/clipboard"},
		"android.content.ContentResolver":                   {"resolver", "content/resolver"},
		"android.content.Context":                           {"app", "app"},
		"android.content.Intent":                            {"app", "app"},
		"android.content.SharedPreferences":                 {"preferences", "content/preferences"},
		"android.content.pm.PackageInfo":                    {"pm", "content/pm"},
		"android.content.pm.PackageManager":                 {"pm", "content/pm"},
		"android.database.Cursor":                           {"resolver", "content/resolver"},
		"android.graphics.Bitmap":                           {"pdf", "graphics/pdf"},
		"android.graphics.pdf.PdfRenderer":                  {"pdf", "graphics/pdf"},
		"android.hardware.ConsumerIrManager":                {"ir", "hardware/ir"},
		"android.hardware.biometrics.BiometricManager":      {"biometric", "hardware/biometric"},
		"android.hardware.biometrics.BiometricPrompt":       {"biometric", "hardware/biometric"},
		"android.hardware.camera2.CameraManager":            {"camera", "hardware/camera"},
		"android.hardware.display.VirtualDisplay":           {"projection", "media/projection"},
		"android.hardware.lights.Light":                     {"lights", "hardware/lights"},
		"android.hardware.lights.LightState":                {"lights", "hardware/lights"},
		"android.hardware.lights.LightsManager":             {"lights", "hardware/lights"},
		"android.hardware.lights.LightsRequest":             {"lights", "hardware/lights"},
		"android.hardware.usb.UsbDevice":                    {"usb", "hardware/usb"},
		"android.hardware.usb.UsbDeviceConnection":          {"usb", "hardware/usb"},
		"android.hardware.usb.UsbEndpoint":                  {"usb", "hardware/usb"},
		"android.hardware.usb.UsbInterface":                 {"usb", "hardware/usb"},
		"android.hardware.usb.UsbManager":                   {"usb", "hardware/usb"},
		"android.location.GnssStatus":                       {"location", "location"},
		"android.location.Location":                         {"location", "location"},
		"android.location.LocationManager":                  {"location", "location"},
		"android.location.altitude.AltitudeConverter":       {"altitude", "location/altitude"},
		"android.media.AudioDeviceInfo":                     {"audiomanager", "media/audiomanager"},
		"android.media.AudioFocusRequest":                   {"audiomanager", "media/audiomanager"},
		"android.media.AudioManager":                        {"audiomanager", "media/audiomanager"},
		"android.media.AudioRecord":                         {"audiorecord", "media/audiorecord"},
		"android.media.MediaPlayer":                         {"player", "media/player"},
		"android.media.MediaRecorder":                       {"recorder", "media/recorder"},
		"android.media.RingtoneManager":                     {"ringtone", "media/ringtone"},
		"android.media.projection.MediaProjection":          {"projection", "media/projection"},
		"android.media.projection.MediaProjectionManager":   {"projection", "media/projection"},
		"android.media.session.MediaController":             {"session", "media/session"},
		"android.media.session.MediaSessionManager":         {"session", "media/session"},
		"android.net.ConnectivityManager":                   {"net", "net"},
		"android.net.NetworkCapabilities":                   {"net", "net"},
		"android.net.Uri":                                   {"resolver", "content/resolver"},
		"android.net.VpnService":                            {"vpn", "net/vpn"},
		"android.net.nsd.NsdManager":                        {"nsd", "net/nsd"},
		"android.net.nsd.NsdServiceInfo":                    {"nsd", "net/nsd"},
		"android.net.wifi.ScanResult":                       {"wifi", "net/wifi"},
		"android.net.wifi.WifiInfo":                         {"wifi", "net/wifi"},
		"android.net.wifi.WifiManager":                      {"wifi", "net/wifi"},
		"android.net.wifi.p2p.WifiP2pConfig":                {"wifi_p2p", "net/wifi/p2p"},
		"android.net.wifi.p2p.WifiP2pDevice":                {"wifi_p2p", "net/wifi/p2p"},
		"android.net.wifi.p2p.WifiP2pGroup":                 {"wifi_p2p", "net/wifi/p2p"},
		"android.net.wifi.p2p.WifiP2pManager":               {"wifi_p2p", "net/wifi/p2p"},
		"android.net.wifi.rtt.RangingRequest":               {"wifi_rtt", "net/wifi/rtt"},
		"android.net.wifi.rtt.RangingResult":                {"wifi_rtt", "net/wifi/rtt"},
		"android.net.wifi.rtt.WifiRttManager":               {"wifi_rtt", "net/wifi/rtt"},
		"android.nfc.NdefMessage":                           {"nfc", "nfc"},
		"android.nfc.NdefRecord":                            {"nfc", "nfc"},
		"android.nfc.NfcAdapter":                            {"nfc", "nfc"},
		"android.nfc.Tag":                                   {"nfc", "nfc"},
		"android.nfc.tech.IsoDep":                           {"nfc", "nfc"},
		"android.nfc.tech.Ndef":                             {"nfc", "nfc"},
		"android.os.BatteryManager":                         {"battery", "os/battery"},
		"android.os.Build":                                  {"build", "os/build"},
		"android.os.Bundle":                                 {"app", "app"},
		"android.os.CancellationSignal":                     {"app", "app"},
		"android.os.Environment":                            {"environment", "os/environment"},
		"android.os.ParcelFileDescriptor":                   {"pdf", "graphics/pdf"},
		"android.os.PowerManager":                           {"power", "os/power"},
		"android.os.Vibrator":                               {"vibrator", "os/vibrator"},
		"android.os.storage.StorageManager":                 {"storage", "os/storage"},
		"android.os.storage.StorageVolume":                  {"storage", "os/storage"},
		"android.print.PrintJob":                            {"print", "print"},
		"android.print.PrintJobInfo":                        {"print", "print"},
		"android.print.PrintManager":                        {"print", "print"},
		"android.provider.CalendarContract":                 {"calendar", "provider/calendar"},
		"android.provider.ContactsContract":                 {"contacts", "provider/contacts"},
		"android.provider.DocumentsContract":                {"documents", "provider/documents"},
		"android.provider.MediaStore":                       {"mediastore", "provider/media"},
		"android.provider.Settings":                         {"settings", "provider/settings"},
		"android.se.omapi.Channel":                          {"omapi", "se/omapi"},
		"android.se.omapi.Reader":                           {"omapi", "se/omapi"},
		"android.se.omapi.SEService":                        {"omapi", "se/omapi"},
		"android.se.omapi.Session":                          {"omapi", "se/omapi"},
		"android.security.keystore.KeyGenParameterSpec":     {"keystore", "security/keystore"},
		"android.service.notification.StatusBarNotification": {"notification", "app/notification"},
		"android.speech.SpeechRecognizer":                   {"speech", "speech"},
		"android.speech.tts.TextToSpeech":                   {"speech", "speech"},
		"android.telecom.TelecomManager":                    {"telecom", "telecom"},
		"android.telephony.TelephonyManager":                {"telephony", "telephony"},
		"android.util.DisplayMetrics":                       {"display", "view/display"},
		"android.view.Display":                              {"display", "view/display"},
		"android.view.WindowManager":                        {"display", "view/display"},
		"android.view.inputmethod.InputMethodManager":       {"inputmethod", "view/inputmethod"},
		"android.widget.Toast":                              {"toast", "widget/toast"},
	}

	if m, ok := knownMappings[className]; ok {
		return PackageMapping{
			Package:  m.pkg,
			GoImport: goModule + "/" + m.goPath,
		}
	}

	return inferPackageMapping(className, goModule)
}

func inferPackageMapping(className string, goModule string) PackageMapping {
	// Map known Android package prefixes to Go packages.
	mappings := []struct {
		prefix  string
		pkg     string
		goPath  string
	}{
		{"android.app.admin.", "admin", "app/admin"},
		{"android.app.blob.", "blob", "app/blob"},
		{"android.app.role.", "role", "app/role"},
		{"android.app.job.", "job", "app/job"},
		{"android.app.usage.", "usage", "app/usage"},
		{"android.app.", "app", "app"},
		{"android.content.", "content", "content"},
		{"android.hardware.camera2.", "camera", "hardware/camera"},
		{"android.hardware.lights.", "lights", "hardware/lights"},
		{"android.hardware.", "hardware", "hardware"},
		{"android.location.altitude.", "altitude", "location/altitude"},
		{"android.location.", "location", "location"},
		{"android.media.session.", "session", "media/session"},
		{"android.media.", "media", "media"},
		{"android.net.wifi.p2p.", "wifi_p2p", "net/wifi/p2p"},
		{"android.net.wifi.rtt.", "wifi_rtt", "net/wifi/rtt"},
		{"android.net.wifi.", "wifi", "net/wifi"},
		{"android.net.", "net", "net"},
		{"android.nfc.", "nfc", "nfc"},
		{"android.os.storage.", "storage", "os/storage"},
		{"android.os.", "os", "os"},
		{"android.provider.", "provider", "provider"},
		{"android.se.omapi.", "omapi", "se/omapi"},
		{"android.telecom.", "telecom", "telecom"},
		{"android.telephony.", "telephony", "telephony"},
		{"android.view.inputmethod.", "inputmethod", "view/inputmethod"},
		{"android.view.", "display", "view/display"},
	}

	for _, m := range mappings {
		if strings.HasPrefix(className, m.prefix) {
			return PackageMapping{
				JavaPrefix: m.prefix,
				Package:    m.pkg,
				GoImport:   goModule + "/" + m.goPath,
			}
		}
	}

	// Fallback: use last segment of the Java package.
	parts := strings.Split(className, ".")
	pkg := parts[len(parts)-2]
	return PackageMapping{
		JavaPrefix: strings.Join(parts[:len(parts)-1], ".") + ".",
		Package:    pkg,
		GoImport:   goModule + "/" + pkg,
	}
}

// ---- YAML spec types (for serialization) ----

// SpecFile is the YAML output structure.
type SpecFile struct {
	Package   string         `yaml:"package"`
	GoImport  string         `yaml:"go_import"`
	Classes   []SpecClass    `yaml:"classes"`
	Callbacks []SpecCallback `yaml:"callbacks,omitempty"`
	Constants []SpecConstant `yaml:"constants,omitempty"`
}

// SpecClass is a class in the YAML spec.
type SpecClass struct {
	JavaClass     string       `yaml:"java_class"`
	GoType        string       `yaml:"go_type"`
	Obtain        string       `yaml:"obtain,omitempty"`
	ServiceName   string       `yaml:"service_name,omitempty"`
	Kind          string       `yaml:"kind,omitempty"`
	Close         bool         `yaml:"close,omitempty"`
	Methods       []SpecMethod `yaml:"methods,omitempty"`
	StaticMethods []SpecMethod `yaml:"static_methods,omitempty"`
	Fields        []SpecField  `yaml:"fields,omitempty"`
}

// SpecMethod is a method in the YAML spec.
type SpecMethod struct {
	JavaMethod string      `yaml:"java_method"`
	GoName     string      `yaml:"go_name"`
	Params     []SpecParam `yaml:"params,omitempty"`
	Returns    string      `yaml:"returns"`
	Error      bool        `yaml:"error"`
}

// SpecParam is a method parameter in the YAML spec.
type SpecParam struct {
	JavaType string `yaml:"java_type"`
	GoName   string `yaml:"go_name"`
}

// SpecField is a data class field in the YAML spec.
type SpecField struct {
	JavaMethod string `yaml:"java_method"`
	Returns    string `yaml:"returns"`
	GoName     string `yaml:"go_name"`
	GoType     string `yaml:"go_type"`
}

// SpecCallback is a callback interface in the YAML spec.
type SpecCallback struct {
	JavaInterface string               `yaml:"java_interface"`
	GoType        string               `yaml:"go_type"`
	Methods       []SpecCallbackMethod `yaml:"methods"`
}

// SpecCallbackMethod is a callback method in the YAML spec.
type SpecCallbackMethod struct {
	JavaMethod string   `yaml:"java_method"`
	Params     []string `yaml:"params"`
	GoField    string   `yaml:"go_field"`
}

// SpecConstant is a constant in the YAML spec.
type SpecConstant struct {
	GoName string `yaml:"go_name"`
	Value  string `yaml:"value"`
	GoType string `yaml:"go_type"`
}

func writeSpecFile(spec *SpecFile, path string) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// ---- Name conversion helpers ----

// javaTypeToSpecType converts a fully-qualified Java type to the
// short form used in YAML specs.
func javaTypeToSpecType(jt string) string {
	switch jt {
	case "void":
		return "void"
	case "boolean":
		return "boolean"
	case "byte":
		return "byte"
	case "char":
		return "char"
	case "short":
		return "short"
	case "int":
		return "int"
	case "long":
		return "long"
	case "float":
		return "float"
	case "double":
		return "double"
	case "java.lang.String", "java.lang.CharSequence":
		return "String"
	case "byte[]":
		return "[B"
	case "int[]":
		return "[I"
	case "long[]":
		return "[J"
	default:
		return jt
	}
}

// javaTypeToGoType converts a Java type name to a Go type for constants.
func javaTypeToGoType(jt string) string {
	switch jt {
	case "int":
		return "int"
	case "long":
		return "int64"
	case "java.lang.String":
		return "string"
	case "boolean":
		return "bool"
	case "float":
		return "float32"
	case "double":
		return "float64"
	default:
		return "int"
	}
}

// javaMethodToGoName converts a Java method name (camelCase) to a Go
// exported name (PascalCase), with raw suffix for complex methods.
func javaMethodToGoName(name string) string {
	if len(name) == 0 {
		return name
	}
	goName := strings.ToUpper(name[:1]) + name[1:]

	// Append "Raw" suffix if the name starts with common patterns
	// indicating it returns a raw JNI object (convention in this project).
	return goName
}

// inferGoType determines the unexported Go type name for a Java class.
func inferGoType(fullClass string) string {
	parts := strings.Split(fullClass, ".")
	name := parts[len(parts)-1]

	// Handle inner classes: Foo$Bar → fooBar (include parent for uniqueness).
	if idx := strings.LastIndex(name, "$"); idx >= 0 {
		parent := name[:idx]
		child := name[idx+1:]
		name = parent + child
	}

	// Unexport it (lowercase first letter).
	return strings.ToLower(name[:1]) + name[1:]
}

// javaConstantToGoName converts SCREAMING_SNAKE_CASE to PascalCase.
func javaConstantToGoName(name string) string {
	parts := strings.Split(strings.ToLower(name), "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// inferConstantDefault returns a placeholder default for a constant.
// The actual values come from the Android SDK; we use 0/"" as placeholders.
func inferConstantDefault(javaType string) string {
	switch javaType {
	case "java.lang.String":
		return `""`
	default:
		return "0"
	}
}
