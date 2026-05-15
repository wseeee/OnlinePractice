package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaim struct {
	Identity string `json:"user_identity"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

var myKey = func() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		key = "Online_Practice"
	}
	return []byte(key)
}()

// 生成Token
func TestGenerateJwt(t *testing.T) {
	u := UserClaim{
		Identity: "user-1",
		Name:     "123456",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(tokenString)
}

// 解析Token
func TestAnalyseJwt(t *testing.T) {
	// 先生成一个有效的 token
	u := UserClaim{
		Identity: "user-1",
		Name:     "123456",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	tokenString, err := token.SignedString(myKey)
	if err != nil {
		t.Fatal(err)
	}

	var parsed UserClaim
	claims, err := jwt.ParseWithClaims(tokenString, &parsed,
		func(token *jwt.Token) (interface{}, error) {
			return myKey, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if claims.Valid {
		fmt.Println(parsed)
	}
}
