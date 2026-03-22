package center.dx.jni.generated;

import android.hardware.camera2.CameraCaptureSession;
import center.dx.jni.internal.GoAbstractDispatch;

public class CaptureSessionCallback extends CameraCaptureSession.StateCallback {
    private final long handlerID;

    public CaptureSessionCallback(long handlerID) {
        this.handlerID = handlerID;
    }

    @Override
    public void onConfigured(CameraCaptureSession session) {
        GoAbstractDispatch.invoke(handlerID, "onConfigured", new Object[]{session});
    }

    @Override
    public void onConfigureFailed(CameraCaptureSession session) {
        GoAbstractDispatch.invoke(handlerID, "onConfigureFailed", new Object[]{session});
    }
}
