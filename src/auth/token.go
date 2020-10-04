package auth

import (
	"errors"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var mySigningKey = []byte("secret")

type Claims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}
func CreateToken(userID primitive.ObjectID) (string, error) {
	// create the token

	expirationTime := time.Now().Add(48 * time.Hour)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		UserID: userID.Hex(),
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//Sign and get the complete encoded token as string
	return token.SignedString(mySigningKey)
}

func CheckToken(tokenStr string) (primitive.ObjectID, error) {
	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	})

	if err!=nil {
		return primitive.ObjectID{}, err
	}

	if !tkn.Valid {
		return primitive.ObjectID{}, errors.New("token not valid")
	}
	spew.Dump(claims)

	return primitive.ObjectIDFromHex(claims.UserID)
}