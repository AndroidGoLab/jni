package main

import (
	"flag"
	"log"

	"github.com/xaionaro-go/jni/tools/pkg/callbackgen"
)

func main() {
	specPath := flag.String("spec", "spec/callbacks.yaml", "path to callbacks spec")
	outputDir := flag.String("output", "cmd/jniservice/apk/src/center/dx/jni/generated", "output directory")
	flag.Parse()

	if err := callbackgen.Generate(*specPath, *outputDir); err != nil {
		log.Fatal(err)
	}
}
