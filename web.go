package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"github.com/gorilla/sessions"
	"html"
	"io/ioutil"
	"net/http"
	"os"
)

// Account representation.
type Account struct {
	Email string `json:"email"`
}

var store = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))

var oauthConfig = &oauth.Config{
	ClientId:     os.Getenv("HEROKU_OAUTH_ID"),
	ClientSecret: os.Getenv("HEROKU_OAUTH_SECRET"),
	AuthURL:      "https://id.heroku.com/oauth/authorize",
	TokenURL:     "https://id.heroku.com/oauth/token",
	RedirectURL:  "http://localhost:5000/heroku/auth/callback",
}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/auth/heroku", handleAuth)
	http.HandleFunc("/auth/heroku/callback", handleAuthCallback)
	http.HandleFunc("/user", handleUser)
	http.ListenAndServe(":5000", nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	body := `<a href="/auth/heroku">Sign in with Heroku</a>`
	w.Write([]byte(body))
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	url := oauthConfig.AuthCodeURL("")
	http.Redirect(w, r, url, http.StatusFound)
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	transport := &oauth.Transport{Config: oauthConfig}
	token, err := transport.Exchange(code)
	if err != nil {
		panic(err)
	}
	session, err := store.Get(r, "heroku-oauth-example-go")
	if err != nil {
		panic(err)
	}
	session.Values["heroku-oauth-token"] = token.AccessToken
	session.Save(r, w)
	http.Redirect(w, r, "/user", http.StatusFound)
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "heroku-oauth-example-go")
	if err != nil {
		panic(err)
	}
	herokuOauthToken := session.Values["heroku-oauth-token"].(string)
	resp, err := http.Get("https://:" + herokuOauthToken + "@api.heroku.com/account")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	account := &Account{}
	if err := json.Unmarshal(responseBody, &account); err != nil {
		panic(err)
	}
	body := "Hi " + html.EscapeString(account.Email)
	w.Write([]byte(body))
}
