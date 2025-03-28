package simple

func NewOtherMyStruct() (m *MyStruct) {
	m = &MyStruct{Name: "John"}
	return
}

func NewMyStruct() *MyStruct {
	return &MyStruct{Name: "John"}
}

func MustMyStruct() *MyStruct {
	return NewMyStruct()
}

type MyStruct struct {
	Name string
}

func (m MyStruct) lenName() int { // want `unexported method \"lenName\" for struct \"MyStruct\" should be placed after the exported method \"SetName\"`
	return len(m.Name)
}

func (m MyStruct) GetName() string {
	return m.Name
}

func (m *MyStruct) SetName(name string) {
	m.Name = name
}
