package services

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"server/configs"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

/*
{
    "web": {
        "client_id":"112484690005-g4a02edc69aq4pv4v6da1tq54laccqr2.apps.googleusercontent.com",
        "project_id":"driveshare-351800",
        "auth_uri":"https://accounts.google.com/o/oauth2/auth",
        "token_uri":"https://oauth2.googleapis.com/token",
        "auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs",
		"client_secret":"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",
        "redirect_uris":["https://127.0.0.1:5000/signin-google"],
        "javascript_origins":["https://127.0.0.1:5000"]
    }
}
*/

var (
	oauthConfGl = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:5000/signin-google",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	oauthStateStringGl = ""
)

func InitializeOAuthGoogle() {
	config, _ := configs.LoadConfig(".")
	oauthConfGl.ClientID = config.GoogleClientId
	oauthConfGl.ClientSecret = config.GoogleClientSecret
	oauthStateStringGl = config.OAuthStateString
}

func HandleGoogleLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		email := ctx.Param("email")
		InitializeOAuthGoogle()
		HandleLogin(ctx, oauthConfGl, oauthStateStringGl, email)
	}
}

func CallBackFromGoogle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		state := ctx.Request.FormValue("state")
		if state != oauthStateStringGl {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid oauth state, expected " + oauthStateStringGl + ", got " + state})
			return
		}

		code := ctx.Request.FormValue("code")
		if code == "" {
			reason := ctx.Request.FormValue("error_reason")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": reason})
		} else {
			token, err := oauthConfGl.Exchange(ctx, code)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer resp.Body.Close()

			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			ctx.JSON(http.StatusOK, gin.H{"success": string(response)})
		}
	}
}
