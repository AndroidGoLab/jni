package main

import (
	"flag"
	"log"

	"github.com/xaionaro-go/jni/tools/pkg/specgen"
)

func main() {
	refDir := flag.String("ref", "ref", "directory containing reference .class files")
	extraCP := flag.String("classpath", "", "additional classpath for javap (e.g. android.jar)")
	outputDir := flag.String("output", "spec/java", "output directory for generated YAML specs")
	goModule := flag.String("go-module", "github.com/xaionaro-go/jni", "Go module path")
	flag.Parse()

	if err := specgen.GenerateFromRefDir(*refDir, *extraCP, *outputDir, *goModule); err != nil {
		log.Fatalf("generate specs: %v", err)
	}
}
