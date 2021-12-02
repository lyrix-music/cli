package config

import (
	"errors"
	"fmt"
	"github.com/lyrix-music/cli/cmd/desktop/logging"
	"os"
	"path/filepath"

	"github.com/lyrix-music/cli/auth"
	"github.com/lyrix-music/cli/types"
	"github.com/spf13/viper"
)

var logger = logging.GetLogger()
var absConfigPathYaml = ""
var configPath = ""

func GetPath(appName string) (string, string) {
	userHomeDir, err := os.UserConfigDir()
	if err != nil {
		logger.Fatal(err)
	}
	configPath := filepath.Join(userHomeDir, "lyrix", appName)
	absConfigPathYaml := filepath.Join(configPath, "config.yaml")
	return configPath, absConfigPathYaml
}

func Preload(appName string) {
	// load the configuration
	viper.SetConfigName("config")                                        // name of config file (without extension)
	viper.SetConfigType("yaml")                                          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/lyrix/%s/", appName)) // call multiple times to add many search paths
	viper.AddConfigPath(fmt.Sprintf("/etc/lyrix/%s/", appName))          // path to look for the config file in
	viper.AddConfigPath(".")
	a, _ := GetPath("lyrixd")
	viper.AddConfigPath(a)

	viper.AddConfigPath(fmt.Sprintf("/etc/lyrix/%s/", appName))

}

func write(absConfigPathYaml string) {
	if err := viper.SafeWriteConfigAs(absConfigPathYaml); err != nil {
		if os.IsNotExist(err) {

			err = viper.WriteConfigAs(absConfigPathYaml)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

}

func Load(appName string) (*types.UserInstance, error) {

	Preload(appName)
	configPath, absConfigPathYaml = GetPath(appName)
	err := os.MkdirAll(configPath, 0o755)
	if err != nil {
		logger.Fatal(err)
	}
	username, token, backendUrl := "", "", ""
	if err := viper.ReadInConfig(); err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok && appName == "lyrixd" {
			// Config file not found; ignore error if desired
			logger.Info("Did not find configuration file. Attempting to interactively create one")
			auth.Login()
		} else if ok {
			logger.Info("Did not find configuration file.")

			return nil, nil
		} else {
			logger.Fatal(err)
			return nil, err
		}

	}

	username = viper.Get("Username").(string)
	token = viper.Get("AuthToken").(string)
	backendUrl = viper.Get("Host").(string)
	if (username == "" || token == "" || backendUrl == "") && appName == "lyrixd" {
		auth.Login()
		logger.Info("Re-run the app to reload from configuration.")
		return nil, errors.New("reload-configuration")
	}

	Write()
	authInstance := &types.UserInstance{Username: username, Token: token, Host: backendUrl}
	return authInstance, nil
}

func Write() {
	write(absConfigPathYaml)
}

func Set(key string, value string) {
	viper.Set(key, value)

}
