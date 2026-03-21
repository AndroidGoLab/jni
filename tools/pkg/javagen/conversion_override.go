package javagen

// ConversionOverride specifies custom type conversion expressions.
type ConversionOverride struct {
	ToGo   string `yaml:"to_go"`
	ToJava string `yaml:"to_java"`
}
