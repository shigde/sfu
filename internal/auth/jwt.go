package auth

import (
	"fmt"
	"time"

	//"github.com/jeroendk/chatApplication/models"

	"github.com/golang-jwt/jwt/v5"
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
	UUID string `json:"uuid"`
	jwt.RegisteredClaims
}

func (c *Claims) GetUuidString() string {
	return c.UUID
}

// CreateJWTToken generates a JWT signed token for for the given user
func CreateJWTToken(uuid string, jwtConfig *JwtToken) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uuid":      uuid,
		"ExpiresAt": time.Now().Unix() + jwtConfig.DefaultExpireTime,
		"nbf":       time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtConfig.Key))

	return tokenString, err
}

func ValidateToken(tokenString string, jwtConfig *JwtToken) (Principal, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, jwtConfig.KeyFunc)

	if err != nil {
		return Principal{}, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		p := Principal{
			UUID: claims.GetUuidString(),
		}
		return p, nil
	}

	return Principal{}, fmt.Errorf("invalidated jwt token")

}
