package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
	"github.com/srevinsaju/lyrix/lyrixd/types"
)

func GetUsernameAndToken() {
	lyrixId := ""
	lyrixIdPrompt := &survey.Input{
		Message: "Enter Lyrix ID (eg: @spongebob@lyrix.local):",
	}
	survey.AskOne(lyrixIdPrompt, &lyrixId)
	if !strings.HasPrefix(lyrixId, "@") {
		logger.Fatal("Invalid Lyrix ID. Your lyrix ID should be in the form of @someone@somedomain.abc")
		return
	}
	lyrixId = lyrixId[1:]
	parts := strings.Split(lyrixId, "@")
	if len(parts) != 2 {
		logger.Fatal("Invalid Lyrix ID. Your lyrix ID should be in the form of @someone@somedomain.abc")
		return
	}

	username := parts[0]
	host := parts[1]

	secure := true
	survey.AskOne(&survey.Confirm{Message: "Do you want to use https?"}, &secure)

	scheme := "https://"
	if !secure {
		scheme = "http://"
	}
	host = fmt.Sprintf("%s%s", scheme, host)

	password := ""
	passwordPrompt := &survey.Password{
		Message: "Enter Password:",
	}
	survey.AskOne(passwordPrompt, &password)

	jsonStr, err := json.Marshal(types.UserLoginRequest{Username: username, Password: password})
	if err != nil {
		logger.Fatal(err)
	}

	req, err := http.Post(
		fmt.Sprintf("%s/login", host),
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		logger.Fatal(err)
		return
	}

	defer req.Body.Close()

	token := &types.UserAuthGrant{}
	logger.Info(req.StatusCode)
	if req.StatusCode == http.StatusOK {

		json.NewDecoder(req.Body).Decode(token)
	} else {
		logger.Fatal("Invalid username or password.")
		return
	}
	if token.Token == "" {
		logger.Fatal("Authentication failed.")
		return
	}

	viper.Set("Username", username)
	viper.Set("AuthToken", token.Token)
	viper.Set("Host", host)

}
