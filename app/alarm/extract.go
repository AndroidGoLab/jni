package alarm

import (
	"fmt"

	"github.com/AndroidGoLab/jni"
)

// ExtractAlarmClockInfo extracts fields from an AlarmManager.AlarmClockInfo JNI object.
func ExtractAlarmClockInfo(
	env *jni.Env,
	obj *jni.Object,
) (*AlarmClockInfo, error) {
	if err := ensureInit(env); err != nil {
		return nil, err
	}

	cls, err := env.FindClass("android/app/AlarmManager$AlarmClockInfo")
	if err != nil {
		return nil, fmt.Errorf("find AlarmClockInfo: %w", err)
	}
	defer env.DeleteLocalRef(&cls.Object)

	mid, err := env.GetMethodID(cls, "getTriggerTime", "()J")
	if err != nil {
		return nil, fmt.Errorf("get getTriggerTime: %w", err)
	}
	triggerTime, err := env.CallLongMethod(obj, mid)
	if err != nil {
		return nil, fmt.Errorf("call getTriggerTime: %w", err)
	}

	showIntentMid, err := env.GetMethodID(cls, "getShowIntent", "()Landroid/app/PendingIntent;")
	if err != nil {
		return nil, fmt.Errorf("get getShowIntent: %w", err)
	}
	showIntent, err := env.CallObjectMethod(obj, showIntentMid)
	if err != nil {
		return nil, fmt.Errorf("call getShowIntent: %w", err)
	}

	return &AlarmClockInfo{
		TriggerTime: triggerTime,
		ShowIntent:  showIntent,
	}, nil
}
