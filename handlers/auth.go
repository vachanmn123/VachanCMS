package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	githubapi "github.com/google/go-github/v62/github"
	"github.com/vachanmn123/vachancms/config"
	"github.com/vachanmn123/vachancms/services"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

var oauthConfig *oauth2.Config = nil

func InitAuthHandler() {
	oauthConfig = &oauth2.Config{
		ClientID:     config.Cfg.GitHubClientID,
		ClientSecret: config.Cfg.GitHubClientSecret,
		Scopes:       []string{"user:email", "repo"},
		Endpoint:     githuboauth.Endpoint,
		RedirectURL:  config.Cfg.GithubRedirectURL,
	}
}

func LoginHandler(c *gin.Context) {

	// Generate random state
	state, err := generateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	// Store state in cookie for verification
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	url := oauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	// Verify state
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	fmt.Printf("Access Token: %s\n", token.AccessToken)
	fmt.Printf("Token Type: %s\n", token.TokenType)
	fmt.Printf("Token Extra: %v\n", token.Extra("scope"))

	// Get user info
	client := githubapi.NewClient(oauthConfig.Client(context.Background(), token))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Generate JWT
	jwtToken, err := services.GenerateJWT(user.GetLogin(), token.AccessToken, config.Cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	// Set JWT in cookie
	c.SetCookie("auth_token", jwtToken, 86400, "/", "", false, true)

	// Redirect to home or frontend
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

func GetMeHandler(c *gin.Context) {
	access_token := c.GetString("user_access_token")

	client := githubapi.NewClient(oauthConfig.Client(context.Background(), &oauth2.Token{AccessToken: access_token}))
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(200, gin.H{
		"login":      user.GetLogin(),
		"name":       user.GetName(),
		"avatar_url": user.GetAvatarURL(),
	})
}

func generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
