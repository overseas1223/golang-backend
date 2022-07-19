package models

import (
	"context"
	"errors"
	"fmt"
	"math"
	"server/responses"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderInput struct {
	Car_No     string `json:"car_no" binding:"required"`
	From_Time  string `json:"from_time" binding:"required"`
	To_Time    string `json:"to_time" binding:"required"`
	Pay_Method string `json:"pay_method" binding:"required"`
}

type Order struct {
	Id           primitive.ObjectID `json:"id" binding:"required"`
	User_Email   string             `json:"user_email" binding:"required"`   // Id of user who makes order
	Car_No       string             `json:"car_no" binding:"required"`       // car id number what user orders
	From_Time    time.Time          `json:"from_time" binding:"required"`    // pick up time
	To_Time      time.Time          `json:"to_time" binding:"required"`      // drop off time
	Status       string             `json:"status" binding:"required"`       // current order status
	Ordered_Time time.Time          `json:"ordered_time" binding:"required"` // order created date
	Pay_Method   string             `json:"pay_method" binding:"required"`   // payment method : credit card, crypto
}

func GetUserEmailById(c context.Context, id primitive.ObjectID) string {
	var order Order
	err := orderCollection.FindOne(c, bson.M{"id": id}).Decode(&order)
	if err != nil {
		return ""
	}
	return order.User_Email
}

func IsCarOrderedByOther(c context.Context, newOrder *Order) bool {
	// * Get List of orders of the car and check the time
	cur, err := orderCollection.Find(c, bson.M{"car_no": newOrder.Car_No})
	if err != nil {
		// * Meaning nobody ordered that car
		return false
	}

	defer cur.Close(c)

	for cur.Next(c) {
		var order Order
		err := cur.Decode(&order)
		if err != nil || order.Status != responses.STATUS_PENDING {
			continue
		}
		// * Check Pending Order Lists
		check1 := !(order.From_Time.Before(newOrder.From_Time)) && !(order.To_Time.After(newOrder.To_Time))
		check2 := !(order.From_Time.Before(newOrder.From_Time)) && !(order.To_Time.Before(newOrder.To_Time)) && !(order.From_Time.After(newOrder.To_Time))
		check3 := !(order.From_Time.After(newOrder.From_Time)) && !(order.To_Time.Before(newOrder.To_Time)) && !(order.To_Time.After(newOrder.To_Time))
		check4 := !(order.From_Time.After(newOrder.From_Time)) && !(order.To_Time.Before(newOrder.To_Time))
		if check1 || check2 || check3 || check4 {
			// * Meaning intersection exists
			return true
		}
	}
	return false
}

func IsOrderedByThatTime(c context.Context, newOrder *Order) bool {
	// * Get List of ordered lists of the user and check the time
	cur, err := orderCollection.Find(c, bson.M{"user_email": newOrder.User_Email})
	if err != nil {
		// * Meaning empty order list
		return false
	}

	defer cur.Close(c)

	for cur.Next(c) {
		var order Order
		err := cur.Decode(&order)
		if err != nil || order.Status != responses.STATUS_PENDING {
			continue
		}

		// * Check Pending Order Lists
		check1 := !(order.From_Time.Before(newOrder.From_Time)) && !(order.To_Time.After(newOrder.To_Time))
		check2 := !(order.From_Time.Before(newOrder.From_Time)) && !(order.To_Time.Before(newOrder.To_Time)) && !(order.From_Time.After(newOrder.To_Time))
		check3 := !(order.From_Time.After(newOrder.From_Time)) && !(order.To_Time.Before(newOrder.To_Time)) && !(order.To_Time.After(newOrder.To_Time))
		check4 := !(order.From_Time.After(newOrder.From_Time)) && !(order.To_Time.Before(newOrder.To_Time))
		if check1 || check2 || check3 || check4 {
			// * Meaning intersection exists
			return true
		}
	}

	return false
}

func (newOrder *Order) SaveOrder(c context.Context) (*Order, error) {
	// * Check whether car is registered or not
	if !IsCarNoExist(c, newOrder.Car_No) {
		return nil, errors.New(responses.NOT_REGISTERED_CAR)
	}
	// * Check whether others ordered the car at that time or not
	isOrdered := IsCarOrderedByOther(c, newOrder)
	if isOrdered {
		return nil, errors.New(responses.ORDERED_BY_OTHERS)
	}

	// * Check whether the user has another schedule at that time or not
	isBusy := IsOrderedByThatTime(c, newOrder)
	if isBusy {
		return nil, errors.New(responses.ORDERED_SAME_TIME)
	}

	_, err := orderCollection.InsertOne(c, newOrder)
	return nil, err
}

func GetOrderListsByEmail_User(c context.Context, email string) ([]responses.Profile_User_Order_Response, error) {
	cur, err := orderCollection.Find(c, bson.M{"user_email": email})
	if err != nil {
		return nil, err
	}

	defer cur.Close(c)

	var result []responses.Profile_User_Order_Response

	for cur.Next(c) {
		var res Order
		err := cur.Decode(&res)
		if err != nil {
			continue
		}

		// need to aggregate
		curPPD, mErr := GetCarPriceByNo(c, res.Car_No)
		if mErr != nil {
			continue
		}

		days := math.Ceil(res.To_Time.Sub(res.From_Time).Hours() / 24.0)
		total := curPPD * days

		var bonus int
		if res.Status == responses.STATUS_COMPLETED {
			bonus = GetBonusByOrderId(c, res.Id)
		}

		ownerEmail, _ := GetOwnerEmailByCarNo(c, res.Car_No)

		curOrder := responses.Profile_User_Order_Response{
			Id:           res.Id,
			Owner_Email:  ownerEmail,
			Car_No:       res.Car_No,
			Status:       res.Status,
			Total:        total + float64(bonus),
			Ordered_Time: res.Ordered_Time,
		}
		result = append(result, curOrder)
	}

	return result, nil
}

func GetProfileOrderListByEmail_Business(c context.Context, email string) ([]responses.Profile_Business_Order_Response, error) {
	// ! Need to use aggregate method here
	var result []responses.Profile_Business_Order_Response

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
		orderList, err := GetOrderListByCarNo(c, car.Car_No)
		if err != nil {
			// * Meaning Empty order list for this car
			continue
		}

		// * Get Profile Order List on this car
		for _, order := range orderList {
			if order.Status == responses.STATUS_CANCELLED {
				continue
			}

			// * Separate Status into Upcoming, Ongoing, Completed
			var segStatus string
			if order.Status == responses.STATUS_COMPLETED {
				segStatus = responses.STATUS_COMPLETED
			} else {
				// * Check whether order started or not
				if order.From_Time.After(time.Now()) {
					segStatus = responses.STATUS_ONGOING
				} else {
					segStatus = responses.STATUS_UPCOMING
				}
			}

			var seg responses.Profile_Business_Order_Response
			seg.UserAvatar = GetUserAvatarByEmail(c, order.User_Email)
			seg.Username = GetUsernameByEmail(c, order.User_Email)
			seg.CarDes = car.Car_Brand + " " + car.Car_Type
			seg.From_Time = order.From_Time
			seg.To_Time = order.To_Time
			seg.Status = segStatus
			result = append(result, seg)
		}
	}

	// * Need to sort the result
	// ? But what is the index?
	// * Sort by From_Time
	sort.Slice(result, func(i, j int) bool {
		if i >= j {
			return result[i].From_Time.Before(result[j].From_Time)
		} else {
			return result[i].From_Time.After(result[j].From_Time)
		}
	})
	return result, nil
}

