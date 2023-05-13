package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"
)

func main() {
	token := getGlobalValue("drone_token")
	host := getGlobalValue("drone_server")

	if token == "" {
		panic("missing drone host")
	}

	if host == "" {
		panic("missing drone token")
	}

	// create an http client with oauth authentication.
	config := new(oauth2.Config)
	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	// create the drone client with authenticator
	client := drone.NewClient(host, auther)

	// gets the current user
	user, err := client.Self()
	if err != nil {
		panic("get user failed: " + err.Error())
	}
	fmt.Printf("login user: %s", user.Login)
}

func getGlobalValue(key string) string {
	key = strings.ToUpper(key) // Convert key to uppercase

	// Check if there is an environment variable with the format "INPUT_<KEY>"
	if value := os.Getenv("INPUT_" + key); value != "" {
		return value // Return the value of the "INPUT_<KEY>" environment variable
	}

	// If the "INPUT_<KEY>" environment variable doesn't exist or is empty,
	// return the value of the "<KEY>" environment variable
	return os.Getenv(key)
}
