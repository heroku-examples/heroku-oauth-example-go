# Heroku OAuth Example: Go

Example Go application that uses the Heroku OAuth web flow for authentication.

The [OAuth developer documentation](https://devcenter.heroku.com/articles/oauth) has additional resources.

## Platform Installation

```
$ heroku create go-heroku-oauth-example-$USER
$ heroku labs:enable runtime-dyno-metadata
$ heroku plugins:install heroku-cli-oauth
$ heroku clients:create  "Go OAuth Example ($USER)" https://go-heroku-oauth-example-$USER.herokuapp.com/auth/heroku/callback
$ heroku config:add HEROKU_OAUTH_ID=     # set to `id` from command output above
$ heroku config:add HEROKU_OAUTH_SECRET= # set to `secret` from command output above
$ heroku config:add COOKIE_SECRET=`openssl rand -hex 32`
$ heroku config:add COOKIE_ENCRYPT=`openssl rand -hex 16`
$ git push heroku master
$ heroku open
```
