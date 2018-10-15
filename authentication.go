package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/bcrypt"
)

const (
	authSecretKey  = "AUTH_SECRET"
	tokenExpiryMin = 60
	bearerTokenKey = "Bearer "
)

type AuthSvc interface {
	Initialize()
	HashPwd(pwd string) (*string, error)
	VerifyPwd(hashedPwd, pwd string) bool
	BuildToken(user User) (*string, *int64, error)
	ValidateToken(authHeader interface{}) (interface{}, error)
}

type authSvc struct {
	authSecret []byte
}

// Initialize the Auth Service.
// Get the Auth Secret out of the environment.
func (a *authSvc) Initialize() {
	secret := os.Getenv(authSecretKey)
	a.authSecret = []byte(secret)
}

// Utilize the bcrypt package to Salt and Hash the incoming password.
// Return the hashed password
func (a *authSvc) HashPwd(pwd string) (*string, error) {
	password := []byte(pwd) // convert to byte array
	// Use GenerateFromPassword to hash & salt pwd.
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashedPwd := string(hash) // convert returned hashed password to string
	return &hashedPwd, nil
}

// Given the hashed password stored for the user and the passed in password to test against,
// use the bcrypt package to compare the passwords and validate they are the same
func (a *authSvc) VerifyPwd(hashedPwd, pwd string) bool {
	storedPwd, submittedPwd := []byte(hashedPwd), []byte(pwd)     // convert both the hashed password and submitted password to byte arrays
	err := bcrypt.CompareHashAndPassword(storedPwd, submittedPwd) // compare the password byte slices for equality
	if err != nil {
		return false // passwords do not match, return false
	}
	return true // passwords match, return true
}

// Utilize the JWT library to generate a token with the given claims
// - email
// - name
// Sign the token with the auth secret
// Get the expires at timestamp: now + 60min
func (a *authSvc) BuildToken(user User) (*string, *int64, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"name":  user.Name,
	})
	signedToken, err := token.SignedString(a.authSecret) // sign the token
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()                                      // get current time
	nowPlusExpiry := now.Add(tokenExpiryMin * time.Minute) // add 60 minutes to current time to get token expiry
	nowPlusExpiryTimestamp := nowPlusExpiry.UnixNano()     // get the expiry timestamp
	return &signedToken, &nowPlusExpiryTimestamp, nil
}

// Validate the authorization token.
// Using the Authorization Header, validate that it contains a token and that the token is valid.
// If the token exists and is valid, return nil; otherwise return the error
func (a *authSvc) ValidateToken(authHeader interface{}) (interface{}, error) {
	// validate an Authorization header token is present in the request
	if authHeader == nil {
		return nil, errors.New("no valid Authorization token in request")
	}
	header := authHeader.(string)
	if header == "" {
		return nil, errors.New("no valid Authorization token in request")
	}
	// validate that it is a Bearer token
	if !strings.HasPrefix(header, bearerTokenKey) {
		return nil, errors.New("authorization token is not valid Bearer token")
	}
	t := strings.Replace(header, bearerTokenKey, "", -1)
	// parse the header token
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("there was an parsing the given token. please validate the token is for this service")
		}
		return a.authSecret, nil
	})
	if err != nil {
		return nil, err
	}
	// validate token and get claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var decodedToken interface{}
		err = mapstructure.Decode(claims, &decodedToken)
		if err != nil {
			return nil, err
		}
		return decodedToken, nil
	}
	return nil, errors.New("invalid authorization token") // token is not valid, return error
}
