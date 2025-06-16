package multipletypes

type (
	MyStructSorted  struct{}
	MyStructSorted2 struct{}
)

func NewMyStructSorted() *MyStructSorted {
	return &MyStructSorted{}
}

func NewMyStructSorted2() *MyStructSorted2 {
	return &MyStructSorted2{}
}

func (MyStructSorted) hello() string {
	return "hello"
}

func (MyStructSorted2) bye() string {
	return "bye"
}
