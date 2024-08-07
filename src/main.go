package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"text/template"

	"github.com/apex/gateway"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var (
	clientID     = os.Getenv("GITHUB_CLIENT_ID")
	clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	callbackURL  = os.Getenv("CALLBACK_URL")
)

var (
	callbackTemplate *template.Template
	config           *oauth2.Config
)

func init() {
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     endpoints.GitHub,
		Scopes:       []string{"repo", "user"},
		RedirectURL:  callbackURL,
	}
	var err error
	callbackTemplate, err = template.New("callback").Parse(callbackHTMLTemplate)
	if err != nil {
		panic(fmt.Sprintf("Error parsing callback template: %v", err))
	}

}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	r.Get("/auth", authHandler)
	r.Get("/callback", callbackHandler)

	http.Handle("/", r)
	gateway.ListenAndServe("", nil)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	if provider := r.URL.Query().Get("provider"); provider != "github" {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
	}

	http.Redirect(w, r, config.AuthCodeURL(oauth2.GenerateVerifier()), http.StatusSeeOther)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(w, "Code query parameter is missing", http.StatusBadRequest)
		return
	}

	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error exchanging code for token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	data := struct {
		Status string
		Token  string
	}{
		Status: "success",
		Token:  token.AccessToken,
	}
	err = callbackTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
	}
	slog.Info("Successful login", "token", token.AccessToken)
}

const callbackHTMLTemplate = `
<html>
<head>
	<script>
		const status = '{{.Status}}'; 
		const token = '{{.Token}}';  
		const data = {token: token};
		window.opener.postMessage("authorizing:github", "*");
		window.opener.postMessage('authorization:github:' + status + ':' + JSON.stringify(data), '*');
	</script>
</head>
<body>
	<p>Authorizing Decap...</p>
</body>
</html>
`

type CallbackkData struct {
	Status string
	Token  string
}
