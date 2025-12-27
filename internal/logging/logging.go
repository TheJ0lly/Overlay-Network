// This is a primitive way of logging. It WILL change in the future but I don't know when.
package logging

import (
	"fmt"
	"os"
)

var Debug bool = false

func LogDebug(format string, args ...any) {
	if Debug {
		fmt.Printf(fmt.Sprintf("[DBG] %s\n", format), args...)
	}
}

func LogInfo(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("[INF] %s\n", format), args...)
}

func LogError(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("[ERR] %s\n", format), args...)
}

func LogErrorWithExit(format string, args ...any) {
	LogError(format, args...)
	os.Exit(1)
}
