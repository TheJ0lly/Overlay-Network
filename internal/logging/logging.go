// This is a primitive way of logging. It WILL change in the future but I don't know when.
package logging

import (
	"fmt"
	"os"
	"time"
)

var Debug bool = false
var defaultTimeFormat string = "02/01/2006 15:04:05"

func LogDebug(format string, args ...any) {
	if Debug {
		fmt.Printf(fmt.Sprintf("[DBG] [%s] %s\n", time.Now().Format(defaultTimeFormat), format), args...)
	}
}

func LogInfo(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("[INF] [%s] %s\n", time.Now().Format(defaultTimeFormat), format), args...)
}

func LogError(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("[ERR] [%s] %s\n", time.Now().Format(defaultTimeFormat), format), args...)
}

func LogErrorWithExit(format string, args ...any) {
	LogError(format, args...)
	os.Exit(1)
}
