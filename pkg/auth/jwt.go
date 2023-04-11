package auth

import (
	"fmt"
	"time"

	//"github.com/jeroendk/chatApplication/models"

	"github.com/dgrijalva/jwt-go"
)

//const defaulExpireTime = 604800 // 1 week

type JwtToken struct {
	Enabled           bool   `mapstructure:"enabled"`
	Key               string `mapstructure:"key"`
	DefaultExpireTime int64  `mapstructure:"defaultexpiretime"`
}

// KeyFunc auth key types
func (jwtToken JwtToken) KeyFunc(token *jwt.Token) (interface{}, error) {
	// nolint: gocritic
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	// hmacSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	return []byte(jwtToken.Key), nil
}

type Claims struct {
	UID       string `json:"uid"`
	SID       string `json:"sid"`
	Publish   bool   `json:"publish"`
	Subscribe bool   `json:"subscribe"`
	jwt.StandardClaims
}

func (c *Claims) GetUid() string {
	return c.UID
}

func (c *Claims) GetSID() string {
	return c.SID
}

func (c *Claims) IsPublisher() bool {
	return c.Publish
}

func (c *Claims) IsSubscriber() bool {
	return c.Subscribe
}

// CreateJWTToken generates a JWT signed token for for the given user
func CreateJWTToken(principal Principal, jwtToken *JwtToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"UID":       principal.GetUid(),
		"SID":       principal.GetSid(),
		"Publish":   principal.IsPublisher(),
		"Subscribe": principal.IsSubscriber(),
		"ExpiresAt": time.Now().Unix() + jwtToken.DefaultExpireTime,
	})
	tokenString, err := token.SignedString([]byte(jwtToken.Key))

	return tokenString, err
}

func ValidateToken(tokenString string, jwtConfig *JwtToken) (Principal, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, jwtConfig.KeyFunc)

	if err != nil {
		return Principal{}, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		p := Principal{
			UID:       claims.GetUid(),
			SID:       claims.GetSID(),
			Publish:   claims.IsPublisher(),
			Subscribe: claims.IsSubscriber(),
		}
		return p, nil
	}

	return Principal{}, fmt.Errorf("invalidated jwt token")

}
