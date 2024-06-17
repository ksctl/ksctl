package consts

import "github.com/fatih/color"

type CustomExternalLogLevel string

var (
	LogWarning = CustomExternalLogLevel(color.HiYellowString("WARN"))
	LogError   = CustomExternalLogLevel(color.HiRedString("ERR"))
	LogInfo    = CustomExternalLogLevel(color.HiBlueString("INFO"))
	LogNote    = CustomExternalLogLevel(color.CyanString("NOTE"))
	LogSuccess = CustomExternalLogLevel(color.HiGreenString("PASS"))
	LogDebug   = CustomExternalLogLevel(color.WhiteString("DEBUG"))
)

type LogClusterDetail uint

const (
	LoggingGetClusters LogClusterDetail = iota
	LoggingInfoCluster LogClusterDetail = iota
)
