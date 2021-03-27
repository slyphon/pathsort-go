package pathsort

import (
	"testing"
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ToQoz/gopwt"
	"github.com/ToQoz/gopwt/assert"
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

type PathsortSuite struct {
	suite.Suite
	config string
}

func TestMain(m *testing.M) {
	flag.Parse()
	gopwt.Empower()
	os.Exit(m.Run())
}

// hook up tests
func TestPathsortSuite(t *testing.T) {
	suite.Run(t, new(PathsortSuite))
}

type (ps *PathsortSuite) SetupTest() {

}
