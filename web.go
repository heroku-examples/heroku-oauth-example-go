package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/heroku"
)

var (
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("HEROKU_OAUTH_ID"),
		ClientSecret: os.Getenv("HEROKU_OAUTH_SECRET"),
		Endpoint:     heroku.Endpoint,
		Scopes:       []string{"identity"},                                                            // See https://devcenter.heroku.com/articles/oauth#scopes
		RedirectURL:  "http://" + os.Getenv("HEROKU_APP_NAME") + "herokuapp.com/auth/heroku/callback", // See https://devcenter.heroku.com/articles/dyno-metadata
	}

	stateToken = os.Getenv("HEROKU_APP_NAME")
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<html><body><a href="/auth/heroku">Florian Otel test for OAuth with Heroku</a></body></html>`)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL(stateToken)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if state := r.FormValue("state"); state != stateToken {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", stateToken, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		// http.Error(w, "Invalid State token", http.StatusBadRequest)
		return
	}
	token, err := oauthConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		log.Printf("Code exchange failed with error: '%s'\n", err)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Received OAuth token: %#v", token)

	//// Get Heroku user information using the acquired token

	client := oauthConfig.Client(context.Background(), token)
	req, _ := http.NewRequest("GET", "https://api.heroku.com/account", nil)

	// Add the correct headers for Heroku API version 3 -- see e.g. https://devcenter.heroku.com/articles/platform-api-reference#clients
	req.Header.Add("Accept", "application/vnd.heroku+json; version=3")
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Error fetching Heroku account information: '%s'\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("===> Response Status: %s", resp.Status)
	log.Printf("===> Response Headers: %s", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	log.Printf("===> Response Body: %s", string(body))
	defer resp.Body.Close()

	var account struct { // See https://devcenter.heroku.com/articles/platform-api-reference#account
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.Unmarshal(body, &account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, `<html><body><h1>Hello %s, your Email is: %s</h1></body></html>`, account.Name, account.Email)
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/auth/heroku", handleAuth)
	http.HandleFunc("/auth/heroku/callback", handleAuthCallback)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
