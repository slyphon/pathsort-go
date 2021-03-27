package config

import (
	"github.com/pelletier/go-toml"
)


func load() {
	toml.Unmarshal(data []byte, v interface{})
}
