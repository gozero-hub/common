package tool

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JwtConfig struct {
	AccessSecret string
	AccessExpire int64
}

// GetJwtToken 获取Token
// claims:uid,email... 	自定义信息
func GetJwtToken(config JwtConfig, claims jwt.MapClaims) (string, error) {
	iat := time.Now().Unix()
	claims["iat"] = iat
	claims["exp"] = iat + config.AccessExpire
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AccessSecret))
}

// ParseToken 解析Token
func ParseToken(tokenStr, secret string) (*jwt.MapClaims, error) {
	claims := &jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil // 这里返回签名密钥
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
