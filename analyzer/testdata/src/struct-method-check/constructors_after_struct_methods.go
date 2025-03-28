package simple

type MyStruct2 struct {
	Name string
}

func (m MyStruct2) GetName() string {
	return m.Name
}

func (m *MyStruct2) SetName(name string) {
	m.Name = name
}

func NewOtherMyStruct2() (m *MyStruct2) {
	m = &MyStruct2{Name: "John"}
	return
}

func NewMyStruct2() *MyStruct2 {
	return &MyStruct2{Name: "John"}
}
