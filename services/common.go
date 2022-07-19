package services

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func HandleLogin(ctx *gin.Context, oauthConf *oauth2.Config, oauthStateString string, email string) {
	URL, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		return
	}
	parameters := url.Values{}
	parameters.Add("client_id", oauthConf.ClientID)
	parameters.Add("scope", "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.email")
	parameters.Add("redirect_uri", oauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", oauthStateString)
	parameters.Add("login_hint", email)
	parameters.Add("display", "popup")
	URL.RawQuery = parameters.Encode()
	url := URL.String()
	ctx.Redirect(http.StatusTemporaryRedirect, url)
}
