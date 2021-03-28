package app

import (
	"log"
	"os"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml"
)

func mustGetEnv(k string) string {
	v, ok := os.LookupEnv(k)
	if !ok {
		log.Panicf("environment var %s was not set", k)
	}
	return v
}

var (
	Path = mustGetEnv("PATH")
)

type (
	Config struct {
		Tags     []string
		Patterns []*regexp.Regexp
		Order    []string
	}
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadConfigFromTree(tree *toml.Tree, homeStr string) (config *Config, err error) {
	var tags []string
	var patterns []*regexp.Regexp

	for tag, ptrn := range tree.Get("patterns").(*toml.Tree).ToMap() {
		tags = append(tags, tag)
		var p string // god go is fucking obnoxious
		p = ptrn.(string)
		p = strings.Replace(p, "@HOME@", homeStr, -1)
		reg := regexp.MustCompile(p)
		patterns = append(patterns, reg)
	}

	config = &Config{
		Tags:     tags,
		Patterns: patterns,
		Order:    tree.GetArray("tag_order").([]string),
	}

	return
}

// "/home/slyphon/code/pathsort-go/config.toml"

func LoadConfigString(tomlStr string, homeStr string) (config *Config, err error) {
	var tree *toml.Tree
	if tree, err = toml.Load(tomlStr); err != nil {
		return nil, err
	}
	return loadConfigFromTree(tree, homeStr)
}

func LoadConfigFile(path string, homeStr string) (config *Config, err error) {
	var tree *toml.Tree

	if tree, err = toml.LoadFile(path); err != nil {
		return nil, err
	}

	return loadConfigFromTree(tree, homeStr)
}
