package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/heroku"
)

// Account representation.
type Account struct {
	Email string `json:"email"`
}

var (
	store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))

	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("HEROKU_OAUTH_ID"),
		ClientSecret: os.Getenv("HEROKU_OAUTH_SECRET"),
		Endpoint:     heroku.Endpoint,
		RedirectURL:  "http://" + os.Getenv("HEROKU_APP_NAME") + "herokuapp.com/auth/heroku/callback",
	}

	stateToken string
)

func main() {
	stateToken = os.Getenv("OAUTH_STATE_TOKEN")
	if stateToken == "" {
		stateToken = string(securecookie.GenerateRandomKey(32))
	}

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/auth/heroku", handleAuth)
	http.HandleFunc("/auth/heroku/callback", handleAuthCallback)
	http.HandleFunc("/user", handleUser)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, `<html><body><a href="/auth/heroku">Sign in with Heroku</a></body></html>`)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL(stateToken)
	http.Redirect(w, r, url, http.StatusFound)
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if v := r.FormValue("state"); v != stateToken {
		// TODO: Handle invalid state token
		panic("Wrong state token")
	}
	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		// TODO: Handle err
		panic(err)
	}
	session, err := store.Get(r, "heroku-oauth-example-go")
	if err != nil {
		panic(err)
	}
	session.Values["heroku-oauth-token"] = token
	if err := session.Save(r, w); err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/user", http.StatusFound)
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "heroku-oauth-example-go")
	if err != nil {
		panic(err)
	}
	token, ok := session.Values["heroku-oauth-token"].(*oauth2.Token)
	if !ok {
		panic("Unable to assert token")
	}
	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.heroku.com/account")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	var account Account
	if err := json.Unmarshal(responseBody, &account); err != nil {
		panic(err)
	}
	fmt.Fprintf(w, `<html><body><h1>Hello %s</h1></body></html>`, account.Email)
}
