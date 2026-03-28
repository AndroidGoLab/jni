package specgen

// SpecAbstractCallback is an abstract class callback in the YAML spec.
// Unlike SpecCallback (which represents a Java interface), this
// represents an abstract class whose abstract methods are delegated
// to Go via GoAbstractDispatch.
type SpecAbstractCallback struct {
	JavaClass string                       `yaml:"java_class"`
	GoType    string                       `yaml:"go_type"`
	Methods   []SpecAbstractCallbackMethod `yaml:"methods"`
}
