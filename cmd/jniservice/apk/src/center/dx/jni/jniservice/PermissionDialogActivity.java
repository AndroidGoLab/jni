package center.dx.jni.jniservice;

import android.app.Activity;
import android.database.sqlite.SQLiteDatabase;
import android.os.Bundle;
import android.view.Gravity;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.ScrollView;
import android.widget.TextView;

/**
 * Displays a permission approval dialog when a client requests method access.
 * Launch via Intent with extras:
 *   - "request_id" (long): the pending request row ID
 *   - "client_id" (String): the client's CN
 *   - "methods" (String): comma-separated method list
 *
 * Approval/denial writes directly to the shared SQLite database using
 * Android's built-in SQLiteDatabase API.
 */
public class PermissionDialogActivity extends Activity {
    private String dbPath;

    private long requestId;
    private String clientId;
    private String methods;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // DB path matches the Go server's data dir detection.
        java.io.File filesDir = new java.io.File(getApplicationInfo().dataDir, "files/jniservice");
        dbPath = new java.io.File(filesDir, "acl.db").getAbsolutePath();

        requestId = getIntent().getLongExtra("request_id", -1);
        clientId = getIntent().getStringExtra("client_id");
        methods = getIntent().getStringExtra("methods");

        if (requestId < 0 || clientId == null || methods == null) {
            finish();
            return;
        }

        LinearLayout layout = new LinearLayout(this);
        layout.setOrientation(LinearLayout.VERTICAL);
        layout.setPadding(48, 48, 48, 48);

        TextView title = new TextView(this);
        title.setText("Permission Request");
        title.setTextSize(20);
        title.setGravity(Gravity.CENTER);
        layout.addView(title);

        TextView body = new TextView(this);
        body.setText("\nClient: " + clientId + "\n\nRequested methods:\n" + methods.replace(",", "\n"));
        body.setTextSize(14);
        body.setPadding(0, 24, 0, 24);

        ScrollView scroll = new ScrollView(this);
        scroll.addView(body);
        layout.addView(scroll, new LinearLayout.LayoutParams(
                LinearLayout.LayoutParams.MATCH_PARENT, 0, 1));

        LinearLayout buttons = new LinearLayout(this);
        buttons.setOrientation(LinearLayout.HORIZONTAL);
        buttons.setGravity(Gravity.CENTER);

        Button approveBtn = new Button(this);
        approveBtn.setText("Approve");
        approveBtn.setOnClickListener(v -> {
            handleApprove();
            finish();
        });
        buttons.addView(approveBtn);

        Button denyBtn = new Button(this);
        denyBtn.setText("Deny");
        denyBtn.setOnClickListener(v -> {
            handleDeny();
            finish();
        });
        buttons.addView(denyBtn);

        layout.addView(buttons);
        setContentView(layout);
    }

    private void handleApprove() {
        SQLiteDatabase db = SQLiteDatabase.openDatabase(
                dbPath, null, SQLiteDatabase.OPEN_READWRITE);
        try {
            db.beginTransaction();
            db.execSQL("UPDATE pending_requests SET status='approved' WHERE id=?",
                    new Object[]{requestId});
            String[] methodList = methods.split(",");
            for (String method : methodList) {
                String trimmed = method.trim();
                if (!trimmed.isEmpty()) {
                    String now = java.time.Instant.now().toString();
                    db.execSQL("INSERT OR IGNORE INTO grants (client_id, method_pattern, granted_at, granted_by) VALUES (?, ?, ?, 'ui')",
                            new Object[]{clientId, trimmed, now});
                }
            }
            db.setTransactionSuccessful();
        } finally {
            db.endTransaction();
            db.close();
        }

        // Also request Android runtime permissions that the client's methods
        // might need. This ensures the SERVICE has the permissions, not just
        // the client's ACL grant.
        java.util.List<String> needed = new java.util.ArrayList<>();
        String[] dangerous = {
            android.Manifest.permission.CAMERA,
            android.Manifest.permission.RECORD_AUDIO,
            android.Manifest.permission.ACCESS_FINE_LOCATION,
            android.Manifest.permission.ACCESS_COARSE_LOCATION,
        };
        for (String perm : dangerous) {
            if (checkSelfPermission(perm) != android.content.pm.PackageManager.PERMISSION_GRANTED) {
                needed.add(perm);
            }
        }
        if (!needed.isEmpty()) {
            requestPermissions(needed.toArray(new String[0]), 100);
        }
    }

    private void handleDeny() {
        SQLiteDatabase db = SQLiteDatabase.openDatabase(
                dbPath, null, SQLiteDatabase.OPEN_READWRITE);
        try {
            db.execSQL("UPDATE pending_requests SET status='denied' WHERE id=?",
                    new Object[]{requestId});
        } finally {
            db.close();
        }
    }
}
