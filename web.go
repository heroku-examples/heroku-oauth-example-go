package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/heroku"
)

var (
	store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")), []byte(os.Getenv("COOKIE_ENCRYPT")))

	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("HEROKU_OAUTH_ID"),
		ClientSecret: os.Getenv("HEROKU_OAUTH_SECRET"),
		Endpoint:     heroku.Endpoint,
		Scopes:       []string{"identity"},                                                            // See https://devcenter.heroku.com/articles/oauth#scopes
		RedirectURL:  "http://" + os.Getenv("HEROKU_APP_NAME") + "herokuapp.com/auth/heroku/callback", // See https://devcenter.heroku.com/articles/dyno-metadata
	}

	stateToken = os.Getenv("HEROKU_APP_NAME")
)

func init() {
	gob.Register(&oauth2.Token{})

	store.MaxAge(60 * 60 * 8)
	store.Options.Secure = true
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
		http.Error(w, "Invalid State token", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, r.FormValue("code"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session, err := store.Get(r, "heroku-oauth-example-go")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["heroku-oauth-token"] = token
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/user", http.StatusFound)
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "heroku-oauth-example-go")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, ok := session.Values["heroku-oauth-token"].(*oauth2.Token)
	if !ok {
		http.Error(w, "Unable to assert token", http.StatusInternalServerError)
		return
	}
	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.heroku.com/account")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	var account struct { // See https://devcenter.heroku.com/articles/platform-api-reference#account
		Email string `json:"email"`
	}
	if err := d.Decode(&account); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, `<html><body><h1>Hello %s</h1></body></html>`, account.Email)
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/auth/heroku", handleAuth)
	http.HandleFunc("/auth/heroku/callback", handleAuthCallback)
	http.HandleFunc("/user", handleUser)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
