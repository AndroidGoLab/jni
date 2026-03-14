// JNIService loads the jniservice shared library and starts the gRPC server.
//
// Run via app_process on an Android device:
//
//   app_process -Djava.class.path=/data/local/tmp/jniservice.dex \
//       /data/local/tmp JNIService
//
// Configuration is via environment variables (set before invoking app_process):
//
//   JNISERVICE_PORT   — TCP port (default 50051)
//   JNISERVICE_LISTEN — listen address (default 0.0.0.0)
//   JNISERVICE_TOKEN  — bearer token for auth (empty = no auth)
public class JNIService {
    public static void main(String[] args) {
        System.err.println("jniservice: loading shared library");
        try {
            System.load("/data/local/tmp/libjniservice.so");
        } catch (Throwable t) {
            System.err.println("jniservice: load failed: " + t);
            System.exit(1);
        }
        // runServer (called from JNI_OnLoad) blocks until the gRPC server
        // shuts down, so this point is reached only on clean exit.
        System.err.println("jniservice: server exited");
    }
}
