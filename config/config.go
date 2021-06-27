package config

import (
	"errors"
	"fmt"
	"github.com/srevinsaju/lyrix/lyrixd/meta"
	"github.com/srevinsaju/lyrix/lyrixd/service"

	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/srevinsaju/lyrix/lyrixd/auth"
	"github.com/srevinsaju/lyrix/lyrixd/types"
)


var absConfigPathYaml = ""
var configPath = ""

func GetLocalConfigPath(appName string) (string, string) {
	userHomeDir, err := os.UserConfigDir()
	if err != nil {
		service.logger.Fatal(err)
	}
	configPath := filepath.Join(userHomeDir, "lyrix", appName)
	absConfigPathYaml := filepath.Join(configPath, "config.yaml")
	return configPath, absConfigPathYaml
}

func PreLoadConfig(appName string) {
	// load the configuration
	viper.SetConfigName("config")                      // name of config file (without extension)
	viper.SetConfigType("yaml")                        // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/lyrix/%s/", appName)) // call multiple times to add many search paths
	viper.AddConfigPath(fmt.Sprintf("/etc/lyrix/%s/", appName))          // path to look for the config file in
	viper.AddConfigPath(".")

}

func PostLoadConfig(absConfigPathYaml string) {
	if err := viper.SafeWriteConfigAs(absConfigPathYaml); err != nil {
		if os.IsNotExist(err) {

			err = viper.WriteConfigAs(absConfigPathYaml)
			if err != nil {
				service.logger.Fatal(err)
			}
		}
	}

}

func LoadConfig(appName string) (*types.UserInstance, error) {

	PreLoadConfig(appName)

	username, token, backendUrl := "", "", ""
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && appName == "lyrixd"{
			// Config file not found; ignore error if desired
			service.logger.Info("Did not find configuration file. Attempting to interactively create one")
			auth.Login()
		} else {
			service.logger.Fatal(err)
			return nil, err
		}

	}

	username = viper.Get("Username").(string)
	token = viper.Get("AuthToken").(string)
	backendUrl = viper.Get("Host").(string)
	if (username == "" || token == "" || backendUrl == "") && appName == "lyrixd" {
		auth.Login()
		service.logger.Info("Re-run the app to reload from configuration.")
		return nil, errors.New("reload-configuration")
	}

	configPath, absConfigPathYaml = GetLocalConfigPath(meta.AppName)
	err := os.MkdirAll(configPath, 0o755)
	if err != nil {
		service.logger.Fatal(err)
	}

	WriteConfig()
	authInstance := &types.UserInstance{Username: username, Token: token, Host: backendUrl}
	return authInstance, nil
}


func WriteConfig() {
	PostLoadConfig(absConfigPathYaml)
}