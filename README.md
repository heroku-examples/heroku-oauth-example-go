# Heroku OAuth Example: Go

Example Go application using OAuth against the Heroku API.

## Usage

First, acquire a `localhost:5000` OAuth key/secret pair from the API team. Then:

```
$ cat > .env <<EOF
HEROKU_OAUTH_ID=...
HEROKU_OAUTH_SECRET=...
COOKIE_SECRET=...
EOF
$ go get
$ foreman start
$ open http://127.0.0.1:5000
```
