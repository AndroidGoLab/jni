package center.dx.jni.jniservice;

import android.content.Context;
import android.graphics.ImageFormat;
import android.hardware.camera2.CameraAccessException;
import android.hardware.camera2.CameraCaptureSession;
import android.hardware.camera2.CameraCharacteristics;
import android.hardware.camera2.CameraDevice;
import android.hardware.camera2.CameraManager;
import android.hardware.camera2.CaptureRequest;
import android.hardware.camera2.TotalCaptureResult;
import android.media.Image;
import android.media.ImageReader;
import android.os.Handler;
import android.os.HandlerThread;
import android.util.Log;
import android.util.Size;
import android.view.Surface;

import java.nio.ByteBuffer;
import java.util.Arrays;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

/**
 * Synchronous camera capture helper for headless (no-preview) use.
 * Call {@link #takePicture(Context, int)} from any thread — it blocks
 * until the JPEG bytes are available (or timeout).
 */
public class CameraCapture {
    private static final String TAG = "CameraCapture";
    private static final int TIMEOUT_SECONDS = 10;

    /**
     * Takes a JPEG photo from the given camera (0 = back, 1 = front).
     * Blocks the calling thread until the image is captured.
     *
     * @return JPEG bytes, or null on failure.
     */
    public static byte[] takePicture(Context context, int cameraIndex) {
        CameraManager manager = (CameraManager) context.getSystemService(Context.CAMERA_SERVICE);
        if (manager == null) {
            Log.e(TAG, "CameraManager not available");
            return null;
        }

        HandlerThread handlerThread = new HandlerThread("CameraCapture");
        handlerThread.start();
        Handler handler = new Handler(handlerThread.getLooper());

        try {
            String[] cameraIds = manager.getCameraIdList();
            if (cameraIndex >= cameraIds.length) {
                Log.e(TAG, "Camera index " + cameraIndex + " out of range (have " + cameraIds.length + ")");
                return null;
            }
            String cameraId = cameraIds[cameraIndex];

            CameraCharacteristics chars = manager.getCameraCharacteristics(cameraId);
            Size[] jpegSizes = chars.get(CameraCharacteristics.SCALER_STREAM_CONFIGURATION_MAP)
                    .getOutputSizes(ImageFormat.JPEG);
            // Use the largest available JPEG size.
            Size size = jpegSizes[0];
            for (Size s : jpegSizes) {
                if (s.getWidth() * s.getHeight() > size.getWidth() * size.getHeight()) {
                    size = s;
                }
            }

            ImageReader imageReader = ImageReader.newInstance(size.getWidth(), size.getHeight(), ImageFormat.JPEG, 1);
            final byte[][] result = {null};
            CountDownLatch imageLatch = new CountDownLatch(1);

            imageReader.setOnImageAvailableListener(reader -> {
                Image image = reader.acquireLatestImage();
                if (image != null) {
                    ByteBuffer buffer = image.getPlanes()[0].getBuffer();
                    result[0] = new byte[buffer.remaining()];
                    buffer.get(result[0]);
                    image.close();
                }
                imageLatch.countDown();
            }, handler);

            // Open camera synchronously.
            CountDownLatch openLatch = new CountDownLatch(1);
            final CameraDevice[] deviceHolder = {null};

            manager.openCamera(cameraId, new CameraDevice.StateCallback() {
                @Override
                public void onOpened(CameraDevice camera) {
                    deviceHolder[0] = camera;
                    openLatch.countDown();
                }
                @Override
                public void onDisconnected(CameraDevice camera) {
                    camera.close();
                    openLatch.countDown();
                }
                @Override
                public void onError(CameraDevice camera, int error) {
                    Log.e(TAG, "Camera open error: " + error);
                    camera.close();
                    openLatch.countDown();
                }
            }, handler);

            if (!openLatch.await(TIMEOUT_SECONDS, TimeUnit.SECONDS) || deviceHolder[0] == null) {
                Log.e(TAG, "Camera open timed out or failed");
                imageReader.close();
                return null;
            }

            CameraDevice camera = deviceHolder[0];

            // Create capture session.
            CountDownLatch sessionLatch = new CountDownLatch(1);
            final CameraCaptureSession[] sessionHolder = {null};

            camera.createCaptureSession(Arrays.asList(imageReader.getSurface()),
                    new CameraCaptureSession.StateCallback() {
                        @Override
                        public void onConfigured(CameraCaptureSession session) {
                            sessionHolder[0] = session;
                            sessionLatch.countDown();
                        }
                        @Override
                        public void onConfigureFailed(CameraCaptureSession session) {
                            Log.e(TAG, "Session configure failed");
                            sessionLatch.countDown();
                        }
                    }, handler);

            if (!sessionLatch.await(TIMEOUT_SECONDS, TimeUnit.SECONDS) || sessionHolder[0] == null) {
                Log.e(TAG, "Session creation timed out or failed");
                camera.close();
                imageReader.close();
                return null;
            }

            // Capture.
            CaptureRequest.Builder captureBuilder = camera.createCaptureRequest(CameraDevice.TEMPLATE_STILL_CAPTURE);
            captureBuilder.addTarget(imageReader.getSurface());
            captureBuilder.set(CaptureRequest.CONTROL_MODE, CaptureRequest.CONTROL_MODE_AUTO);
            captureBuilder.set(CaptureRequest.JPEG_QUALITY, (byte) 90);

            sessionHolder[0].capture(captureBuilder.build(), new CameraCaptureSession.CaptureCallback() {
                @Override
                public void onCaptureCompleted(CameraCaptureSession session, CaptureRequest request, TotalCaptureResult r) {
                    Log.d(TAG, "Capture completed");
                }
            }, handler);

            // Wait for image.
            if (!imageLatch.await(TIMEOUT_SECONDS, TimeUnit.SECONDS)) {
                Log.e(TAG, "Image capture timed out");
            }

            // Cleanup.
            sessionHolder[0].close();
            camera.close();
            imageReader.close();

            return result[0];

        } catch (CameraAccessException | InterruptedException e) {
            Log.e(TAG, "Camera capture failed", e);
            return null;
        } finally {
            handlerThread.quitSafely();
        }
    }
}
