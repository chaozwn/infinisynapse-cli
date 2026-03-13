package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	configFileName = ".isc"
	configFileType = "yaml"

	KeyServer = "server"
	KeyToken  = "token"
	KeyOutput = "default_output"
	KeyLang   = "lang"
)

var configPath string

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}

	configPath = filepath.Join(home, configFileName+"."+configFileType)

	viper.SetConfigName(configFileName)
	viper.SetConfigType(configFileType)
	viper.AddConfigPath(home)

	viper.SetDefault(KeyOutput, "json")
	viper.SetDefault(KeyLang, "zh-CN")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config: %w", err)
		}
	}

	return nil
}

func GetServer() string {
	return viper.GetString(KeyServer)
}

func GetToken() string {
	return viper.GetString(KeyToken)
}

func GetDefaultOutput() string {
	return viper.GetString(KeyOutput)
}

func GetLang() string {
	return viper.GetString(KeyLang)
}

func Set(key, value string) error {
	viper.Set(key, value)
	return viper.WriteConfigAs(configPath)
}

func SetMultiple(values map[string]string) error {
	for k, v := range values {
		viper.Set(k, v)
	}
	return viper.WriteConfigAs(configPath)
}

func Clear() error {
	viper.Set(KeyServer, "")
	viper.Set(KeyToken, "")
	return viper.WriteConfigAs(configPath)
}

func ConfigFilePath() string {
	return configPath
}
