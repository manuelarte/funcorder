package functionorder

// exported functions appear before unexported — no violation.

func AnotherPublicFunc() string {
	return "another public"
}

func anotherHelper() string {
	return "another helper"
}
