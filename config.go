package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/srevinsaju/lyrix/lyrixd/types"
)

func LoadConfig() (types.UserInstance, error) {

	// load the configuration
	viper.SetConfigName("config")                      // name of config file (without extension)
	viper.SetConfigType("yaml")                        // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.config/lyrix/lyrixd/") // call multiple times to add many search paths
	viper.AddConfigPath("/etc/lyrix/lyrixd/")          // path to look for the config file in
	viper.AddConfigPath(".")

	if os.Args[len(os.Args)-1] == "reset" {
		GetUsernameAndToken()
		logger.Info("Re-run the app to reload from configuration.")
		return types.UserInstance{}, errors.New("reload-configuration")
	}

	username, token, backendUrl := "", "", ""
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			logger.Info("Did not find configuration file. Attempting to interactively create one")
			GetUsernameAndToken()
		} else {
			logger.Fatal(err)
			return types.UserInstance{}, err
		}

	}

	username = viper.Get("Username").(string)
	token = viper.Get("AuthToken").(string)
	backendUrl = viper.Get("Host").(string)
	if username == "" || token == "" || backendUrl == "" {
		GetUsernameAndToken()
		logger.Info("Re-run the app to reload from configuration.")
		return types.UserInstance{}, errors.New("reload-configuration")
	}

	userHomeDir, err := os.UserConfigDir()
	if err != nil {
		logger.Fatal(err)
	}
	configPath := filepath.Join(userHomeDir, "lyrix", "lyrixd")
	absConfigPathYaml := filepath.Join(configPath, "config.yaml")
	os.MkdirAll(configPath, 0o755)

	if err := viper.SafeWriteConfigAs(absConfigPathYaml); err != nil {
		if os.IsNotExist(err) {

			err = viper.WriteConfigAs(absConfigPathYaml)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

	auth := types.UserInstance{Username: username, Token: token, Host: backendUrl}
	return auth, nil
}
