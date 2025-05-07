package constructor_check_fix

// NewOtherConstructorAfterStructMethodsWithComments This constructor creates the
// struct ConstructorBeforeStructWithComments
// with a named return
func NewOtherConstructorBeforeStructWithComments() (m *ConstructorBeforeStructWithComments) { // want `constructor \"NewOtherConstructorBeforeStructWithComments\" for struct \"ConstructorBeforeStructWithComments\" should be placed after the struct declaration`
	m = &ConstructorBeforeStructWithComments{Name: "John"}
	return
}

func NewConstructorBeforeStructWithComments() *ConstructorBeforeStructWithComments { // want `constructor \"NewConstructorBeforeStructWithComments\" for struct \"ConstructorBeforeStructWithComments\" should be placed after the struct declaration`
	return &ConstructorBeforeStructWithComments{Name: "John"}
}

type ConstructorBeforeStructWithComments struct {
	Name string
}

// GetName returns the name.
func (m ConstructorBeforeStructWithComments) GetName() string {
	return m.Name
}

// SetName sets the name
// multi line comment
func (m *ConstructorBeforeStructWithComments) SetName(name string) {
	m.Name = name
}
