package main

import (
	"log"
	"os"
	"github.com/pelletier/go-toml"
)

var (
	origPath string
)

var tomlESQU = toml

func init() {
	origPath = os.Getenv("PATH")
}

func main() {
	if origPath == "" {
		log.Fatal("PATH was not set!")
	}

}
