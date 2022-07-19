package controllers

import (
	"context"
	"fmt"
	"net/http"
	"server/models"
	"server/responses"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateCar() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		email, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var input models.CarInput
		user_email := fmt.Sprint(email)
		defer cancel()

		check := models.IsEmailRegistered(user_email, c)
		if !check {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.TRYING_WITH_NOT_REGISTERED_USER, "status": "failed"})
			return
		}

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		if validationErr := validate.Struct(&input); validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error(), "status": "failed"})
			return
		}

		newCar := models.Car{
			Id:          primitive.NewObjectID(),
			Car_Type_Id: primitive.NewObjectID(),
			Car_Price:   input.Car_Price,
			Car_No:      input.Car_No,
			Owner_Email: user_email,
			Car_Avatar:  input.Car_Avatar,
			From_Where:  input.Car_From,
			To_Where:    input.Car_To,
		}

		newCarType := models.CarType{
			Id:          primitive.NewObjectID(),
			Car_Brand:   input.Car_Brand,
			Car_Type:    input.Car_Type,
			Car_Seats:   input.Car_Seats,
			Car_Gearbox: input.Car_Gearbox,
			Car_Fuel:    input.Car_Fuel,
		}

		_, err := newCar.SaveCar(c, newCarType)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": responses.CAR_REGISTER_SUCCESS, "status": "success"})
	}
}

func GetProfileCarByEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		email := fmt.Sprint(vEmail)
		defer cancel()

		check := models.IsEmailRegistered(email, c)
		if !check {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.TRYING_WITH_NOT_REGISTERED_USER})
			return
		}

		res, err := models.GetCarListsByEmail(c, email)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error(), "status": "faild"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": res, "status": "success"})
	}
}
