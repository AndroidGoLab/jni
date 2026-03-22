# Telephony API

The `telephony` package wraps `android.telephony.TelephonyManager` for querying phone type, SIM state, network operator, and data connectivity.

## Basic Usage

```go
import "github.com/AndroidGoLab/jni/telephony"

mgr, err := telephony.NewManager(ctx)
if err != nil {
    return fmt.Errorf("telephony.NewManager: %w", err)
}
defer mgr.Close()

// Phone type
phoneType, _ := mgr.GetPhoneType()
// Returns: telephony.PhoneTypeNone, PhoneTypeGsm, PhoneTypeCdma, PhoneTypeSip

// SIM state
simState, _ := mgr.GetSimState0()
// Returns: telephony.SimStateReady, SimStateAbsent, etc.

// Network operator name (e.g., "T-Mobile")
operatorName, _ := mgr.GetNetworkOperatorName()

// MCC+MNC code (e.g., "310260")
operator, _ := mgr.GetNetworkOperator()

// Roaming status
roaming, _ := mgr.IsNetworkRoaming()

// Data connection state
dataState, _ := mgr.GetDataState()
// Returns: telephony.DataConnected, DataDisconnected, etc.
```

## Constants

### Phone Types

```go
telephony.PhoneTypeNone // 0 - no phone radio
telephony.PhoneTypeGsm  // 1 - GSM
telephony.PhoneTypeCdma // 2 - CDMA
telephony.PhoneTypeSip  // 3 - SIP
```

### SIM States

```go
telephony.SimStateUnknown       // 0
telephony.SimStateAbsent        // 1 - no SIM card
telephony.SimStatePinRequired   // 2
telephony.SimStatePukRequired   // 3
telephony.SimStateNetworkLocked // 4
telephony.SimStateReady         // 5 - SIM ready
telephony.SimStateNotReady      // 6
telephony.SimStatePermDisabled  // 7
telephony.SimStateCardIoError   // 8
telephony.SimStateCardRestricted // 9
```

### Data Connection States

```go
telephony.DataDisconnected       // 0
telephony.DataConnecting         // 1
telephony.DataConnected          // 2
telephony.DataSuspended          // 3
telephony.DataDisconnecting      // 4
telephony.DataHandoverInProgress // 5
telephony.DataUnknown            // 6
```

## Complete Example

```go
import (
    "fmt"
    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/app"
    "github.com/AndroidGoLab/jni/telephony"
)

func getTelephonyInfo(vm *jni.VM, activityRef *jni.GlobalRef) error {
    ctx, _ := app.ContextFromObject(vm, activityRef)
    defer ctx.Close()

    mgr, err := telephony.NewManager(ctx)
    if err != nil {
        return err
    }
    defer mgr.Close()

    phoneType, _ := mgr.GetPhoneType()
    switch int(phoneType) {
    case telephony.PhoneTypeGsm:
        fmt.Println("Phone: GSM")
    case telephony.PhoneTypeCdma:
        fmt.Println("Phone: CDMA")
    default:
        fmt.Println("Phone: none/other")
    }

    simState, _ := mgr.GetSimState0()
    if int(simState) == telephony.SimStateReady {
        operatorName, _ := mgr.GetNetworkOperatorName()
        operator, _ := mgr.GetNetworkOperator()
        roaming, _ := mgr.IsNetworkRoaming()
        fmt.Printf("Operator: %s (%s), roaming: %v\n", operatorName, operator, roaming)
    }

    dataState, _ := mgr.GetDataState()
    switch int(dataState) {
    case telephony.DataConnected:
        fmt.Println("Data: connected")
    case telephony.DataDisconnected:
        fmt.Println("Data: disconnected")
    default:
        fmt.Printf("Data: state %d\n", dataState)
    }

    return nil
}
```

## Device Identity (Deprecated)

`TelephonyManager.getDeviceId()` was deprecated in API 26 and returns null on Android 10+. For device identification, use:

- `android.provider.Settings.Secure.ANDROID_ID` via the settings provider
- Or `Build.getSerial()` (requires `READ_PHONE_STATE` permission)

```go
import "github.com/AndroidGoLab/jni/os/build"

// Read Build.getSerial() (requires READ_PHONE_STATE)
b := build.Build{VM: vm}
serial, err := b.GetSerial()
```

## Required Permissions

```xml
<!-- Basic telephony info (phone type, SIM state, operator) - no permission needed -->

<!-- For device identifiers -->
<uses-permission android:name="android.permission.READ_PHONE_STATE" />
```
