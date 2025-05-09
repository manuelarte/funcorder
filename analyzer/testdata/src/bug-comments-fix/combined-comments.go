package bug_comments_fix

type CombinedComments struct {
	Name string
}

func (m CombinedComments) getName() string { // want `method "getName" for struct "CombinedComments" should be placed after the exported method "SetName"`
	// foo is foo
	return m.Name
}

func (m *CombinedComments) SetName(name string) {
	// foo is foo
	m.Name = name
}

func NewOtherCombinedComments() (m *CombinedComments) { // want `constructor "NewOtherCombinedComments" for struct "CombinedComments" should be placed before struct method "getName"`
	// foo is foo
	m = &CombinedComments{Name: "John"}
	return
}

func NewCombinedComments() *CombinedComments { // want `constructor "NewCombinedComments" for struct "CombinedComments" should be placed before struct method "getName"`
	// foo is foo
	return &CombinedComments{Name: "John"}
}
