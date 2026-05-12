package test

import (
	"fmt"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaim struct {
	Identity string `json:"user_identity"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

var myKey = []byte("OnlinePractice")

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
	tokenString := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkZW50aXR5IjoidXNlci0xIiwibmFtZSI6IjEyMzQ1NiJ9.GnuZb1g__v3YJ8lnJA23QBXkFqcUcC45d8pPQhQB70M"
	var u UserClaim
	claims, err := jwt.ParseWithClaims(tokenString, &u,
		func(token *jwt.Token) (interface{}, error) {
			return myKey, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	if claims.Valid {
		fmt.Println(u)
	}
}
