// TestRunner loads a Go shared library and invokes its test entry point.
// Run via: app_process -Djava.class.path=/data/local/tmp/test.dex /data/local/tmp TestRunner
public class TestRunner {
    public static void main(String[] args) {
        System.out.println("E2E_START");
        try {
            System.load("/data/local/tmp/libe2etest.so");
            System.out.println("E2E_LOADED");
        } catch (Throwable t) {
            System.err.println("E2E_LOAD_FAILED: " + t);
            System.exit(1);
        }
        System.out.println("E2E_DONE");
    }
}
