package center.dx.jni.jniservice;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.Service;
import android.content.Intent;
import android.os.IBinder;

public class JNIServiceForeground extends Service {
    private static final String CHANNEL_ID = "jniservice";
    private static final int NOTIFICATION_ID = 1;
    private static boolean loaded = false;

    // Native method to pass the APK's Application context to Go.
    // Called after System.loadLibrary so the Go side can store a handle
    // to the real app context (with the correct ClassLoader and permissions).
    private static native void setAppContext(Object context);

    @Override
    public void onCreate() {
        super.onCreate();
        createNotificationChannel();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Notification notification = new Notification.Builder(this, CHANNEL_ID)
                .setContentTitle("jniservice")
                .setContentText("gRPC server running on port 50051")
                .setSmallIcon(android.R.drawable.ic_menu_manage)
                .setOngoing(true)
                .build();

        startForeground(NOTIFICATION_ID, notification);

        if (!loaded) {
            loaded = true;
            try {
                System.loadLibrary("jniservice");
                android.util.Log.i("jniservice", "native library loaded, gRPC server started");

                // Pass the real app context to the Go side. getApplicationContext()
                // returns the Application with the APK's PathClassLoader and permissions.
                setAppContext(getApplicationContext());
                android.util.Log.i("jniservice", "app context passed to native side");
            } catch (Throwable t) {
                android.util.Log.e("jniservice", "failed to load native library", t);
            }
        }

        return START_STICKY;
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    private void createNotificationChannel() {
        NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID, "jniservice", NotificationManager.IMPORTANCE_LOW);
        channel.setDescription("jniservice gRPC server status");
        getSystemService(NotificationManager.class).createNotificationChannel(channel);
    }
}
