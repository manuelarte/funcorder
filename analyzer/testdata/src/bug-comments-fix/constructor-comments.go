package bug_comments_fix

func NewConstructorComments() *ConstructorComments { // want `constructor "NewConstructorComments" for struct "ConstructorComments" should be placed after the struct declaration`
	// foo is foo
	return &ConstructorComments{Name: "John"}
}

type ConstructorComments struct {
	Name string
}
