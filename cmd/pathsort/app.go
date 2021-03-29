package pathsort

import (
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	app "github.com/slyphon/pathsort-go/internal/pathsorter"
)

var (
	origPath          string
	home              string
	defaultConfigPath string
	envConfigPath     string
)

func init() {
	origPath = os.Getenv("PATH")
	home = os.Getenv("HOME")
	defaultConfigPath = path.Join(home, ".pathsort.toml")
	envConfigPath = os.Getenv("PATHSORT_CONFIG")

	if os.Getenv("PATHSORT_DEBUG") != "" {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
}

func isDir(path string) bool {
	st, err := os.Stat(path)

	if err == nil {
		return st.IsDir()
	}
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		return false
	} else {
		log.Printf("Error: could not stat path %v", path)
		return false
	}
}

func App() {
	if origPath == "" {
		log.Fatal("PATH was not set!")
	}

	var config *app.Config
	var err error

	configPath := defaultConfigPath
	if envConfigPath != "" {
		configPath = envConfigPath
	}

	if config, err = app.LoadConfigFile(configPath, home, os.Environ()); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}

	newPaths := config.Fix(origPath)

	cleanPaths := make([]string, 0, len(newPaths))
	for _, p := range newPaths {
		if isDir(p) {
			cleanPaths = append(cleanPaths, p)
		}
	}

	fmt.Printf("export PATH=\"%s\"\n", strings.Join(cleanPaths, ":"))
}
