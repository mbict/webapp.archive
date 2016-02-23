package webapp

import (
	. "gopkg.in/check.v1"
)

type EnvSuite struct{}

var _ = Suite(&EnvSuite{})

func (s *EnvSuite) TestEnv(c *C) {

	tests := []struct {
		expected string
		values   []string
	}{
		{
			expected: Production,
			values:   []string{"RELEASE", "release", "live", "production"},
		}, {
			expected: Development,
			values:   []string{"DEV", "dev", "debug", "development"},
		}, {
			expected: Testing,
			values:   []string{"TEST", "test", "testing"},
		},
	}

	for index, test := range tests {
		for _, env := range test.values {
			SetEnv(env)

			c.Check(Env(), Equals, test.expected, Commentf("test %d failed on SET(`%s`) value %s == %s", index, env, Env(), test.expected))
		}
	}
}

func (s *EnvSuite) TestPanicUnkownEnv(c *C) {
	wrapPanic := func() {
		SetEnv("unkown")
	}

	c.Assert(wrapPanic, PanicMatches, "unknown environment: `unkown`")
}

func (s *EnvSuite) TestSetDefaultEnv(c *C) {
	setDefaultEnv("Production")

	c.Assert(Env(), Equals, Production)
}

func (s *EnvSuite) TestSetDefaultEnvDefaultsToDevWhenEmptyString(c *C) {
	setDefaultEnv("")

	c.Assert(Env(), Equals, Development)
}