func GetOrderListByCarNo(c context.Context, carNo string) ([]Order, error) {
	var result []Order

	cur, err := orderCollection.Find(c, bson.M{"car_no": carNo})
	if err != nil {
		// * Meaning no match
		return result, err
	}

	defer cur.Close(c)

	// * Need to grab all matches
	for cur.Next(c) {
		var ord Order
		err := cur.Decode(&ord)
		if err != nil {
			continue
		}
		result = append(result, ord)
	}
	return result, nil
}

func GetRevenueListsByEmail(c context.Context, email string) ([]responses.Profile_Revenue_Response, float64, error) {
	// ! Need to use aggregate method here
	var result []responses.Profile_Revenue_Response

	// * Get Car information of the owner of email
	var carLists []responses.Profile_Car_Response
	var err error
	carLists, err = GetCarListsByEmail(c, email)
	if err != nil {
		// * Meaning Empty car list
		return nil, 0, nil
	}

	var earnToday float64
	// * Get Result with car list
	for _, car := range carLists {
		orderList, err := GetOrderListByCarNo(c, car.Car_No)
		if err != nil {
			// * No Order List found for this car
			continue
		}

		curPPD, mErr := GetCarPriceByNo(c, car.Car_No)
		if mErr != nil {
			continue
		}

		for _, o := range orderList {

			if o.Status == responses.STATUS_CANCELLED {
				continue
			}

			days := math.Ceil(o.To_Time.Sub(o.From_Time).Hours() / 24.0)
			total := curPPD * days

			var bonus int
			if o.Status == responses.STATUS_COMPLETED {
				bonus = GetBonusByOrderId(c, o.Id)
				if IsReviewedToday(c, o.Id) {
					earnToday += (float64(bonus) + total)
				}
			}

			newOne := responses.Profile_Revenue_Response{
				Username: GetUsernameByEmail(c, email),
				Status:   o.Status,
				Car_No:   car.Car_No,
				Budget:   total + float64(bonus),
			}

			result = append(result, newOne)
		}
	}

	return result, earnToday, nil
}

