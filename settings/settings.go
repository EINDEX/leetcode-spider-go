package settings

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

var Setting setting

type setting struct {
	Enter    string
	Username string
	Password string
}

func init() {
	if _, err := toml.DecodeFile("settings.toml", &Setting); err != nil {
		fmt.Errorf("%v", err)
		return
	}

}
