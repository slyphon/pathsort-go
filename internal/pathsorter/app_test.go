package pathsorter

import (
	"fmt"
	"testing"

	r "github.com/stretchr/testify/require"
)

const PathFull = "/test/slyphon/.goenv/shims:/test/slyphon/.goenv/bin:/opt/wtfisthis:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/opt/pyenv/shims:/opt/pyenv/bin:/sbin:/bin:/test/slyphon/bin:/usr/bin:/usr/sbin"

var TestEnv = []string{
	"HOME=/test/slyphon", // use 'test' here to make sure we aren't picking up the actual env
	"PYENV_ROOT=/opt/pyenv",
}

// let's try variable substitution
const Config1 = `
tag_order = ["pyenv", "goenv", "home_bin", "usr_local", "usr_bins", "bins"]
[patterns]
usr_bins = "\\A/usr/s?bin$"
usr_local = "\\A/usr/local/s?bin"
bins = "\\A/s?bin$"
goenv = "/\\.goenv(/|$)"
home_bin = "\\A$HOME/bin$"
pyenv = "\\A$PYENV_ROOT/(bin|shims)$"
NULL = "/wtfisthis$"
`

var configPatternMap = map[string]string{
	"usr_bins":  "\\A/usr/s?bin$",
	"usr_local": "\\A/usr/local/s?bin",
	"bins":      "\\A/s?bin$",
	"goenv":     "/\\.goenv(/|$)",
	"NULL":      "/wtfisthis$",
	"home_bin":  "\\A/test/slyphon/bin$",
	"pyenv":     "\\A/opt/pyenv/(bin|shims)$",
}

const Home = "/home/slyphon"

func newTestConfig(t *testing.T) *Config {
	conf, err := LoadConfigString(Config1, Home, TestEnv)
	r.NoError(t, err, "failed to parse config in newTestConfig")
	r.NotNil(t, conf)
	return conf
}

func TestConfigParsing(t *testing.T) {
	config := newTestConfig(t)
	r.NotNil(t, config.Order)
	r.NotNil(t, config.Patterns)
	r.NotNil(t, config.Tags)
}

func TestOrderProperty(t *testing.T) {
	config := newTestConfig(t)
	r.Equal(
		t,
		[]string{"pyenv", "goenv", "home_bin", "usr_local", "usr_bins", "bins"},
		config.Order)
}

func TestPatternsProperty(t *testing.T) {
	config := newTestConfig(t)
	for i, tag := range config.Tags {
		p, ok := configPatternMap[tag]
		r.True(t, ok, "key %v not found in map", tag)
		r.Equal(t, p, config.Patterns[i].String())
	}
}

func TestFixOrdering(t *testing.T) {
	config := newTestConfig(t)
	expected := []string{
		"/opt/pyenv/shims",
		"/opt/pyenv/bin",
		"/test/slyphon/.goenv/shims",
		"/test/slyphon/.goenv/bin",
		"/test/slyphon/bin",
		"/usr/local/bin",
		"/usr/local/sbin",
		"/usr/bin",
		"/usr/sbin",
		"/sbin",
		"/bin",
	}

	got := config.Fix(PathFull)

	fmt.Printf("%+v", got)
	r.Equal(t, expected, got)
}

func TestReplaceEnvVars(t *testing.T) {
	pattern := "\\A$PYENV_ROOT/(bin|shims)$"
	env := map[string]string{
		"PYENV_ROOT": "/opt/pyenv",
	}

	r.Equal(t, "\\A/opt/pyenv/(bin|shims)$", replaceEnvVars(pattern, env))
}

func TestReplaceOnlyOnce(t *testing.T) {
	// we don't recurse, there's only one level of replacement
	pattern := "\\A$PYENV_ROOT/(bin|shims)$"

	env := map[string]string{
		"PYENV_ROOT": "$HOME/.pyenv",
	}

	r.Equal(t, "\\A$HOME/.pyenv/(bin|shims)$", replaceEnvVars(pattern, env))
}
