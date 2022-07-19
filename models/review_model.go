package models

import (
	"context"
	"errors"
	"math"
	"server/responses"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReviewInput struct {
	Order_Id primitive.ObjectID `json:"order_id" binding:"required"`
	Content  string             `json:"content" binding:"required"`
	Rating   float64            `json:"rating" binding:"required"`
	Bonus    int                `json:"bonus"`
	Avatars  []string           `json:"avatars"`
}

type Review struct {
	Id         primitive.ObjectID `json:"id" binding:"required"`
	Order_Id   primitive.ObjectID `json:"order_id" binding:"required"`    // Order Id
	Content    string             `json:"content" binding:"required"`     // Content of Review
	Rating     float64            `json:"rating" binding:"required"`      // Rating
	Avatars    []string           `json:"avatars"`                        // Review Pictures of Car
	Bonus      int                `json:"bonus" binding:"required"`       // Reviewed Bonus
	ReviewedAt time.Time          `json:"reviewed_at" binding:"required"` // Reviewed time
}

func GetBonusByOrderId(c context.Context, orderId primitive.ObjectID) int {
	var review Review
	err := reviewCollection.FindOne(c, bson.M{"order_id": orderId}).Decode(&review)
	if err != nil {
		// * Meaning not reviewed
		return 0
	}
	return review.Bonus
}

func IsReviewedToday(c context.Context, orderId primitive.ObjectID) bool {
	var review Review
	err := reviewCollection.FindOne(c, bson.M{"order_id": orderId}).Decode(&review)
	if err != nil {
		// * Meaning not reviewed
		return false
	}

	if math.Abs(time.Until(review.ReviewedAt).Hours()) <= 24.0 && review.ReviewedAt.Day() == time.Now().Day() {
		return true
	}
	return false
}

func (newReview *Review) SaveReview(c context.Context, email string) (*Review, error) {
	// * Check whether order exists or not
	isOrderExist := IsOrderExistByOrderId(c, newReview.Order_Id)
	if !isOrderExist {
		return nil, errors.New(responses.NOT_ORDER_EXIST)
	}

	// * Check whether order matches with email
	isOrderMatchWithEmail := IsOrderIdMatchWithEmail(c, newReview.Order_Id, email)
	if !isOrderMatchWithEmail {
		return nil, errors.New(responses.ORDER_UNMATCH_WITH_EMAIL)
	}

	// * Check whether reviewed already or not
	var review Review
	err := reviewCollection.FindOne(c, bson.M{"order_id": newReview.Order_Id}).Decode(&review)
	if err == nil {
		// !!! Hack catch
		return nil, errors.New(responses.HACK_ALERT)
	}

	// * Update Order status to Completed
	isUpdated := UpdateOrderStatus(c, newReview.Order_Id, responses.STATUS_COMPLETED)
	if !isUpdated {
		return nil, errors.New(responses.UNABLE_TO_UPDATE)
	}

	_, err = reviewCollection.InsertOne(c, newReview)
	return nil, err
}

func GetReviewListByEmail(c context.Context, email string) ([]responses.Profile_Review_Response, error) {
	// ! Need to use aggregate method here
	var result []responses.Profile_Review_Response

	// * Get Car information of the owner of email
	var carLists []responses.Profile_Car_Response
	var err error
	carLists, err = GetCarListsByEmail(c, email)
	if err != nil {
		// * Meaning Empty car list
		return nil, nil
	}

	// * Get Result with car list
	for _, car := range carLists {
		reviewList, err := GetReviewListByCarNo(c, car.Car_No)
		if err != nil {
			// * Meaning empty review list for this car
			continue
		}

		// * Get Profile Review List on this car
		for _, review := range reviewList {

			// * Get reviewed user_email
			user_email := GetUserEmailById(c, review.Order_Id)

			var seg responses.Profile_Review_Response
			seg.UserAvatar = GetUserAvatarByEmail(c, user_email)
			seg.CarAvatars = review.Avatars
			seg.Content = review.Content
			seg.Rating = review.Rating
			seg.When = int(math.Ceil((time.Since(review.ReviewedAt)).Hours() / 24))
			seg.ReviewedAt = review.ReviewedAt
			seg.Username = GetUsernameByEmail(c, user_email)
			result = append(result, seg)
		}
	}
	// * Sort by When
	sort.Slice(result, func(i, j int) bool {
		if i >= j {
			return result[i].ReviewedAt.Before(result[j].ReviewedAt)
		} else {
			return result[i].ReviewedAt.After(result[j].ReviewedAt)
		}
	})
	return result, nil
}

func GetReviewListByCarNo(c context.Context, carNo string) ([]Review, error) {
	var result []Review

	cur, err := orderCollection.Find(c, bson.M{"car_no": carNo})
	if err != nil {
		// * Meaning no match
		return result, err
	}

	defer cur.Close(c)

	for cur.Next(c) {
		var order Order
		err = cur.Decode(&order)
		if err != nil {
			continue
		}

		if order.Status == responses.STATUS_CANCELLED {
			continue
		}

		var review Review
		err = reviewCollection.FindOne(c, bson.M{"order_id": order.Id}).Decode(&review)
		if err != nil {
			continue
		}

		result = append(result, review)
	}

	return result, nil
}
