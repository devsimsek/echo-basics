package modules

import "go.smsk.dev/pkgs/basics/echo-basics/models"

func GetLogLevel(flag models.FlagEnum) int {
	switch flag {
	case models.LogFlag:
		return 0
	case models.DebugFlag:
		return 1
	case models.InfoFlag:
		return 2
	case models.WarnFlag:
		return 3
	case models.ErrorFlag:
		return 4
	case models.TraceFlag:
		return 5
	default:
		return -1
	}
}
