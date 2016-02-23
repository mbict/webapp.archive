package webapp

import (
	"fmt"
	"os"
	"strings"
)

const ENV = "ENV"

const (
	Production  string = "production"
	Development string = "development"
	Testing     string = "testing"
)

var env string = Development

func init() {
	setDefaultEnv(os.Getenv(ENV))
}

func setDefaultEnv(env string) {
	if len(env) == 0 {
		SetEnv(Development)
	} else {
		SetEnv(env)
	}
}

func SetEnv(value string) {
	switch strings.ToLower(value) {
	case "dev":
		fallthrough
	case "debug":
		fallthrough
	case "development":
		env = Development

	case "release":
		fallthrough
	case "live":
		fallthrough
	case "production":
		env = Production

	case "test":
		fallthrough
	case "testing":
		env = Testing

	default:
		panic(fmt.Sprintf("unknown environment: `%s`", value))
	}
}

func Env() string {
	return env
}
