package config

import (
	"io/fs"

	"github.com/charmbracelet/log"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type UserConfig struct {
	RootPaths []string `mapstructure:"rootpaths"`
	MountDir  string   `mapstructure:"mount"`
	CacheDir  string   `mapstructure:"cachedir"`
	Debug     bool     `mapstructure:"debug"`
}

var config UserConfig

func InitConfigWithViper() {
	viper.Unmarshal(&config)
}

func InitializeConfig() {
	flag.StringSliceP("rootpaths",
		"r",
		[]string{"~/"},
		"Provide a comma separated paths to mount,"+
			" e.g. mutifs --rootpaths /home/john/Movies,/mnt/Drive/Music",
	)

	flag.StringP("mount", "m", "/tmp/mnt", "Directory to mount at. If it does"+
		" not exist, it is created before mounting.")

	flag.StringP("cachedir",
		"l",
		"~/.multifs/cache",
		"Cache directory to store file to hash mappings")

	flag.BoolP(
		"debug",
		"d",
		false,
		"should show debug information",
	)

	configFile := flag.StringP("config", "c", "", "config file in .yaml format")

	flag.Parse()

	if *configFile == "" {
		viper.BindPFlags(flag.CommandLine)
	} else {
		viper.SetConfigFile(*configFile)
		err := viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(*fs.PathError); ok {
				log.Warn("config file not found, running with command-line and default arguments")
				viper.BindPFlags(flag.CommandLine)
			} else {
				log.Fatal("Error in parsing config file : ", "error", err)
			}
		}
	}
	InitConfigWithViper()
}

func GetConfig() UserConfig {
	return config
}
