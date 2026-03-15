package com.xaionaro.jniservice;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.widget.TextView;

public class JNIServiceActivity extends Activity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        TextView tv = new TextView(this);
        tv.setText("jniservice is running.\n\nConnect with:\n  jnicli --addr <device-ip>:50051 --insecure jni get-version");
        tv.setPadding(48, 48, 48, 48);
        tv.setTextSize(16);
        setContentView(tv);

        Intent intent = new Intent(this, JNIServiceForeground.class);
        startForegroundService(intent);
    }
}
