package packagelevelorder

// unexported function appears before exported — violation expected.

func helper() string { // want `unexported function "helper" should be placed after the exported function "PublicFunc"`
	return "helper"
}

func PublicFunc() string {
	return "public"
}
