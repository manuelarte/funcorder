package multipletypes

type (
	MyStruct  struct{}
	MyStruct2 struct{}
)

func NewMyStruct() *MyStruct {
	return &MyStruct{}
}

func NewMyStruct2() *MyStruct2 {
	return &MyStruct2{}
}
