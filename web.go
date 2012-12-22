package main

import (
    "code.google.com/p/oauth2"
	"os"
    "http"
)

var oauthCfg = &oauth.Config {
	ClientId: os.Getenv("HEROKU_ID"),
	ClientSecret: os.Getenv("HEROKU_SECRET"),
	AuthURL: "https://api.heroku.com/oauth/authorize",
	TokenURL: "https://api.heroku.com/oauth/token",
	RedirectURL: "http://localhost:5000/heroku/auth/callback",
}

func main() {
    http.HandleFunc("/", handleRoot)
    http.HandleFunc("/heroku/auth", handleAuth)
    http.HandleFunc("/heroku/auth/callback", handleAuthCallback)
	http.HandleFunc("/user", handleUser)
    http.ListenAndServe(":5000", nil)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	body := `<a href="/auth/heroku">Sign in with Heroku</a>`
	w.Write([]byte(body))
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
    //Get the Google URL which shows the Authentication page to the user
    url := oauthCfg.AuthCodeURL("")

    //redirect user to that page
    http.Redirect(w, r, "url", http.StatusFound)
}

// Function that handles the callback from the Google server
func handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
    //Get the code from the response
    code := r.FormValue("code")

    t := &oauth.Transport{oauth.Config: oauthCfg}

    // Exchange the received code for a token
    t.Exchange(code)

    //now get user data based on the Transport which has the token
    resp, _ := t.Client().Get(profileInfoURL)

    buf := make([]byte, 1024)
    resp.Body.Read(buf)
    userInfoTemplate.Execute(w, string(buf))
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	
}
