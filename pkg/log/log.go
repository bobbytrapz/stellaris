package log

import (
	"fmt"
	"os"
)

func Info(format string, a ...interface{}) {
	fmt.Printf("\033[34m[*]\033[0m "+format+"\n", a...)
}

func Success(format string, a ...interface{}) {
	fmt.Printf("\033[32m[+]\033[0m "+format+"\n", a...)
}

func Warning(format string, a ...interface{}) {
	fmt.Printf("\033[33m[!]\033[0m "+format+"\n", a...)
}

func Fatal(format string, a ...interface{}) {
	fmt.Printf("\033[31m[-]\033[0m "+format+"\n", a...)
	os.Exit(1)
}

// Errorf returns an error formatted with the red failure prefix
func Errorf(format string, a ...interface{}) error {
	return fmt.Errorf("\033[31m[-]\033[0m "+format, a...)
}
