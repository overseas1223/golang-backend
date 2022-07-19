package models

import (
	"context"
	"errors"
	"server/responses"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CarInput struct {
	Car_Brand   string  `json:"car_brand" binding:"required"`
	Car_Type    string  `json:"car_type" binding:"required"`
	Car_Seats   int     `json:"car_seats" binding:"required"`
	Car_Gearbox string  `json:"car_gearbox" binding:"required"`
	Car_Fuel    string  `json:"car_fuel" binding:"required"`
	Car_Price   float64 `json:"car_price"`
	Car_No      string  `json:"car_no"`
	Car_Avatar  string  `json:"car_avatar"`
	Car_From    string  `json:"car_from"`
	Car_To      string  `json:"car_to"`
}

type CarType struct {
	Id          primitive.ObjectID `json:"id"`
	Car_Brand   string             `json:"car_brand" binding:"required"`   // car brand
	Car_Type    string             `json:"car_type" binding:"required"`    // car type
	Car_Seats   int                `json:"car_seats" binding:"required"`   // seat count
	Car_Gearbox string             `json:"car_gearbox" binding:"required"` // gearbox type
	Car_Fuel    string             `json:"car_fuel" binding:"required"`    // capacity
}

type Car struct {
	Id          primitive.ObjectID `json:"id" binding:"required"`
	Car_Type_Id primitive.ObjectID `json:"car_type_id" binding:"required"` // CarType Id
	Car_Price   float64            `json:"car_price" binding:"required"`   // daily
	Car_No      string             `json:"car_no" binding:"required"`      // car id number
	Owner_Email string             `json:"owner_email"`                    // owner email
	Car_Avatar  string             `json:"car_avatar" binding:"required"`  // car avatar
	From_Where  string             `json:"from_where" binding:"required"`  // pick up place
	To_Where    string             `json:"to_where" binding:"required"`    // drop off place
}

func IsCarNoExist(c context.Context, no string) bool {
	car := Car{}
	err := carCollection.FindOne(c, bson.M{"car_no": no}).Decode(&car)
	return err == nil
}

func (newCar *Car) SaveCar(c context.Context, newType CarType) (*Car, error) {

	if IsCarNoExist(c, newCar.Car_No) {
		return nil, errors.New(responses.CAR_NO_EXIST_ERROR)
	}

	_, sErr := newType.SaveCarType(c)
	if sErr != nil && sErr.Error() != "exist" {
		return nil, sErr
	}

	id, mErr := FindCarTypeId(c, newType)
	if mErr != nil {
		return nil, mErr
	}

	newCar.Car_Type_Id = id

	_, err := carCollection.InsertOne(c, newCar)
	if err != nil {
		if er, ok := err.(mongo.WriteException); ok && er.WriteErrors[0].Code == 11000 {
			return nil, errors.New(responses.CAR_NO_EXIST_ERROR)
		}
	}

	opt := options.Index()
	opt.SetUnique(true)
	index := mongo.IndexModel{Keys: bson.M{"car_no": 1}, Options: opt}

	if _, err := carCollection.Indexes().CreateOne(c, index); err != nil {
		return nil, errors.New("could not create index of car number")
	}

	return nil, nil
}

func (newCarType *CarType) SaveCarType(c context.Context) (*CarType, error) {
	// validate if cartype is existing.
	var ntp CarType
	res := cartypeCollection.FindOne(c, bson.M{
		"car_brand":   newCarType.Car_Brand,
		"car_type":    newCarType.Car_Type,
		"car_seats":   newCarType.Car_Seats,
		"car_gearbox": newCarType.Car_Gearbox,
		"car_fuel":    newCarType.Car_Fuel,
	}).Decode(&ntp)

	if res != nil {
		_, err := cartypeCollection.InsertOne(c, newCarType)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	return nil, errors.New("exist")
}

func FindCarTypeId(c context.Context, newType CarType) (primitive.ObjectID, error) {
	var resType CarType
	res := cartypeCollection.FindOne(c, bson.M{
		"car_brand":   newType.Car_Brand,
		"car_type":    newType.Car_Type,
		"car_seats":   newType.Car_Seats,
		"car_gearbox": newType.Car_Gearbox,
		"car_fuel":    newType.Car_Fuel,
	}).Decode(&resType)

	if res != nil {
		return primitive.NewObjectID(), res
	}
	return resType.Id, nil
}

func GetCarTypeById(c context.Context, id primitive.ObjectID) (*CarType, error) {
	var resType CarType
	res := cartypeCollection.FindOne(c, bson.M{"id": id}).Decode(&resType)

	if res != nil {
		return nil, res
	}

	return &resType, nil
}

func GetCarPriceByNo(c context.Context, car_no string) (float64, error) {
	var cur Car
	err := carCollection.FindOne(c, bson.M{"car_no": car_no}).Decode(&cur)
	if err != nil {
		return 0.0, err
	}
	return cur.Car_Price, nil
}

func GetCarListsByEmail(c context.Context, email string) ([]responses.Profile_Car_Response, error) {
	cur, err := carCollection.Find(c, bson.M{"owner_email": email})
	if err != nil {
		return nil, errors.New(responses.CAR_EMPTY)
	}

	defer cur.Close(c)

	var result []responses.Profile_Car_Response

	for cur.Next(c) {
		var res Car
		err := cur.Decode(&res)
		if err != nil {
			continue
		}

		// need to aggregate
		curType, mErr := GetCarTypeById(c, res.Car_Type_Id)
		if mErr != nil {
			continue
		}

		cnt, rating, err := GetCarRating(c, res.Car_No)
		if err != nil {
			rating = 0.0
		} else {
			rating = rating / float64(cnt)
		}

		curCar := responses.Profile_Car_Response{
			Car_Brand:   curType.Car_Brand,
			Car_Type:    curType.Car_Type,
			Car_Seats:   curType.Car_Seats,
			Car_Gearbox: curType.Car_Gearbox,
			Car_No:      res.Car_No,
			Car_Price:   res.Car_Price,
			Car_Rating:  rating,
			Car_Avatar:  res.Car_Avatar,
		}
		result = append(result, curCar)
	}

	return result, nil
}

func GetCarRating(c context.Context, car_no string) (int, float64, error) {
	cur, err := orderCollection.Find(c, bson.M{"car_no": car_no})
	if err != nil {
		return 0, 0, err
	}

	defer cur.Close(c)
	sum := 0.0
	cnt := 0

	for cur.Next(c) {
		var res1 Order
		err = cur.Decode(&res1)
		if err != nil {
			continue
		}
		var res2 Review
		err = reviewCollection.FindOne(c, bson.M{"order_id": res1.Id}).Decode(&res2)
		if err != nil {
			continue
		}
		sum = sum + res2.Rating
		cnt = cnt + 1
	}

	if cnt < 1 {
		return 0, 0, nil
	}

	return cnt, sum, nil
}

func GetOwnerEmailByCarNo(c context.Context, car_no string) (string, error) {
	var car Car
	err := carCollection.FindOne(c, bson.M{"car_no": car_no}).Decode(&car)
	if err != nil {
		// * Meaning no match
		return "", err
	}

	return car.Owner_Email, nil
}

func GetCarDetailsByNo(c context.Context, car_no string) (string, string, string, string, float64) {
	var car Car
	err := carCollection.FindOne(c, bson.M{"car_no": car_no}).Decode(&car)
	if err != nil {
		return "", "", "", "", 0
	}
	var car_type CarType
	err = cartypeCollection.FindOne(c, bson.M{"id": car.Car_Type_Id}).Decode(&car_type)
	if err != nil {
		return "", "", "", "", 0
	}

	return car.Car_Avatar, car_type.Car_Brand + " " + car_type.Car_Type, car.From_Where, car.To_Where, car.Car_Price
}
