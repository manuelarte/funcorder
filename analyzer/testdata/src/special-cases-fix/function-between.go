package special_cases_fix

// FunctionBetweenStruct comment
type FunctionBetweenStruct struct {
}

// Open comment
func Open() (*FunctionBetweenStruct, error) {
	return &FunctionBetweenStruct{}, nil
}

// isOpen comment
func (f *FunctionBetweenStruct) isOpen() bool { // want `unexported method "isOpen" for struct "FunctionBetweenStruct" should be placed after the exported method "Close"`
	return true
}

// cannotOpenFileError comment
type cannotOpenFileError struct {
	err error
}

// Unwrap comment
func (e *cannotOpenFileError) Unwrap() error {
	return e.err
}

func (f *FunctionBetweenStruct) Close() {
}
