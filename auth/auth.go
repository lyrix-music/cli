package auth

import (
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lyrix-music/cli/types"
	"github.com/spf13/viper"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stderr)

func Login() {
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

func Register() {
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
		Message: "Create a strong password:",
	}
	survey.AskOne(passwordPrompt, &password)

	if len(password) < 8 {
		logger.Fatal("Create a password with more than 8 characters.")
		return
	}

	confirmPassword := ""
	confirmPasswordPrompt := &survey.Password{
		Message: "Re-enter your password:",
	}
	survey.AskOne(confirmPasswordPrompt, &confirmPassword)
	if password != confirmPassword {
		logger.Fatal("Passwords do not match.")
	}

	telegramId := 0
	telegramIdPrompt := &survey.Input{Message: "Enter your telegram ID (send /mytelegramid to Lyrix bot):"}
	survey.AskOne(telegramIdPrompt, &telegramId)

	if telegramId == 0 {
		logger.Fatal("Please enter a valid telegram Id. Telegram IDs are unsigned integers. (eg: 123456789)")
	}

	jsonStr, err := json.Marshal(types.UserRegisterRequest{Username: username, Password: password, TelegramId: telegramId})
	if err != nil {
		logger.Fatal(err)
	}

	req, err := http.Post(
		fmt.Sprintf("%s/register", host),
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		logger.Fatal(err)
		return
	}

	defer req.Body.Close()

	logger.Info(req.StatusCode)
	if req.StatusCode != http.StatusOK && req.StatusCode != http.StatusAccepted {
		logger.Fatal("Invalid username or password. Registration failed. Contact your server admin for more details")
	}
	logger.Info("Registration successful. Now, use 'lyrixd' command to login to lyrix.")
	logger.Info("To logout, use 'lyrixd reset-config'.")
}
