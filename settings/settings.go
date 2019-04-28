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
	SaveFile string
	Out      string
}

func init() {
	if _, err := toml.DecodeFile("settings.toml", &Setting); err != nil {
		fmt.Errorf("%v", err)
		return
	}
	if Setting.Out[len(Setting.Out)-1] == '/' {
		Setting.Out = Setting.Out[:len(Setting.Out)-1]
	}
}
