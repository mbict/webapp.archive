package jwt

//https://github.com/brainattica/golang-jwt-authentication-api-sample
//https://github.com/brainattica/golang-jwt-authentication-api-sample/tree/master/services

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	expireTime = 86400
)

type BackendOptions struct {
	ExpireTime     int
	PublicKeyFile  string
	PrivateKeyFile string
}

type Backend struct {
	expireTime int
	signKey    *rsa.PrivateKey
	verifyKey  *rsa.PublicKey
}

func NewBackend(options BackendOptions) *Backend {
	signBytes, err := ioutil.ReadFile(options.PrivateKeyFile)
	if err != nil {
		panic(err)
	}
	
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(options.PublicKeyFile)
	if err != nil {
		panic(err)
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

	var expireTime = expireTime
	if options.ExpireTime <= 0 {
		expireTime = options.ExpireTime
	}

	return &Backend{
		expireTime: expireTime,
		signKey:    signKey,
		verifyKey:  verifyKey,
	}
}

func (backend *Backend) GenerateToken(userUUID interface{}, claims ...map[string]interface{}) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)

	for _, claim := range claims {
		for key, value := range claim {
			token.Claims[key] = value
		}
	}

	token.Claims["exp"] = time.Now().Add(time.Second * time.Duration(backend.expireTime)).Unix()
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["sub"] = userUUID

	tokenString, err := token.SignedString(backend.signKey)
	if err != nil {
		panic(err)
		return "", err
	}
	return tokenString, nil
}

func (backend *Backend) validationKey(token *jwt.Token) (interface{}, error) {
	return backend.verifyKey, nil
}

func (backend *Backend) ValidateToken(jwtToken string) (*jwt.Token, error) {
	//parse token
	token, err := jwt.Parse(jwtToken, backend.validationKey)

	// branch out into the possible error from signing
	switch err.(type) {
	case nil: // no error

		if !token.Valid { // but may still be invalid
			return nil, fmt.Errorf("Invalid/malformed token")
		}

		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok { //invalid token sign method
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		//log.Printf("Someone accessed resricted area! Token:%+v\n", token)
		return token, nil

	case *jwt.ValidationError: // something was wrong during the validation
		vErr := err.(*jwt.ValidationError)

		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			return nil, fmt.Errorf("Token expired, aquire a new token")

		default:
			return nil, fmt.Errorf("Error token %+v\n", vErr.Errors)
		}

	default: // something else went wrong
		return nil, fmt.Errorf("Error parsing token %v\n", err)
	}
}