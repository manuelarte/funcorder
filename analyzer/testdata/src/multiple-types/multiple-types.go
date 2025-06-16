package multipletypes

type (
	MyStruct  struct{}
	MyStruct2 struct{}
)

func (MyStruct) hello() string {
	return "hello"
}

func (MyStruct2) bye() string {
	return "bye"
}

func NewMyStruct() *MyStruct { // want `constructor "NewMyStruct" for struct "MyStruct" should be placed before struct method "hello"`
	return &MyStruct{}
}

func NewMyStruct2() *MyStruct2 { // want `constructor "NewMyStruct2" for struct "MyStruct2" should be placed before struct method "bye`
	return &MyStruct2{}
}
