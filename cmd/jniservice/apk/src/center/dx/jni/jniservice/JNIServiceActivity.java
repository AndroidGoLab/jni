package center.dx.jni.jniservice;

import android.Manifest;
import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.os.Build;
import android.os.Bundle;
import android.view.Gravity;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;

import java.util.ArrayList;
import java.util.List;

public class JNIServiceActivity extends Activity {
    private static final int PERMISSION_REQUEST_CODE = 1;
    private static final int BG_LOCATION_REQUEST_CODE = 2;

    private static final String[] DANGEROUS_PERMISSIONS = {
        Manifest.permission.ACCESS_FINE_LOCATION,
        Manifest.permission.ACCESS_COARSE_LOCATION,
        Manifest.permission.CAMERA,
        Manifest.permission.RECORD_AUDIO,
        Manifest.permission.READ_PHONE_STATE,
        Manifest.permission.CALL_PHONE,
        Manifest.permission.BODY_SENSORS,
    };

    private TextView statusText;

    @Override
    protected void onResume() {
        super.onResume();
        updateStatus();
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        LinearLayout layout = new LinearLayout(this);
        layout.setOrientation(LinearLayout.VERTICAL);
        layout.setPadding(48, 48, 48, 48);

        statusText = new TextView(this);
        statusText.setTextSize(16);
        layout.addView(statusText);

        Button grantBtn = new Button(this);
        grantBtn.setText("Grant All Permissions");
        grantBtn.setOnClickListener(v -> requestAllPermissions());
        LinearLayout.LayoutParams btnParams = new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.WRAP_CONTENT,
                LinearLayout.LayoutParams.WRAP_CONTENT);
        btnParams.topMargin = 32;
        btnParams.gravity = Gravity.CENTER_HORIZONTAL;
        layout.addView(grantBtn, btnParams);

        setContentView(layout);
        updateStatus();

        // Start the foreground service (no permissions needed for gRPC itself).
        Intent intent = new Intent(this, JNIServiceForeground.class);
        startForegroundService(intent);
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);

        // After foreground location is granted, request background location separately
        // (Android requires this two-step flow).
        if (requestCode == PERMISSION_REQUEST_CODE) {
            if (checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED
                    && checkSelfPermission(Manifest.permission.ACCESS_BACKGROUND_LOCATION) != PackageManager.PERMISSION_GRANTED) {
                requestPermissions(
                        new String[]{Manifest.permission.ACCESS_BACKGROUND_LOCATION},
                        BG_LOCATION_REQUEST_CODE);
                return;
            }
        }

        updateStatus();
    }

    private void updateStatus() {
        List<String> missing = getMissingPermissions();
        StringBuilder sb = new StringBuilder();
        sb.append("jniservice is running.\n\n");
        sb.append("Connect with:\n  jnicli --addr <device-ip>:50051 --insecure jni get-version\n\n");
        if (missing.isEmpty()) {
            sb.append("All permissions granted.");
        } else {
            sb.append("Missing permissions (").append(missing.size()).append("):\n");
            for (String perm : missing) {
                sb.append("  - ").append(perm.replace("android.permission.", "")).append("\n");
            }
        }
        statusText.setText(sb.toString());
    }

    private void requestAllPermissions() {
        // First request all non-background-location permissions.
        List<String> needed = getMissingPermissions();
        // Background location must be requested separately after foreground location.
        needed.remove(Manifest.permission.ACCESS_BACKGROUND_LOCATION);

        if (!needed.isEmpty()) {
            requestPermissions(needed.toArray(new String[0]), PERMISSION_REQUEST_CODE);
        } else if (checkSelfPermission(Manifest.permission.ACCESS_BACKGROUND_LOCATION) != PackageManager.PERMISSION_GRANTED) {
            // Foreground location already granted, just need background.
            requestPermissions(
                    new String[]{Manifest.permission.ACCESS_BACKGROUND_LOCATION},
                    BG_LOCATION_REQUEST_CODE);
        } else {
            updateStatus();
        }
    }

    private List<String> getMissingPermissions() {
        List<String> missing = new ArrayList<>();
        for (String perm : DANGEROUS_PERMISSIONS) {
            if (checkSelfPermission(perm) != PackageManager.PERMISSION_GRANTED) {
                missing.add(perm);
            }
        }

        if (Build.VERSION.SDK_INT >= 31) {
            String[] btPerms = {
                "android.permission.BLUETOOTH_CONNECT",
                "android.permission.BLUETOOTH_SCAN",
                "android.permission.BLUETOOTH_ADVERTISE",
            };
            for (String perm : btPerms) {
                if (checkSelfPermission(perm) != PackageManager.PERMISSION_GRANTED) {
                    missing.add(perm);
                }
            }
        }

        if (Build.VERSION.SDK_INT >= 33) {
            if (checkSelfPermission("android.permission.POST_NOTIFICATIONS") != PackageManager.PERMISSION_GRANTED) {
                missing.add("android.permission.POST_NOTIFICATIONS");
            }
        }

        if (checkSelfPermission(Manifest.permission.ACCESS_BACKGROUND_LOCATION) != PackageManager.PERMISSION_GRANTED) {
            missing.add(Manifest.permission.ACCESS_BACKGROUND_LOCATION);
        }

        return missing;
    }
}
