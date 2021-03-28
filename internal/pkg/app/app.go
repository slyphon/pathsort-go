package app

import (
	"fmt"
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

	if err = config.validateNames(); err != nil {
		return nil, err
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

func LoadConfigFile(configPath string, homeStr string) (config *Config, err error) {
	var tree *toml.Tree

	if tree, err = toml.LoadFile(configPath); err != nil {
		return nil, err
	}

	return loadConfigFromTree(tree, homeStr)
}

/// validateNames iterates over the Order list of tags and will panic
/// if they are not all defined in the Tags slice
func (c *Config) validateNames() (err error) {
	tmap := make(map[string]bool, len(c.Tags))
	for _, t := range c.Tags {
		tmap[t] = true
	}
	for _, o := range c.Order {
		if _, ok := tmap[o]; !ok {
			return fmt.Errorf("Order name %v not declared in tag patterns. Please check your config", o)
		}
	}
	return nil
}

func (c *Config) makeIndexMap() (imap map[string]int) {
	imap = make(map[string]int, len(c.Order))
	for i, tag := range c.Order {
		imap[tag] = i
	}
	return imap
}

func (c *Config) makeBuckets() (buckets [][]string) {
	buckets = make([][]string, len(c.Order))
	for i := range c.Order {
		buckets[i] = make([]string, 0, 5)
	}
	return buckets
}

/// isDuplicate returns true if needle is found in haystack
func isDuplicate(needle string, haystack []string) bool {
	for _, h := range haystack {
		if needle == h {
			return true
		}
	}
	return false
}

func (c *Config) Fix(pathstr string) (newPathEls []string) {
	imap := c.makeIndexMap()
	buckets := c.makeBuckets()
	var other []string

	pathEls := strings.Split(pathstr, ":")

	var foundMatch bool

	for _, el := range pathEls {
		foundMatch = false
		for i, re := range c.Patterns {
			if re.MatchString(el) {
				foundMatch = true
				// using the index of Patterns, we know the Tag, so we use the
				// indexMap to look up what ordered bucket it goes in, and add
				// this path element to the correct bucket.
				// index -> tag name -> order index
				//
				tag := c.Tags[i]

				// if the tag is the special NULL tag, we drop this path element
				if tag == "NULL" {
					break
				}

				bi := imap[tag]
				if !isDuplicate(el, buckets[bi]) {
					buckets[bi] = append(buckets[bi], el)
					break
				}
			}
		}
		// if we didn't match any patterns, put the path element into the
		// "other" slice for use later
		if !foundMatch {
			other = append(other, el)
		}
	}

	result := make([]string, 0, len(pathEls))

	for _, bucket := range buckets {
		for _, el := range bucket {
			result = append(result, el)
		}
	}

	if other != nil {
		result = append(result, other...)
	}

	return result
}
