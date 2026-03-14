package main

import (
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/xaionaro-go/jni/tools/pkg/grpcgen"
)

func main() {
	specsDir := flag.String("specs", "spec/java", "directory containing per-package YAML specs")
	overlaysDir := flag.String("overlays", "spec/overlays/java", "directory containing overlays")
	outputDir := flag.String("output", ".", "base output directory")
	goModule := flag.String("go-module", "github.com/xaionaro-go/jni", "Go module path")
	protoDir := flag.String("proto-dir", "proto", "directory containing compiled proto Go stubs")
	flag.Parse()

	specs, err := filepath.Glob(filepath.Join(*specsDir, "*.yaml"))
	if err != nil {
		log.Fatalf("glob specs: %v", err)
	}
	if len(specs) == 0 {
		log.Fatalf("no spec files found in %s", *specsDir)
	}

	var entries []grpcgen.CompositeEntry

	for _, specPath := range specs {
		baseName := strings.TrimSuffix(filepath.Base(specPath), ".yaml")
		overlayPath := filepath.Join(*overlaysDir, baseName+".yaml")

		serverEntries, err := grpcgen.GenerateServer(specPath, overlayPath, *outputDir, *goModule, *protoDir)
		if err != nil {
			log.Fatalf("generate server %s: %v", baseName, err)
		}
		entries = append(entries, serverEntries...)

		if err := grpcgen.GenerateClient(specPath, overlayPath, *outputDir, *goModule, *protoDir); err != nil {
			log.Fatalf("generate client %s: %v", baseName, err)
		}
	}

	if err := grpcgen.GenerateCompositeServer(entries, *outputDir, *goModule); err != nil {
		log.Fatalf("generate composite server: %v", err)
	}
	if err := grpcgen.GenerateCompositeClient(entries, *outputDir, *goModule); err != nil {
		log.Fatalf("generate composite client: %v", err)
	}
}
