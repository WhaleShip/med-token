package entity

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UID        string `json:"uid"`
	IP         string `json:"ip"`
	RefreshJTI string `json:"rjti"`
	jwt.RegisteredClaims
}
