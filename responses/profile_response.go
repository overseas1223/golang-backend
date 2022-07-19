package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile_Car_Response struct {
	Car_Brand   string  `json:"car_brand" binding:"required"`
	Car_Type    string  `json:"car_type" binding:"required"`
	Car_No      string  `json:"car_no" binding:"required"`
	Car_Price   float64 `json:"car_price" binding:"required"`
	Car_Rating  float64 `json:"car_rating" binding:"required"`
	Car_Gearbox string  `json:"car_gearbox" binding:"required"`
	Car_Seats   int     `json:"car_seats" binding:"required"`
	Car_Avatar  string  `json:"car_avatar" binding:"required"`
}

type Profile_User_Order_Response struct {
	Id           primitive.ObjectID `json:"id,omitempty"`
	Owner_Email  string             `json:"owner_email,omitempty"`
	Car_No       string             `json:"car_no,omitempty"`
	Status       string             `json:"status,omitempty"`
	Total        float64            `json:"total,omitempty"`
	Ordered_Time time.Time          `json:"ordered_time,omitempty"`
}

type Profile_Business_Order_Response struct {
	UserAvatar string    `json:"user_avatar"`
	Username   string    `json:"username,omitempty"`
	CarDes     string    `json:"car_des,omitempty"`
	From_Time  time.Time `json:"from_time,omitempty"`
	To_Time    time.Time `json:"to_time,omitempty"`
	Status     string    `json:"status,omitempty"`
}

type Profile_Revenue_Response struct {
	Username string  `json:"username,omitempty"`
	Status   string  `json:"status,omitempty"`
	Car_No   string  `json:"car_no,omitempty"`
	Budget   float64 `json:"budget,omitempty"`
}

type Profile_Review_Response struct {
	Username   string    `json:"username,omitempty"`
	Rating     float64   `json:"rating,omitempty"`
	When       int       `json:"when,omitempty"`
	Content    string    `json:"content,omitempty"`
	UserAvatar string    `json:"user_avatar"`
	CarAvatars []string  `json:"avatars"`
	ReviewedAt time.Time `json:"reviewed_at,omitempty"`
}
