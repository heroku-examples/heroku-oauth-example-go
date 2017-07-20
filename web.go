package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/heroku"
)

type myjson struct {
	Id            string `json:"id"`
	Issued_at     string `json:"issued_at"`
	Scope         string `json:"scope"`
	Instance_url  string `json:"instance_url"`
	Token_type    string `json:"token_type"`
	Refresh_token string `json:"refresh_token"`
	Id_token      string `json:"id_token"`
	Signature     string `json:"signature"`
	Access_token  string `json:"access_token"`
}

var (
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("HEROKU_OAUTH_ID"),
		ClientSecret: os.Getenv("HEROKU_OAUTH_SECRET"),
		Endpoint:     heroku.Endpoint,
		Scopes:       []string{"identity"},                                                            // See https://devcenter.heroku.com/articles/oauth#scopes
		RedirectURL:  "http://" + os.Getenv("HEROKU_APP_NAME") + "herokuapp.com/auth/heroku/callback", // See https://devcenter.heroku.com/articles/dyno-metadata
	}

	stateToken = os.Getenv("HEROKU_APP_NAME")

	jsondata = myjson{Id: "My Id", Issued_at: "My Issued at", Scope: "My Scope", Instance_url: "My Instace_URL", Token_type: "My Token Type", Refresh_token: "My Refresh token", Id_token: "My ID token", Signature: "My Signature", Access_token: "My access token"}

	authclient *http.Client // Pointer to the OAuth'ed http client
)

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

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.Default()
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	router.GET("/auth/heroku", func(c *gin.Context) {
		oauthurl := oauthConfig.AuthCodeURL(stateToken)
		c.Redirect(http.StatusPermanentRedirect, oauthurl)
	})

	router.GET("/auth/callback", func(c *gin.Context) {
		log.Print("===> Callback received")
		state := c.Query("state") // shortcut for c.Request.URL.Query().Get("state")
		if state != stateToken {
			log.Printf("invalid oauth state, expected '%s', got '%s'\n", stateToken, state)
			c.Redirect(http.StatusPermanentRedirect, "/")
			return
		}

		token, err := oauthConfig.Exchange(oauth2.NoContext, c.Query("code"))
		if err != nil {
			log.Printf("Code exchange failed with error: '%s'\n", err)
			return
		}
		log.Printf("Received OAuth token: %#v", token)

		authclient = oauthConfig.Client(context.Background(), token) // Save the OAuth'ed http client

		c.Redirect(http.StatusPermanentRedirect, "/home")
		return
	})

	// Dispatcher page for Heroku API actions
	router.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.tmpl.html", nil)
	})

	router.GET("/home/account", func(c *gin.Context) {

		if authclient == nil { // Not OAuth'ed yet
			// c.HTML(http.StatusOK, "emptyhome.tmpl.html", nil)
			c.Redirect(http.StatusPermanentRedirect, "/")
			return
		}

		// c.HTML(http.StatusOK, "emptyhome.tmpl.html", nil)
		// c.String(http.StatusOK, fmt.Sprintf("Hello %s, Your account information is:\n\n", "bobo"))
		c.IndentedJSON(http.StatusOK, jsondata)
	})

	//
	router.Run(":" + port)

	// http.HandleFunc("/auth/heroku/callback", handleAuthCallback)
}
