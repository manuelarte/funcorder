package simple

type Greetings struct {
}

func NewOtherGreetings() (m *Greetings) {
	m = &Greetings{}
	return
}

func NewGreetings() *Greetings { // want `constructor \"NewGreetings" should be placed before \"NewOtherGreetings\"`
	return &Greetings{}
}

func (m Greetings) GoodMorning() string {
	return "hello"
}

func (m *Greetings) GoodAfternoon(name string) string { // want `method \"GoodAfternoon\" for struct \"Greetings\" should be placed before \"GoodMorning\"`
	return "bye"
}

func (m Greetings) hello() string {
	return "hello"
}

func (m *Greetings) bye(name string) string {
	return "bye"
} // want `method \"bye\" for struct \"Greetings\" should be placed before \"hello\"`
