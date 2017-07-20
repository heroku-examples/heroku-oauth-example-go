package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	herokuV3api "github.com/cyberdelia/heroku-go/v3"
	"github.com/gin-gonic/gin"
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

	authclient *http.Client // Pointer to the OAuth'ed http client
)

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

	router.GET("/auth/heroku/callback", func(c *gin.Context) {
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

	// Get Heroku user account information
	router.GET("/home/account", func(c *gin.Context) {
		if authclient == nil { // Not OAuth'ed yet
			c.Redirect(http.StatusPermanentRedirect, "/")
			return
		}

		// https://devcenter.heroku.com/articles/platform-api-reference#account
		var account herokuV3api.Account
		req, _ := http.NewRequest("GET", "https://api.heroku.com/account", nil)

		// Add the correct headers for Heroku API version 3 -- see e.g. https://devcenter.heroku.com/articles/platform-api-reference#clients
		req.Header.Add("Accept", "application/vnd.heroku+json; version=3")
		resp, err := authclient.Do(req)

		if err != nil {
			log.Printf("Error fetching Heroku account information: '%s'\n", err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		log.Printf("===> Response Status: %s", resp.Status)
		log.Printf("===> Response Headers: %s", resp.Header)

		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		log.Printf("===> Response Body: %s", string(body))
		defer resp.Body.Close()

		if err := json.Unmarshal(body, &account); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// c.HTML(http.StatusOK, "emptyhome.tmpl.html", nil)
		// c.String(http.StatusOK, fmt.Sprintf("Hello %s, Your account information is:\n\n", "bobo"))
		c.IndentedJSON(http.StatusOK, account)
	})

	//
	router.Run(":" + port)

	// http.HandleFunc("/auth/heroku/callback", handleAuthCallback)
}
