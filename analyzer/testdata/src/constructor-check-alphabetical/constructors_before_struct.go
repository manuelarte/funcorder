package simple

func NewOtherMyStruct() (m *MyStruct) { // want "should be placed after the struct declaration"
	m = &MyStruct{Name: "John"}
	return
}

func NewMyStruct() *MyStruct { // want "should be placed after the struct declaration"
	return &MyStruct{Name: "John"}
}

func MustMyStruct() *MyStruct { // want `function \"MustMyStruct\" for struct \"MyStruct\" should be placed after the struct declaration`
	return NewMyStruct()
}

type MyStruct struct {
	Name string
}

func (m MyStruct) lenName() int {
	return len(m.Name)
}

func (m MyStruct) GetName() string {
	return m.Name
}

func (m *MyStruct) SetName(name string) {
	m.Name = name
}
