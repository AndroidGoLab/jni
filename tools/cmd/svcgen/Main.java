import java.lang.reflect.Field;
import java.lang.reflect.Modifier;
import java.util.ArrayList;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

/**
 * Extracts the mapping from Android service class names to their
 * Context.*_SERVICE constant values by reflecting on android.jar.
 *
 * Usage: java -cp android.jar:. Main
 * Output: JSON object on stdout, e.g. {"android.app.AlarmManager": "alarm", ...}
 */
public class Main {
    public static void main(String[] args) throws Exception {
        String jarPath = findAndroidJar();
        if (jarPath == null) {
            System.err.println("error: android.jar not found on classpath");
            System.exit(1);
        }

        // Extract *_SERVICE constant values from android.content.Context.
        Class<?> contextClass = Class.forName("android.content.Context");
        // constant name -> service value, e.g. "ALARM_SERVICE" -> "alarm"
        Map<String, String> serviceConstants = new HashMap<>();
        for (Field f : contextClass.getFields()) {
            int mods = f.getModifiers();
            if (!Modifier.isPublic(mods) || !Modifier.isStatic(mods) || !Modifier.isFinal(mods)) {
                continue;
            }
            if (f.getType() != String.class) {
                continue;
            }
            String name = f.getName();
            if (!name.endsWith("_SERVICE")) {
                continue;
            }
            String value = (String) f.get(null);
            if (value != null) {
                serviceConstants.put(name, value);
            }
        }

        // Collect all android.* class names from android.jar, grouped by simple name.
        // Also build a reverse index: normalized simple name -> list of full names,
        // for fuzzy fallback matching.
        Map<String, List<String>> simpleNameToFull = new HashMap<>();
        // normalizedToFull: key = lowercase class name without Manager/Service suffix
        Map<String, List<String>> normalizedToFull = new HashMap<>();
        try (JarFile jar = new JarFile(jarPath)) {
            Enumeration<JarEntry> entries = jar.entries();
            while (entries.hasMoreElements()) {
                JarEntry entry = entries.nextElement();
                String entryName = entry.getName();
                if (!entryName.endsWith(".class") || entryName.contains("$")) {
                    continue;
                }
                String className = entryName.replace('/', '.').replaceAll("\\.class$", "");
                if (!className.startsWith("android.")) {
                    continue;
                }
                String simpleName = className.substring(className.lastIndexOf('.') + 1);
                simpleNameToFull.computeIfAbsent(simpleName, k -> new ArrayList<>()).add(className);

                String normalized = simpleName.toLowerCase();
                for (String suffix : new String[]{"manager", "service"}) {
                    if (normalized.endsWith(suffix)) {
                        normalized = normalized.substring(0, normalized.length() - suffix.length());
                        break;
                    }
                }
                normalizedToFull.computeIfAbsent(normalized, k -> new ArrayList<>()).add(className);
            }
        }

        // Match each *_SERVICE constant to a class.
        //
        // Three-pass approach:
        //  Pass 1: process *_MANAGER_SERVICE constants — their stem already ends in
        //          "Manager", so the exact match finds the Manager class directly.
        //  Pass 2: process remaining *_SERVICE constants with stem-based heuristic.
        //  Pass 3: fuzzy fallback for any still-unmatched constants — normalize the
        //          service value and search by lowercased class stem.
        TreeMap<String, String> result = new TreeMap<>();
        Set<String> claimedClasses = new TreeSet<>();
        List<Map.Entry<String, String>> unmatched = new ArrayList<>();

        // Pass 1: *_MANAGER_SERVICE constants.
        for (Map.Entry<String, String> e : serviceConstants.entrySet()) {
            if (!e.getKey().endsWith("_MANAGER_SERVICE")) {
                continue;
            }
            String matched = findClassByStem(e.getKey(), simpleNameToFull, claimedClasses);
            if (matched != null) {
                result.put(matched, e.getValue());
                claimedClasses.add(matched);
            }
        }

        // Pass 2: remaining *_SERVICE constants, stem-based.
        for (Map.Entry<String, String> e : serviceConstants.entrySet()) {
            if (e.getKey().endsWith("_MANAGER_SERVICE")) {
                continue;
            }
            String matched = findClassByStem(e.getKey(), simpleNameToFull, claimedClasses);
            if (matched != null) {
                result.put(matched, e.getValue());
                claimedClasses.add(matched);
            } else {
                unmatched.add(e);
            }
        }

        // Pass 3: fuzzy fallback — match by normalized service value.
        for (Map.Entry<String, String> e : unmatched) {
            String svcValue = e.getValue();
            // Normalize: "wifirtt" stays "wifirtt", "app_search" -> "appsearch"
            String normalized = svcValue.replace("_", "");
            List<String> candidates = normalizedToFull.get(normalized);
            if (candidates == null) {
                continue;
            }
            String best = pickBest(candidates, claimedClasses);
            if (best != null) {
                result.put(best, svcValue);
                claimedClasses.add(best);
            }
        }

        // Output as JSON.
        StringBuilder sb = new StringBuilder();
        sb.append("{\n");
        int i = 0;
        for (Map.Entry<String, String> entry : result.entrySet()) {
            if (i > 0) {
                sb.append(",\n");
            }
            sb.append("  ");
            sb.append(jsonQuote(entry.getKey()));
            sb.append(": ");
            sb.append(jsonQuote(entry.getValue()));
            i++;
        }
        sb.append("\n}\n");
        System.out.print(sb.toString());
    }

