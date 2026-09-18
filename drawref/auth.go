package drawref

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type DrawRefAuth struct {
	key       paseto.V4SymmetricKey
	adminHash []byte
}

var TheAuth *DrawRefAuth

func SetupAuth(pasetoKeyHex string, adminPass string) error {
	key, err := paseto.V4SymmetricKeyFromHex(pasetoKeyHex)
	if err != nil {
		return err
	}

	adminHash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	TheAuth = &DrawRefAuth{
		key,
		adminHash,
	}

	return nil
}

type LoginParams struct {
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token  string `json:"token"`
	Level  string `json:"level"`
	Expiry string `json:"exp"`
}

func login(c *gin.Context) {
	var params LoginParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	level := ""
	if bcrypt.CompareHashAndPassword(TheAuth.adminHash, []byte(params.Password)) == nil {
		level = "admin"
	}

	if level == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		LogFailedLoginAttempt(c.ClientIP())
		return
	}

	LogSuccessfulLoginAttempt(level)

	token := paseto.NewToken()
	token.SetIssuedAt(time.Now())
	token.SetNotBefore(time.Now())
	token.SetExpiration(time.Now().Add(24 * time.Hour))

	token.Set("level", level)

	c.JSON(http.StatusOK, LoginResponse{
		Token:  token.V4Encrypt(TheAuth.key, nil),
		Level:  level,
		Expiry: time.Now().Add(24 * time.Hour).UTC().String(),
	})
}

func extractBearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("bad header value given")
	}

	jwtToken := strings.Split(header, " ")
	if len(jwtToken) != 2 {
		return "", errors.New("incorrectly formatted authorization header")
	}

	return jwtToken[1], nil
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "auth token not given"})
			return
		}

		parser := paseto.NewParser()
		parser.AddRule(paseto.NotExpired())
		parser.AddRule(paseto.ValidAt(time.Now()))

		token, err := parser.ParseV4Local(TheAuth.key, tokenString, nil)
		if err != nil {
			fmt.Println("Token error:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		level, err := token.GetString("level")
		if err != nil {
			fmt.Println("Token doesn't include 'level' claim")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		}

		if level == "admin" {
			// continue!!
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
		}
	}
}
