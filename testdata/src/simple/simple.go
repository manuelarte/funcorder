package simple

//nolint:nonamedreturns // testing linter
func NewOtherMyStruct() (m *MyStruct) { // want "should be placed after the struct declaration"
	m = &MyStruct{Name: "John"}
	return
}

func NewMyStruct() *MyStruct { // want "should be placed after the struct declaration"
	return &MyStruct{Name: "John"}
}

type MyStruct struct {
	Name string
}

func (m MyStruct) GetName() string {
	return m.Name
}
