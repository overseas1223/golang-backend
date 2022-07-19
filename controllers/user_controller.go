package controllers

import (
	"context"
	"net/http"
	"server/configs"
	"server/models"
	"server/responses"
	"server/utilities"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var input models.RegisterInput
		defer cancel()

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if validationErr := validate.Struct(&input); validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if _, err := models.ValidateEmail(input.Email); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		hashedPassword, _ := utilities.HashPassword(input.Password)
		newUser := models.User{
			Id:        primitive.NewObjectID(),
			Firstname: input.Firstname,
			Lastname:  input.Lastname,
			Username:  input.Username,
			Email:     input.Email,
			Password:  hashedPassword,
			Avatar:    input.Avatar,
			Method:    input.Method,
		}

		_, err := newUser.SaveUser(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"data": responses.SIGN_UP_SUCCESS, "status": "success"})
	}
}

func SignIn() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var input models.LoginInput
		defer cancel()

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		user, err := models.LoginCheck(input.Email, input.Password, input.Method, c)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		config, _ := configs.LoadConfig(".")
		access_token, err := utilities.CreateToken(config.AccessTokenExpiresIn, user.Email, config.AccessTokenPrivateKey)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": "failed"})
			return
		}

		// refresh_token, err := utilities.CreateToken(config.RefreshTokenExpiresIn, user.Id, config.RefreshTokenPrivateKey)
		// if err != nil {
		// 	ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// }
		/*
			ctx.SetCookie("access_token", access_token, config.AccessTokenMaxAge*60, "/", "localhost", false, true)
			ctx.SetCookie("refresh_token", refresh_token, config.RefreshTokenMaxAge*60, "/", "localhost", false, true)
			ctx.SetCookie("logged_in", "true", config.AccessTokenMaxAge*60, "/", "localhost", false, false)*/
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "access_token": access_token})
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"Hello": "Welcome"})
	}
}

/*
func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var user models.User
		defer cancel()

		// validate the request body
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// user the validate library to validate required fields
		if validationErr := validate.Struct(&user); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": validationErr.Error(),
				},
			})
			return
		}

		newUser := models.User{
			Id:       primitive.NewObjectID(),
			Name:     user.Name,
			Location: user.Location,
			Title:    user.Title,
		}

		result, err := userCollection.InsertOne(ctx, newUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusCreated, responses.UserResponse{
			Status:  http.StatusCreated,
			Message: "success",
			Data: map[string]interface{}{
				"data": result,
			},
		})
	}
}

func GetAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		var user models.User
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(userId)

		err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": user,
			},
		})
	}
}

func EditAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		var user models.User
		defer cancel()
		objId, _ := primitive.ObjectIDFromHex(userId)

		// validate the request body
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// use the validator library to validate required fields
		if validationErr := validate.Struct(&user); validationErr != nil {
			c.JSON(http.StatusBadRequest, responses.UserResponse{
				Status:  http.StatusBadRequest,
				Message: "error",
				Data: map[string]interface{}{
					"data": validationErr.Error(),
				},
			})
			return
		}

		update := bson.M{
			"name":     user.Name,
			"location": user.Location,
			"title":    user.Title,
		}
		result, err := userCollection.UpdateOne(ctx, bson.M{"id": objId}, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// get updated user details
		var updatedUser models.User
		if result.MatchedCount == 1 {
			err := userCollection.FindOne(ctx, bson.M{"id": objId}).Decode(&updatedUser)
			if err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data: map[string]interface{}{
						"data": err.Error(),
					},
				})
				return
			}
		}

		c.JSON(http.StatusOK, responses.UserResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": updatedUser,
			},
		})
	}
}

func DeleteAUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		userId := c.Param("userId")
		defer cancel()

		objId, _ := primitive.ObjectIDFromHex(userId)

		result, err := userCollection.DeleteOne(ctx, bson.M{"id": objId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		if result.DeletedCount < 1 {
			c.JSON(http.StatusNotFound, responses.UserResponse{
				Status:  http.StatusNotFound,
				Message: "error",
				Data: map[string]interface{}{
					"data": "User with specified ID not found.",
				},
			})
			return
		}

		c.JSON(http.StatusOK, responses.UserResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": "User successfully deleted.",
			},
		})
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var users []models.User
		defer cancel()

		results, err := userCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, responses.UserResponse{
				Status:  http.StatusInternalServerError,
				Message: "error",
				Data: map[string]interface{}{
					"data": err.Error(),
				},
			})
			return
		}

		// reading from the db in an optimal way
		defer results.Close(ctx)
		for results.Next(ctx) {
			var singleUser models.User
			if err := results.Decode(&singleUser); err != nil {
				c.JSON(http.StatusInternalServerError, responses.UserResponse{
					Status:  http.StatusInternalServerError,
					Message: "error",
					Data: map[string]interface{}{
						"data": err.Error(),
					},
				})
			}
			users = append(users, singleUser)
		}

		c.JSON(http.StatusOK, responses.UserResponse{
			Status:  http.StatusOK,
			Message: "success",
			Data: map[string]interface{}{
				"data": users,
			},
		})
	}
}
*/
