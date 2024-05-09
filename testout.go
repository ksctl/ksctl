package main

import (
	"os"

	"github.com/ksctl/ksctl/pkg/logger"
)

func main() {
	l := logger.NewGeneralLogger(-1, os.Stdout)
	l.SetPackageName("testing")
	l.Print("Creating a resource", "id", 23423)
	l.Debug("Printing Logger", "loggerStruct", l)
	l.Success("Successfull created data", "name", "dipankar")
	l.Note(`IMPORTANT
Test sample
	> Sample data
`, "foo", "bar")
	l.Warn("retrying", "failed", 0, "max", 15)
	l.Error("Failed in exec", "Reason", l.NewError("invalid data"))
}
