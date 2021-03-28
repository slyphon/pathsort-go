package pathsort

import (
	"fmt"
	"log"
	"os"
	"path"

	app "github.com/slyphon/pathsort-go/internal/pkg/app"
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
}

func App() {
	if origPath == "" {
		log.Fatal("PATH was not set!")
	}

	var config *app.Config
	var err error

	path := defaultConfigPath
	if envConfigPath != "" {
		path = envConfigPath
	}

	if config, err = app.LoadConfigFile(path, defaultConfigPath); err != nil {
		log.Fatalf("error loading config file: %v", err)
	}

	fmt.Printf("config order: %s\n", config.Order)
}
