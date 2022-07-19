package models

import (
	"context"
	"errors"
	"server/responses"
	"server/utilities"

	emailVerifier "github.com/AfterShip/email-verifier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required,min=8"`
	Avatar    string `json:"avatar"`
	Method    string `json:"method" binding:"required"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Method   string `json:"method" binding:"required"`
}

var verifier = emailVerifier.NewVerifier().EnableSMTPCheck()

type User struct {
	Id        primitive.ObjectID `json:"id" binding:"required"`             // User Index
	Firstname string             `json:"firstname" binding:"required"`      // User FirstName
	Lastname  string             `json:"lastname" binding:"required"`       // User LastName
	Username  string             `json:"username" binding:"required"`       // User_Name
	Email     string             `json:"email" binding:"required,email"`    // User Email
	Password  string             `json:"password" binding:"required,min=8"` // User Password
	Avatar    string             `json:"avatar" binding:"required"`         // User Avatar
	Method    string             `json:"method" binding:"required"`         // Login Method (google, apple, facebook, driveshare)
}

func (newUser *User) SaveUser(c context.Context) (*User, error) {

	_, err := userCollection.InsertOne(c, newUser)
	if err != nil {
		if er, ok := err.(mongo.WriteException); ok && er.WriteErrors[0].Code == 11000 {
			return nil, errors.New(responses.EMAIL_ALREADY_EXISTS)
		}
	}

	opt := options.Index()
	opt.SetUnique(true)
	index := mongo.IndexModel{Keys: bson.M{"email": 1}, Options: opt}

	if _, err := userCollection.Indexes().CreateOne(c, index); err != nil {
		return nil, errors.New("could not create index of email")
	}

	return nil, nil
}

func ValidateEmail(email string) (bool, error) {
	_, err := verifier.Verify(email)

	if err != nil {
		return false, errors.New(responses.EMAIL_VERIFY_ERROR)
	}

	return true, nil
}

func IsEmailRegistered(email string, c context.Context) bool {
	user := User{}
	err := userCollection.FindOne(c, bson.M{"email": email}).Decode(&user)
	return err == nil
}

func LoginCheck(email string, password string, method string, c context.Context) (User, error) {
	var err error

	user := User{}
	err = userCollection.FindOne(c, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return User{}, errors.New(responses.EMAIL_NOT_FOUND)
	}

	err = utilities.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return User{}, errors.New(responses.WRONG_PASSWORD)
	}

	if method != user.Method {
		return User{}, errors.New(responses.EMAIL_NOT_FOUND)
	}

	return user, nil
}

func GetUsernameByEmail(c context.Context, email string) string {
	var user User
	err := userCollection.FindOne(c, bson.M{"email": email}).Decode(&user)
	if err != nil {
		// * Meaning no match
		return ""
	}

	return user.Username
}

func GetUserAvatarByEmail(c context.Context, email string) string {
	var user User
	err := userCollection.FindOne(c, bson.M{"email": email}).Decode(&user)
	if err != nil {
		// * Meaning no match
		return ""
	}

	return user.Avatar
}

func GetUserOverallRating(c context.Context, email string) (int, float64) {
	cur, err := carCollection.Find(c, bson.M{"owner_email": email})
	if err != nil {
		return 0, 0
	}

	tripCnt := 0
	totalRating := 0.0

	defer cur.Close(c)

	for cur.Next(c) {
		var car Car
		err = cur.Decode(&car)
		if err != nil {
			continue
		}

		carTripCnt, carTotalRating, err := GetCarRating(c, car.Car_No)
		if err != nil {
			continue
		}
		tripCnt += carTripCnt
		totalRating += carTotalRating
	}

	return tripCnt, totalRating
}
