package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/srevinsaju/lyrix/lyrixd/auth"
	"github.com/srevinsaju/lyrix/lyrixd/types"
)

func GetLocalConfigPath() (string, string) {
	userHomeDir, err := os.UserConfigDir()
	if err != nil {
		logger.Fatal(err)
	}
	configPath := filepath.Join(userHomeDir, "lyrix", "lyrixd")
	absConfigPathYaml := filepath.Join(configPath, "config.yaml")
	return configPath, absConfigPathYaml
}

func PreLoadConfig() {
	// load the configuration
	viper.SetConfigName("config")                      // name of config file (without extension)
	viper.SetConfigType("yaml")                        // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("$HOME/.config/lyrix/lyrixd/") // call multiple times to add many search paths
	viper.AddConfigPath("/etc/lyrix/lyrixd/")          // path to look for the config file in
	viper.AddConfigPath(".")

}

func PostLoadConfig(absConfigPathYaml string) {
	if err := viper.SafeWriteConfigAs(absConfigPathYaml); err != nil {
		if os.IsNotExist(err) {

			err = viper.WriteConfigAs(absConfigPathYaml)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

}

func LoadConfig() (*types.UserInstance, error) {

	PreLoadConfig()
	if os.Args[len(os.Args)-1] == "reset" {
		auth.Login()
		logger.Info("Re-run the app to reload from configuration.")
		return nil, errors.New("reload-configuration")
	}

	username, token, backendUrl := "", "", ""
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			logger.Info("Did not find configuration file. Attempting to interactively create one")
			auth.Login()
		} else {
			logger.Fatal(err)
			return nil, err
		}

	}

	username = viper.Get("Username").(string)
	token = viper.Get("AuthToken").(string)
	backendUrl = viper.Get("Host").(string)
	if username == "" || token == "" || backendUrl == "" {
		auth.Login()
		logger.Info("Re-run the app to reload from configuration.")
		return nil, errors.New("reload-configuration")
	}

	configPath, absConfigPathYaml := GetLocalConfigPath()
	err := os.MkdirAll(configPath, 0o755)
	if err != nil {
		logger.Fatal(err)
	}

	PostLoadConfig(absConfigPathYaml)

	authInstance := &types.UserInstance{Username: username, Token: token, Host: backendUrl}
	return authInstance, nil
}
