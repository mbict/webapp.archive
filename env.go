package webapp

import (
	"os"

	"github.com/mattn/go-colorable"
)

const ENV = "ENV"

const (
	Debug   string = "debug"
	Release string = "release"
	Test    string = "test"
)
const (
	debugCode   = iota
	releaseCode = iota
	testCode    = iota
)

var DefaultWriter = colorable.NewColorableStdout()
var env int = debugCode
var envName string = Debug

func init() {
	value := os.Getenv(ENV)
	if len(value) == 0 {
		SetEnv(Debug)
	} else {
		SetEnv(value)
	}
}

func SetEnv(value string) {
	switch value {
	case "DEV":
		env = debugCode
	case Debug:
		env = debugCode
	case Release:
		env = releaseCode
	case "LIVE":
		env = releaseCode
	case Test:
		env = testCode
	default:
		panic("webapp unknown env: " + value)
	}
	envName = value
}

func Env() string {
	return envName
}
