// Package spec embeds the Java API spec and overlay YAML files.
package spec

import "embed"

// Java contains the Java API spec YAML files.
//
//go:embed java/*.yaml
var Java embed.FS

// Overlays contains the overlay YAML files.
//
//go:embed overlays/java/*.yaml
var Overlays embed.FS
