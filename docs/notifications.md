# Notifications

The `app/notification` package wraps `android.app.NotificationManager` for posting notifications from Go with zero Java code.

## Post a Notification

```go
import "github.com/AndroidGoLab/jni/app/notification"

// 1. Create NotificationManager from app context
mgr, err := notification.NewManager(ctx)
if err != nil {
    return err
}
defer mgr.Close()

// 2. Create a notification channel (required on API 26+)
ch, err := notification.NewChannel(
    vm,
    "alerts",                              // channel ID
    "Alerts",                              // display name
    int32(notification.ImportanceHigh),     // importance level
)
if err != nil {
    return err
}
if err := mgr.CreateNotificationChannel(ch.Obj); err != nil {
    return err
}

// 3. Get the application context for the builder
appCtxObj, err := ctx.GetApplicationContext()
if err != nil {
    return err
}

// 4. Build the notification
b, err := notification.NewBuilder(vm, appCtxObj, "alerts")
if err != nil {
    return err
}
b.SetSmallIcon1_1(17301620) // android.R.drawable.ic_dialog_info
b.SetContentTitle("Hello from Go!")
b.SetContentText("Posted via JNI — no Java code")
notif, err := b.Build()
if err != nil {
    return err
}

// 5. Post it
if err := mgr.Notify2(1, notif); err != nil {
    return err
}
```

## Importance Levels

```go
notification.ImportanceNone    // 0 - no sound or visual
notification.ImportanceMin     // 1 - no sound, low visual priority
notification.ImportanceLow     // 2 - no sound
notification.ImportanceDefault // 3 - sound
notification.ImportanceHigh    // 4 - sound + heads-up display
```

## Required Permissions (Android 13+)

```xml
<uses-permission android:name="android.permission.POST_NOTIFICATIONS" />
```

On Android 13+ the permission must be requested at runtime before posting.
