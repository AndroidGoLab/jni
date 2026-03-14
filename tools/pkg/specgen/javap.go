// Package specgen generates Java API YAML specs from .class files using javap.
package specgen

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// JavapClass holds the parsed output of javap for one class.
type JavapClass struct {
	FullName    string // e.g. "android.app.KeyguardManager"
	IsInterface bool
	IsAbstract  bool
	IsFinal     bool
	Constants   []JavapConstant
	Methods     []JavapMethod
	Implements  []string
}

// JavapConstant is a public static final field.
type JavapConstant struct {
	Name     string // e.g. "ERROR_BAD_VALUE"
	JavaType string // e.g. "int", "java.lang.String"
}

// JavapMethod is a public method parsed from javap output.
type JavapMethod struct {
	Name       string
	ReturnType string // "void", "int", "boolean", "java.lang.String", etc.
	Params     []JavapParam
	IsStatic   bool
	Throws     bool
}

// JavapParam is a method parameter.
type JavapParam struct {
	JavaType string
}

var (
	classLineRe    = regexp.MustCompile(`^public\s+(abstract\s+)?(final\s+)?(class|interface)\s+(\S+)`)
	implementsRe   = regexp.MustCompile(`implements\s+(.+)\s*\{`)
	constantRe     = regexp.MustCompile(`^\s+public static final\s+(\S+)\s+(\w+);`)
	methodRe       = regexp.MustCompile(`^\s+public\s+(static\s+)?(\S+)\s+(\w+)\(([^)]*)\)(.*);\s*$`)
	constructorRe  = regexp.MustCompile(`^\s+public\s+\S+\(([^)]*)\)(.*);\s*$`)
)

// RunJavap executes javap and parses the output for a single class.
// classPath can contain multiple entries separated by ":".
func RunJavap(classPath string, className string) (*JavapClass, error) {
	var cmd *exec.Cmd
	if classPath != "" {
		cmd = exec.Command("javap", "-public", "-cp", classPath, className)
	} else {
		// No classpath — javap uses the JDK's default (for java.* classes).
		cmd = exec.Command("javap", "-public", className)
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("javap %s: %w", className, err)
	}
	return parseJavap(string(out))
}

func parseJavap(output string) (*JavapClass, error) {
	lines := strings.Split(output, "\n")
	jc := &JavapClass{}

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")

		// Parse class declaration.
		if m := classLineRe.FindStringSubmatch(line); m != nil {
			jc.IsAbstract = strings.TrimSpace(m[1]) == "abstract"
			jc.IsFinal = strings.TrimSpace(m[2]) == "final"
			jc.IsInterface = m[3] == "interface"
			jc.FullName = m[4]

			if im := implementsRe.FindStringSubmatch(line); im != nil {
				for _, iface := range strings.Split(im[1], ",") {
					jc.Implements = append(jc.Implements, strings.TrimSpace(iface))
				}
			}
			continue
		}

		// Parse constants (public static final).
		if m := constantRe.FindStringSubmatch(line); m != nil {
			jc.Constants = append(jc.Constants, JavapConstant{
				JavaType: m[1],
				Name:     m[2],
			})
			continue
		}

		// Parse methods.
		if m := methodRe.FindStringSubmatch(line); m != nil {
			isStatic := strings.TrimSpace(m[1]) == "static"
			retType := m[2]
			name := m[3]
			paramsStr := m[4]
			throwsStr := m[5]

			method := JavapMethod{
				Name:       name,
				ReturnType: retType,
				IsStatic:   isStatic,
				Throws:     strings.Contains(throwsStr, "throws"),
				Params:     parseParams(paramsStr),
			}
			jc.Methods = append(jc.Methods, method)
			continue
		}

		// Skip constructors for now (handled by obtain type).
		_ = constructorRe
	}

	if jc.FullName == "" {
		return nil, fmt.Errorf("could not parse class name from javap output")
	}
	return jc, nil
}

func parseParams(s string) []JavapParam {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	params := make([]JavapParam, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		params = append(params, JavapParam{JavaType: p})
	}
	return params
}
