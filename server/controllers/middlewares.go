package controllers

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/akhilrex/hammond/service"
	"github.com/gin-gonic/gin"
)

const accessCookie = "hammond_access"
const refreshCookie = "hammond_refresh"

func bearer(c *gin.Context) string {
	header := strings.Fields(c.GetHeader("Authorization"))
	if len(header) == 2 && strings.EqualFold(header[0], "Bearer") {
		return header[1]
	}
	if len(header) == 1 {
		return header[0]
	}
	return ""
}

func secureSession(c *gin.Context) bool {
	return c.Request.TLS != nil || os.Getenv("SESSION_COOKIE_SECURE") == "1"
}

func sessionOrigin(c *gin.Context) bool {
	origin := c.GetHeader("Origin")
	if origin == "" {
		site := c.GetHeader("Sec-Fetch-Site")
		return site == "" || site == "same-origin"
	}
	parsed, err := url.Parse(origin)
	scheme := "http"
	if secureSession(c) {
		scheme = "https"
	}
	return err == nil && parsed.Scheme == scheme && strings.EqualFold(parsed.Host, c.Request.Host) &&
		parsed.User == nil && parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}

func sessionCookies(c *gin.Context, tokens service.SessionTokens, active bool) {
	accessAge, refreshAge := -1, -1
	if active {
		accessAge = int(service.AccessLifetime.Seconds())
		refreshAge = int(service.RefreshLifetime.Seconds())
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(accessCookie, tokens.Access, accessAge, "/api", "", secureSession(c), true)
	c.SetCookie(refreshCookie, tokens.Refresh, refreshAge, "/api", "", secureSession(c), true)
}

func AuthMiddleware(auto401 bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearer(c)
		if token == "" {
			token, _ = c.Cookie(accessCookie)
			switch c.Request.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
			default:
				if !sessionOrigin(c) {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
		}
		user, err := service.AccessSession(token)
		if err != nil {
			if auto401 {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			return
		}
		c.Set("userId", user.ID)
		c.Set("userModel", *user)
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}
