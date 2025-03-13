package simple

func NewOtherMyStruct() (m *MyStruct) {
	m = &MyStruct{Name: "John"}
	return
}

func NewMyStruct() *MyStruct {
	return &MyStruct{Name: "John"}
}

type MyStruct struct {
	Name string
}

func (m MyStruct) GetName() string {
	return m.Name
}
