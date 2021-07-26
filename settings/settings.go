package settings

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
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
	viper.SetDefault("leetcode.enter", "cn")
	viper.SetDefault("out", ".")
	viper.SetDefault("datafile", "data.json")
	viper.SetConfigName("settings")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.lctool")
	viper.SetEnvPrefix("lctool")
	//viper.BindEnv("leetcode.username")
	//viper.BindEnv("leetcode.password")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Errorf("find settings.toml failed, %v", err)
	}
	Setting = setting{
		Enter:    viper.GetString("leetcode.enter"),
		Username: viper.GetString("leetcode.username"),
		Password: viper.GetString("leetcode.password"),
		SaveFile: viper.GetString("datafile"),
		Out:      viper.GetString("out"),
	}
}
