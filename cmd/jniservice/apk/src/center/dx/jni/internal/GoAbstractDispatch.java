package center.dx.jni.internal;

/**
 * GoAbstractDispatch provides a static native method for generated callback
 * adapters to delegate abstract method calls to Go via JNI.
 *
 * The native implementation is registered via JNI RegisterNatives by the
 * Go c-shared library during proxy initialization. Generated adapter classes
 * (produced by callbackgen) call this method from each overridden abstract
 * method, passing the handler ID, method name, and boxed arguments.
 */
public class GoAbstractDispatch {
    /**
     * Called by generated adapter classes when an abstract method is invoked.
     * Primitive arguments must be auto-boxed into the Object array by the caller.
     *
     * @param handlerID unique ID mapping to the Go proxy handler
     * @param methodName the name of the invoked abstract method
     * @param args method arguments (primitives boxed as wrapper types)
     * @return the return value (or null for void methods)
     */
    public static native Object invoke(long handlerID, String methodName, Object[] args);
}
