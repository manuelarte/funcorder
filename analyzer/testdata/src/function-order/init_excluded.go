package packagelevelorder

// init is excluded from the rule; exported function after init must not be reported.

func init() {
	// setup
}

func ExportedAfterInit() string {
	return "ok"
}
