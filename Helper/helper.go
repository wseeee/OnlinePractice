package Helper

import (
	"crypto/md5"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

func Md5(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

type UserJwt struct {
	Identity string `json:"identity"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

var myKey = []byte("Online_Practice")

func GenerateToken(identity, name string) (string, error) {
	u := UserJwt{
		Identity: identity,
		Name:     name,
	}
	tokenstring := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	token, err := tokenstring.SignedString(myKey)
	if err != nil {
		log.Print("Failed to generate token: %v", err)
	}
	return token, nil
}

func AnalyseToken(token string) (*UserJwt, error) {
	user := new(UserJwt)
	claims, err := jwt.ParseWithClaims(token, user, func(token *jwt.Token) (interface{}, error) {
		return myKey, nil
	})
	if err != nil {
		log.Print("Failed to analyse token: %v", err)
	}
	if claims.Valid {
		return nil, fmt.Errorf("Analyse Token Error")
	}
	return user, nil
}
