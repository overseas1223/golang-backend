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

func CreateOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		email, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var input models.OrderInput
		defer cancel()

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		if validationErr := validate.Struct(&input); validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error(), "status": "failed"})
			return
		}

		f_t, _ := time.Parse(time.RFC3339, input.From_Time)
		t_t, _ := time.Parse(time.RFC3339, input.To_Time)

		if f_t.After(t_t) || f_t.Equal(t_t) {
			// * Meaning time error
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.TIME_ERROR, "status": "failed"})
			return
		}

		newOrder := models.Order{
			Id:           primitive.NewObjectID(),
			User_Email:   fmt.Sprint(email),
			Car_No:       input.Car_No,
			From_Time:    f_t,
			To_Time:      t_t,
			Status:       responses.STATUS_PENDING,
			Pay_Method:   input.Pay_Method,
			Ordered_Time: time.Now(),
		}

		_, err := newOrder.SaveOrder(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": responses.CREATE_ORDER_SUCCESS, "state": "success"})
	}
}

func GetProfileUserOrderByEmail() gin.HandlerFunc {
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.TRYING_WITH_NOT_REGISTERED_USER, "status": "failed"})
			return
		}

		res, err := models.GetOrderListsByEmail_User(c, email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": res, "status": "success"})
	}
}

func GetProfileBusinessOrderByEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		email := fmt.Sprint(vEmail)
		defer cancel()

		check := models.IsEmailRegistered(email, c)
		if !check {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.EMAIL_NOT_FOUND, "status": "failed"})
			return
		}

		res, err := models.GetProfileOrderListByEmail_Business(c, email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": res, "status": "success"})
	}
}

func GetProfileRevenueByEmail() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		email := fmt.Sprint(vEmail)
		defer cancel()

		check := models.IsEmailRegistered(email, c)
		if !check {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.EMAIL_NOT_FOUND, "status": "failed"})
			return
		}

		res, total, err := models.GetRevenueListsByEmail(c, email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": res, "total": total, "status": "success"})
	}
}

func CancelOrder() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		email := fmt.Sprint(vEmail)
		defer cancel()

		check := models.IsEmailRegistered(email, c)
		if !check {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.EMAIL_NOT_FOUND, "status": "failed"})
			return
		}

		orderId, err := primitive.ObjectIDFromHex(ctx.PostForm("order_id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.BAD_REQUEST, "status": "failed"})
			return
		}

		// * Check whether the order-id matches with
		isMatch := models.IsOrderIdMatchWithEmail(c, orderId, email)
		if !isMatch {
			// ! HACKING
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.HACK_ALERT, "status": "failed"})
			return
		}

		res := models.UpdateOrderStatus(c, orderId, responses.STATUS_CANCELLED)
		if !res {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": responses.UNABLE_TO_CANCEL, "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": responses.CANCEL_SUCCESS, "status": "success"})
	}
}

func GetOrderDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		vEmail, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.NOT_LOGGED_IN, "status": "failed"})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		email := fmt.Sprint(vEmail)
		defer cancel()

		check := models.IsEmailRegistered(email, c)
		if !check {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.EMAIL_NOT_FOUND, "status": "failed"})
			return
		}

		orderId, err := primitive.ObjectIDFromHex(ctx.GetHeader("orderId"))
		fmt.Println(orderId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": responses.BAD_REQUEST, "status": "failed"})
			return
		}

		res, err := models.GetOrderDetailedInformation(c, orderId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": responses.SERVER_INTERNAL_ERROR, "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": res, "status": "success"})
	}
}
