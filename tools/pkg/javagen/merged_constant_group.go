package javagen

import "strings"

// MergedConstantGroup groups constants by their Go type.
type MergedConstantGroup struct {
	GoType   string
	BaseType string
	Values   []MergedConstant
}

// HasVars reports whether this group contains any var-only values
// (e.g. NaN, Infinity).
func (g MergedConstantGroup) HasVars() bool {
	for _, v := range g.Values {
		if v.IsVar {
			return true
		}
	}
	return false
}

// ConstValues returns only the values that can be Go constants.
func (g MergedConstantGroup) ConstValues() []MergedConstant {
	var result []MergedConstant
	for _, v := range g.Values {
		if !v.IsVar {
			result = append(result, v)
		}
	}
	return result
}

// VarValues returns only the values that must be Go vars (NaN, Infinity).
func (g MergedConstantGroup) VarValues() []MergedConstant {
	var result []MergedConstant
	for _, v := range g.Values {
		if v.IsVar {
			result = append(result, v)
		}
	}
	return result
}

// NeedsMathImport reports whether any value in the group uses math package
// functions (e.g. math.NaN(), math.Inf()).
func (g MergedConstantGroup) NeedsMathImport() bool {
	for _, v := range g.Values {
		if strings.Contains(v.Value, "math.") {
			return true
		}
	}
	return false
}
