package jnigen

import "fmt"

// FamilyOverlay defines overlay for a typed function family.
type FamilyOverlay struct {
	GoPattern      string         `yaml:"go_pattern"`
	CheckException bool           `yaml:"check_exception"`
	Params         []ParamOverlay `yaml:"params"`
	// Returns can be a map (keyed by "primitive", "Object", "Void") or a string.
	// We use a custom unmarshaler.
	Returns any `yaml:"returns,omitempty"`
}

// FamilyReturnMap returns the returns field as a map if it is one, or nil.
func (fo *FamilyOverlay) FamilyReturnMap() map[string]string {
	if fo.Returns == nil {
		return nil
	}
	switch v := fo.Returns.(type) {
	case map[string]any:
		result := make(map[string]string, len(v))
		for k, val := range v {
			if val == nil {
				result[k] = ""
			} else {
				result[k] = fmt.Sprintf("%v", val)
			}
		}
		return result
	case string:
		// A single string means all types use this return type.
		return map[string]string{"_all": v}
	}
	return nil
}
