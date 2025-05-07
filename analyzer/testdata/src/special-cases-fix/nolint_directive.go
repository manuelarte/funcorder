package special_cases_fix

//nolint:lll // test directive
func NewNoLLLDirective() *NoLLLDirective { // want `constructor "NewNoLLLDirective" for struct "NoLLLDirective" should be placed after the struct declaration`
	return &NoLLLDirective{}
}

type NoLLLDirective struct{}

// NewNoLLLDirectiveWithDoc constructor for NoLLLDirectiveWithDoc
//
//nolint:lll // test directive
func NewNoLLLDirectiveWithDoc() *NoLLLDirectiveWithDoc { // want `constructor "NewNoLLLDirectiveWithDoc" for struct "NoLLLDirectiveWithDoc" should be placed after the struct declaration`
	return &NoLLLDirectiveWithDoc{}
}

// NoLLLDirectiveWithDoc with docs.
type NoLLLDirectiveWithDoc struct{}
