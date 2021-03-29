package pathsorter

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/shell"
	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

func mustGetEnv(k string) string {
	v, ok := os.LookupEnv(k)
	if !ok {
		log.Panicf("environment var %s was not set", k)
	}
	return v
}

var (
	Path     = mustGetEnv("PATH")
	envVarRe = regexp.MustCompile("([$][A-Za-z_][A-Za-z0-9_]+)\\b")
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

// if we couldn't locate a var in the environment, we return the empty string
func replaceEnvVars(ptrn string, envMap map[string]string) string {
	var i []int
	var result []string
	remain := ptrn
	for {

		i = envVarRe.FindStringIndex(remain)

		if i == nil {
			// we're done here, so take on the remaining string
			// to the result slice
			result = append(result, remain)
			break
		}

		// i was not nil, so we must hve found a result, therefore the
		// positions to the left of i[0] are non-match, and should be added to
		// result
		result = append(result, remain[0:i[0]])

		// look for varname in environment, here we strip the leading '$'
		varname := remain[i[0]+1 : i[1]]

		if v, ok := envMap[varname]; ok {
			// we found varname in the environment, that means we
			// need to append the value (instead of the variable)
			result = append(result, v)
			// and the rest of the string becomes "remain"
			remain = remain[i[1]:]
		} else {
			// we didn't find the variable in the given env map
			// so we return the empty string to indicate that we were
			// unable to make a substitution and allow the caller to
			// decide how to handle it
			return ""
		}
	}

	return strings.Join(result, "")
}

const notFoundVarLogMessage = "pattern %q contained environment variables that could not be expanded, it will be ignored"

func loadConfigFromTree(tree *toml.Tree, homeStr string, env []string) (config *Config, err error) {
	var tags []string
	var patterns []*regexp.Regexp
	var notFoundVarTags map[string]bool
	envMap := shell.BuildEnvs(env)

	for tag, ptrn := range tree.Get("patterns").(*toml.Tree).ToMap() {
		tags = append(tags, tag)
		var p string // god go is fucking obnoxious
		p = ptrn.(string)

		newp := replaceEnvVars(p, envMap)
		if newp == "" {
			log.Warningf(notFoundVarLogMessage, p)
			notFoundVarTags[tag] = true
			continue
		}
		reg := regexp.MustCompile(newp)
		patterns = append(patterns, reg)
	}

	// filter out tags with not found environment variables from the ordering list
	// since we won't ever find those tags in PATH
	var tagOrder []string
	for _, tag := range tree.GetArray("tag_order").([]string) {
		if _, ok := notFoundVarTags[tag]; ok {
			continue
		} else {
			tagOrder = append(tagOrder, tag)
		}
	}

	config = &Config{
		Tags:     tags,
		Patterns: patterns,
		Order:    tagOrder,
	}

	log.Debugf("config loaded:")
	log.Debugf("tags:     %+v", config.Tags)
	log.Debugf("patterns: %+v", config.Patterns)
	log.Debugf("order:    %+v", config.Order)

	if err = config.validateNames(); err != nil {
		return nil, err
	}

	return
}

// "/home/slyphon/code/pathsort-go/config.toml"

func LoadConfigString(tomlStr string, homeStr string, env []string) (config *Config, err error) {
	var tree *toml.Tree
	if tree, err = toml.Load(tomlStr); err != nil {
		return nil, err
	}
	return loadConfigFromTree(tree, homeStr, env)
}

func LoadConfigFile(configPath string, homeStr string, env []string) (config *Config, err error) {
	var tree *toml.Tree

	if tree, err = toml.LoadFile(configPath); err != nil {
		return nil, err
	}

	return loadConfigFromTree(tree, homeStr, env)
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
