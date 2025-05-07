package special_cases_fix

//nolint:lll // test directive
func NewNoLLLDirective() *NoLLLDirective { // want `constructor "NewNoLLLDirective" for struct "NoLLLDirective" should be placed after the struct declaration`
	return &NoLLLDirective{}
}

type NoLLLDirective struct{}
