package utils

import "fmt"


func T(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}