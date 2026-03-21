package javagen

// Method describes a Java method to generate a Go wrapper for.
type Method struct {
	JavaMethod string  `yaml:"java_method"`
	GoName     string  `yaml:"go_name"`
	Static     bool    `yaml:"static"`
	Params     []Param `yaml:"params"`
	Returns    string  `yaml:"returns"`
	Error      bool    `yaml:"error"`
}
