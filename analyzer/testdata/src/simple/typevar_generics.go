package simple

import "strings"

type logOutputFormat int

const (
	LogOutputFormatAuto logOutputFormat = iota
	LogOutputFormatConsole
	LogOutputFormatJSON
)

type logoutPutFormatOrString interface {
	string | logOutputFormat
}

func NewLogOutputFormat[T logoutPutFormatOrString](format T) logOutputFormat {
	switch v := any(format).(type) {
	case string:
		switch strings.ToLower(v) {
		case "auto":
			return LogOutputFormatAuto
		case "console":
			return LogOutputFormatConsole
		case "json":
			return LogOutputFormatJSON
		default:
			panic("unknown log output format: " + v)
		}
	case logOutputFormat:
		return v
	default:
		panic("unknown type for log output format")
	}
}
