package main

import (
	"encoding/json"
	"fmt"
)

func printResult(v any) error {
	switch flagFormat {
	case "json":
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
		fmt.Println(string(data))
	case "text":
		switch val := v.(type) {
		case bool:
			fmt.Println(val)
		case string:
			fmt.Println(val)
		case int32:
			fmt.Println(val)
		case int64:
			fmt.Println(val)
		default:
			data, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal result: %w", err)
			}
			fmt.Println(string(data))
		}
	default:
		return fmt.Errorf("unknown format: %s", flagFormat)
	}
	return nil
}
