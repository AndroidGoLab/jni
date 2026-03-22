package center.dx.jni.generated;

import android.hardware.camera2.CameraDevice;
import center.dx.jni.internal.GoAbstractDispatch;

public class CameraDeviceCallback extends CameraDevice.StateCallback {
    private final long handlerID;

    public CameraDeviceCallback(long handlerID) {
        this.handlerID = handlerID;
    }

    @Override
    public void onOpened(CameraDevice camera) {
        GoAbstractDispatch.invoke(handlerID, "onOpened", new Object[]{camera});
    }

    @Override
    public void onDisconnected(CameraDevice camera) {
        GoAbstractDispatch.invoke(handlerID, "onDisconnected", new Object[]{camera});
    }

    @Override
    public void onError(CameraDevice camera, int error) {
        GoAbstractDispatch.invoke(handlerID, "onError", new Object[]{camera, Integer.valueOf(error)});
    }
}
