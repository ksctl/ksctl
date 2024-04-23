package consts

import "github.com/fatih/color"

type CustomExternalLogLevel string

var (
	LOG_WARNING = CustomExternalLogLevel(color.HiYellowString("WARN"))
	LOG_ERROR   = CustomExternalLogLevel(color.HiRedString("ERR"))
	LOG_INFO    = CustomExternalLogLevel(color.HiGreenString("INFO"))
	LOG_DEBUG   = CustomExternalLogLevel(color.HiBlueString("DEBUG"))
)