    /**
     * Find a class matching the given *_SERVICE constant name using stem-based
     * heuristics. Tries PascalCase stem + common suffixes.
     * Skips classes already in claimedClasses.
     */
    private static String findClassByStem(
            String constName,
            Map<String, List<String>> simpleNameToFull,
            Set<String> claimedClasses
    ) {
        String stem = constName.substring(0, constName.length() - "_SERVICE".length());
        String pascalStem = toPascalCase(stem);

        // Candidate suffixes: try Manager first (most common), then exact, then Service.
        String[] suffixes = {"Manager", "", "Service"};
        for (String suffix : suffixes) {
            String candidate = pascalStem + suffix;
            List<String> fullNames = simpleNameToFull.get(candidate);
            if (fullNames == null) {
                continue;
            }
            String best = pickBest(fullNames, claimedClasses);
            if (best != null) {
                return best;
            }
        }
        return null;
    }

    /**
     * Pick the best fully-qualified class name from candidates,
     * preferring android.content.* / android.app.* over deprecated packages
     * like android.text.*, and skipping already-claimed classes.
     */
    private static String pickBest(List<String> candidates, Set<String> claimed) {
        String best = null;
        int bestScore = Integer.MAX_VALUE;
        for (String c : candidates) {
            if (claimed.contains(c)) {
                continue;
            }
            int score = packageScore(c);
            if (best == null || score < bestScore) {
                best = c;
                bestScore = score;
            }
        }
        return best;
    }

    /**
     * Assign a preference score to a package path. Lower is better.
     * Prefer modern packages over deprecated ones.
     */
    private static int packageScore(String className) {
        // android.text.ClipboardManager is deprecated in favor of
        // android.content.ClipboardManager — prefer content/app packages.
        if (className.startsWith("android.content.")) return 0;
        if (className.startsWith("android.app.")) return 1;
        if (className.startsWith("android.text.")) return 100;
        return 10;
    }

    /** Convert SCREAMING_SNAKE to PascalCase: "ALARM" -> "Alarm", "WIFI_P2P" -> "WifiP2p". */
    private static String toPascalCase(String screaming) {
        StringBuilder sb = new StringBuilder();
        for (String part : screaming.split("_")) {
            if (part.isEmpty()) continue;
            sb.append(part.charAt(0));
            sb.append(part.substring(1).toLowerCase());
        }
        return sb.toString();
    }

    /** Find the android.jar path from the java.class.path system property. */
    private static String findAndroidJar() {
        String cp = System.getProperty("java.class.path");
        if (cp == null) return null;
        for (String entry : cp.split(System.getProperty("path.separator"))) {
            if (entry.endsWith("android.jar")) {
                return entry;
            }
        }
        return null;
    }

    /** JSON-escape a string value. */
    private static String jsonQuote(String s) {
        return "\"" + s.replace("\\", "\\\\").replace("\"", "\\\"") + "\"";
    }
}
