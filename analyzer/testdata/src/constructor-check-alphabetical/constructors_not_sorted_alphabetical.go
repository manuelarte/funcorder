package simple

type Greetings struct {
}

func NewOtherGreetings() (m *Greetings) {
	m = &Greetings{}
	return
}

func NewGreetings() *Greetings { // want `constructor \"NewGreetings\" for struct \"Greetings\" should be placed before constructor \"NewOtherGreetings\"`
	return &Greetings{}
}

func (m Greetings) GoodMorning() string {
	return "hello"
}

func (m *Greetings) GoodAfternoon(name string) string {
	return "bye"
}

func (m Greetings) hello() string {
	return "hello"
}

func (m *Greetings) bye(name string) string {
	return "bye"
}
