package auth

import (
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

func HashPassword(password string) (string, error) {
	hashed, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	return hashed, err
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	// create claims
	claims := jwt.RegisteredClaims{}
	claims.Issuer = "chirpy-access"
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiresIn))
	claims.Subject = userID.String() //user id as string

	// create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign the token with the secret key
	signed, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {

	// parse the token string into a jwt token struct
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (any, error) { return []byte(tokenSecret), nil })

	if err != nil {
		fmt.Print("error parsing jwt")
		return uuid.UUID{}, err
	}

	// get user id
	user, err := token.Claims.GetSubject()
	if err != nil {
		fmt.Print("error getting user id")
		return uuid.UUID{}, err
	}

	id, err := uuid.Parse(user)
	if err != nil {
		fmt.Print("error parsing user id")
		return uuid.UUID{}, err
	}

	return id, nil
}
