package app

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const PathFull = "/home/slyphon/.goenv/shims:/home/slyphon/.goenv/bin:/opt/wtfisthis:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/sbin:/bin:/home/slyphon/bin"

const Config1 = `
tag_order = ["goenv", "home_bin", "usr_local", "usr_bins", "bins"]
[patterns]
usr_bins = "\\A/usr/s?bin$"
usr_local = "\\A/usr/local/s?bin"
bins = "\\A/s?bin$"
goenv = "/\\.goenv(/|$)"
home_bin = "\\A@HOME@/bin$"
NULL = "/wtfisthis$"
`

var configPatternMap = map[string]string{
	"usr_bins":  "\\A/usr/s?bin$",
	"usr_local": "\\A/usr/local/s?bin",
	"bins":      "\\A/s?bin$",
	"goenv":     "/\\.goenv(/|$)",
	"NULL":      "/wtfisthis$",
	"home_bin":  "\\A/home/slyphon/bin$",
}

const Home = "/home/slyphon"

type PathsortSuite struct {
	suite.Suite
	config *Config
	r      *require.Assertions
}

var _ suite.SetupAllSuite = (*PathsortSuite)(nil)
var _ suite.SetupTestSuite = (*PathsortSuite)(nil)

// hook up tests
func TestPathsortSuite(t *testing.T) {
	suite.Run(t, new(PathsortSuite))
}

func (ps *PathsortSuite) SetupSuite() {
	ps.r = ps.Require()
}

func (ps *PathsortSuite) SetupTest() {
	var err error
	ps.config, err = LoadConfigString(Config1, Home)
	ps.r.NoError(err, "failed to LoadConfigString in SetupTest: %v", Config1)
	ps.r.NotNil(ps.config)
}

func (s *PathsortSuite) TestConfigParsing() {
	s.r.NotNil(s.config.Order)
	s.r.NotNil(s.config.Patterns)
	s.r.NotNil(s.config.Tags)
}

func (s *PathsortSuite) TestOrderProperty() {
	s.r.Equal(
		[]string{"goenv", "home_bin", "usr_local", "usr_bins", "bins"},
		s.config.Order)
}

func (s *PathsortSuite) TestPatternsProperty() {
	for i, tag := range s.config.Tags {
		p, ok := configPatternMap[tag]
		s.r.True(ok, "key %v not found in map", tag)
		s.r.Equal(p, s.config.Patterns[i].String())
	}
}

func (s *PathsortSuite) TestFixOrdering() {
	expected := "/home/slyphon/.goenv/shims:/home/slyphon/.goenv/bin:/home/slyphon/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/sbin:/bin"
	s.r.Equal(expected, s.config.Fix(PathFull))
}
