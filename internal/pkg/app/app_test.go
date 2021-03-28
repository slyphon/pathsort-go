package app

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const PathFull = "/home/slyphon/.goenv/shims:/home/slyphon/.goenv/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/sbin:/bin"

const Config1 = `
tag_order = ["goenv", "usr_local", "usr_bins", "bins"]
[patterns]
usr_bins = "\\A/usr/s?bin$"
usr_local = "\\A/usr/local/bin"
bins = "\\A/s?bin$"
goenv = "/\\.goenv(/|$)"
`
const Home = "/home/slyphon"

type PathsortSuite struct {
	suite.Suite
	config string
}

// hook up tests
func TestPathsortSuite(t *testing.T) {
	suite.Run(t, new(PathsortSuite))
}

func (ps *PathsortSuite) TestConfigParsing() {
	var config *Config
	var err error

	config, err = LoadConfigString(Config1, Home)
	ps.NoError(err)
	ps.NotNil(config)
}
