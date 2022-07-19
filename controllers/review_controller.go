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

func CreateReview() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			return
		}

		if !exists {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var input models.ReviewInput
		email := fmt.Sprint(vEmail)
		fmt.Println(email)
		defer cancel()

		if err := ctx.ShouldBindJSON(&input); err != nil {
			fmt.Println("Bind JSON ERROR!")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		if validationErr := validate.Struct(&input); validationErr != nil {
			fmt.Println("Validation ERROR!")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error(), "status": "failed"})
			return
		}

		newReview := models.Review{
			Id:         primitive.NewObjectID(),
			Order_Id:   input.Order_Id,
			Content:    input.Content,
			Rating:     input.Rating,
			Avatars:    input.Avatars,
			Bonus:      input.Bonus,
			ReviewedAt: time.Now(),
		}

		_, err := newReview.SaveReview(c, email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": responses.REVIEW_SUCCESS, "status": "success"})
	}
}

func GetProfileReviewByEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": responses.NOT_LOGGED_IN})
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

		res, err := models.GetReviewListByEmail(c, email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": res, "status": "success"})
	}
}
