package constructor_check_alphabetical_fix

type ConstructorAfterStructMethodsWithComments struct {
	Name string
}

// GetName returns the name.
func (m ConstructorAfterStructMethodsWithComments) GetName() string {
	return m.Name
}

// SetName sets the name
// multi line comment
func (m *ConstructorAfterStructMethodsWithComments) SetName(name string) {
	m.Name = name
}

// NewOtherConstructorAfterStructMethodsWithComments This constructor creates the
// struct ConstructorAfterStructMethodsWithComments
// with a named return
func NewOtherConstructorAfterStructMethodsWithComments() (m *ConstructorAfterStructMethodsWithComments) { // want `constructor \"NewOtherConstructorAfterStructMethodsWithComments\" for struct \"ConstructorAfterStructMethodsWithComments\" should be placed before struct method \"GetName\"`
	m = &ConstructorAfterStructMethodsWithComments{Name: "John"}
	return
}

func NewConstructorAfterStructMethodsWithComments() *ConstructorAfterStructMethodsWithComments { // want `constructor \"NewConstructorAfterStructMethodsWithComments\" for struct \"ConstructorAfterStructMethodsWithComments\" should be placed before struct method \"GetName\"` `constructor \"NewConstructorAfterStructMethodsWithComments\" for struct \"ConstructorAfterStructMethodsWithComments\" should be placed before constructor \"NewOtherConstructorAfterStructMethodsWithComments\"`
	return &ConstructorAfterStructMethodsWithComments{Name: "John"}
}
