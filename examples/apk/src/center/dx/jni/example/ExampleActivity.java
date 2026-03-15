package center.dx.jni.example;

import android.app.Activity;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.graphics.Typeface;
import android.os.Bundle;
import android.util.Log;
import android.widget.ScrollView;
import android.widget.TextView;

import java.util.ArrayList;

public class ExampleActivity extends Activity {
    static {
        System.loadLibrary("example");
    }

    private static final int PERM_REQUEST_CODE = 1;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        String[] needed = getUngranted();
        if (needed.length > 0) {
            requestPermissions(needed, PERM_REQUEST_CODE);
        } else {
            runAndDisplay();
        }
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] grantResults) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults);
        runAndDisplay();
    }

    private String[] getDeclaredPermissions() {
        try {
            ApplicationInfo ai = getPackageManager()
                    .getApplicationInfo(getPackageName(), PackageManager.GET_META_DATA);
            if (ai.metaData == null) {
                return new String[0];
            }
            String csv = ai.metaData.getString("example.permissions", "");
            if (csv.isEmpty()) {
                return new String[0];
            }
            return csv.split(",");
        } catch (PackageManager.NameNotFoundException e) {
            return new String[0];
        }
    }

    private String[] getUngranted() {
        String[] declared = getDeclaredPermissions();
        ArrayList<String> needed = new ArrayList<>();
        for (String perm : declared) {
            if (checkSelfPermission(perm) != PackageManager.PERMISSION_GRANTED) {
                needed.add(perm);
            }
        }
        return needed.toArray(new String[0]);
    }

    private void runAndDisplay() {
        // Show a loading message immediately so the user knows it's working.
        ScrollView sv = new ScrollView(this);
        TextView tv = new TextView(this);
        tv.setPadding(32, 32, 32, 32);
        tv.setTextSize(14f);
        tv.setTypeface(Typeface.MONOSPACE);
        tv.setText("Running example…");
        sv.addView(tv);
        setContentView(sv);

        // Run native code on a background thread to avoid ANR on slow
        // operations (e.g. waiting for a GPS fix).
        new Thread(() -> {
            nativeRun();
            String output = nativeGetOutput();
            Log.i("GoJNI", output);
            runOnUiThread(() -> tv.setText(output));
        }).start();
    }

    private native void nativeRun();
    private native String nativeGetOutput();
}
