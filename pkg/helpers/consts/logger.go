package consts

import "github.com/fatih/color"

type CustomExternalLogLevel string

var (
	LOG_WARNING = CustomExternalLogLevel(color.HiYellowString("WARN"))
	LOG_ERROR   = CustomExternalLogLevel(color.HiRedString("ERR"))
	LOG_INFO    = CustomExternalLogLevel(color.HiBlueString("INFO"))
	LOG_NOTE    = CustomExternalLogLevel(color.CyanString("NOTE"))
	LOG_SUCCESS = CustomExternalLogLevel(color.HiGreenString("PASS"))
	LOG_DEBUG   = CustomExternalLogLevel(color.WhiteString("DEBUG"))
)
