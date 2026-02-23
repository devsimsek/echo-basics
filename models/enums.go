package models

type FlagEnum string

const (
	LogFlag   FlagEnum = "log"
	DebugFlag FlagEnum = "debug"
	InfoFlag  FlagEnum = "info"
	WarnFlag  FlagEnum = "warn"
	ErrorFlag FlagEnum = "error"
	TraceFlag FlagEnum = "trace"
)
