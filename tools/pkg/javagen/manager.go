package javagen

// ManagerData holds template data for generating a system service manager constructor.
type ManagerData struct {
	GoType      string
	ServiceName string
	HasClose    bool
}

// IsManager returns true if the class is obtained via system service.
func IsManager(cls *MergedClass) bool {
	return cls.Obtain == "system_service"
}

// ManagerConstructorData returns template data for the NewManager constructor.
func ManagerConstructorData(cls *MergedClass) ManagerData {
	return ManagerData{
		GoType:      cls.GoType,
		ServiceName: cls.ServiceName,
		HasClose:    cls.Close,
	}
}
