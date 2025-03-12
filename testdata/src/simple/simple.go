package simple

func NewMyStruct() *MyStruct {
	return &MyStruct{Name: "John"}
}

type MyStruct struct {
	Name string
}

func (m MyStruct) GetName() string {
	return m.Name
}
