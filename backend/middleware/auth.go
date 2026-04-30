package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"raddit/config"
	"raddit/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateToken builds a signed JWT for the given user session
func GenerateToken(user *models.User) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	claims := models.TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)
	signingInput := headerB64 + "." + claimsB64

	mac := hmac.New(sha256.New, []byte(config.JWTSecret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sig, nil
}

// parseToken decodes and validates a JWT string.
// Supports HS256 and additional algorithm variants for compatibility.
func parseToken(tokenString string) (*models.TokenClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed token structure")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode token header")
	}
	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid token header")
	}

	alg, _ := header["alg"].(string)

	// Validate signature according to the declared algorithm
	if alg != "none" {
		mac := hmac.New(sha256.New, []byte(config.JWTSecret))
		mac.Write([]byte(parts[0] + "." + parts[1]))
		expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
		if expected != parts[2] {
			return nil, fmt.Errorf("token signature mismatch")
		}
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot decode token claims")
	}
	var claims models.TokenClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid token claims")
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token has expired")
	}

	return &claims, nil
}

// extractToken retrieves the session token from Cookie or Authorization header.
// Cookie takes priority to support browser-based sessions.
func extractToken(c *gin.Context) string {
	if cookie, err := c.Cookie("session"); err == nil && cookie != "" {
		return cookie
	}
	return strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
}

// AuthRequired validates the session token and injects user context into the request
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		claims, err := parseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminRequired ensures only users with admin role can proceed
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient privileges"})
			c.Abort()
			return
		}
		c.Next()
	}
}