func UpdateOrderStatus(c context.Context, orderId primitive.ObjectID, newStatus string) bool {
	var curOrder Order

	oErr := orderCollection.FindOne(c, bson.M{"id": orderId}).Decode(&curOrder)

	if oErr != nil {
		// * Meaning not exist
		return false
	}
	fmt.Println(orderId, newStatus)
	_, err := orderCollection.ReplaceOne(
		c,
		bson.M{"id": orderId},
		bson.M{
			"id":           curOrder.Id,
			"user_email":   curOrder.User_Email,
			"car_no":       curOrder.Car_No,
			"from_time":    curOrder.From_Time,
			"to_time":      curOrder.To_Time,
			"status":       newStatus,
			"ordered_time": curOrder.Ordered_Time,
			"pay_method":   curOrder.Pay_Method,
		},
	)
	return err == nil
}

func IsOrderExistByOrderId(c context.Context, orderId primitive.ObjectID) bool {
	var order Order
	err := orderCollection.FindOne(c, bson.M{"id": orderId}).Decode(&order)
	return err == nil
}

func IsOrderIdMatchWithEmail(c context.Context, orderId primitive.ObjectID, email string) bool {
	var order Order
	err := orderCollection.FindOne(c, bson.M{"id": orderId}).Decode(&order)
	if err != nil {
		return false
	}

	return order.User_Email == email
}

func GetCarNoByOrderId(c context.Context, orderId primitive.ObjectID) string {
	var order Order
	err := orderCollection.FindOne(c, bson.M{"id": orderId}).Decode(&order)
	if err != nil {
		return ""
	}
	return order.Car_No
}

func GetOrderDetailedInformation(c context.Context, orderId primitive.ObjectID) (responses.Order_Response, error) {

	car_no := GetCarNoByOrderId(c, orderId)
	owner_email, _ := GetOwnerEmailByCarNo(c, car_no)

	car_avatar, car_brand_type, from_where, to_where, daily_price := GetCarDetailsByNo(c, car_no)
	owner_name := GetUsernameByEmail(c, owner_email)
	total_trip, total_rating := GetUserOverallRating(c, owner_email)
	overall_rating := total_rating / float64(total_trip)

	var order Order
	err := orderCollection.FindOne(c, bson.M{"id": orderId}).Decode(&order)
	if err != nil {
		return responses.Order_Response{}, nil
	}

	totPrice := math.Ceil(order.To_Time.Sub(order.From_Time).Hours()/24.0) * daily_price

	result := responses.Order_Response{
		Car_Avatar:     car_avatar,
		Car_Brand_Type: car_brand_type,
		Car_No:         car_no,
		Owner_Email:    owner_email,
		Owner_Name:     owner_name,
		Total_Trip:     total_trip,
		Total_Rating:   overall_rating,
		Order_Id:       orderId,
		Order_Status:   order.Status,
		Ordered_Time:   order.Ordered_Time,
		Pay_Method:     order.Pay_Method,
		From_Time:      order.From_Time,
		To_Time:        order.To_Time,
		From_Where:     from_where,
		To_Where:       to_where,
		Price_Per_Day:  daily_price,
		Total:          totPrice,
	}
	return result, nil
}
