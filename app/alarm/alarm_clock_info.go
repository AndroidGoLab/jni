package alarm

import "github.com/AndroidGoLab/jni"

// AlarmClockInfo holds extracted fields from AlarmManager.AlarmClockInfo.
type AlarmClockInfo struct {
	TriggerTime int64
	ShowIntent  *jni.Object
}
