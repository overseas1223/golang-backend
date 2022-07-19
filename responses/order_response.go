package responses

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order_Response struct {
	Car_Avatar     string             `json:"car_avatar,omitempty"`
	Car_Brand_Type string             `json:"car_brand_type,omitempty"`
	Car_No         string             `json:"car_no,omitempty"`
	Owner_Email    string             `json:"owner_email,omitempty"`
	Owner_Name     string             `json:"owner_name,omitempty"`
	Total_Trip     int                `json:"total_trip,omitempty"`
	Total_Rating   float64            `json:"total_rating,omitempty"`
	Order_Id       primitive.ObjectID `json:"order_id,omitempty"`
	Order_Status   string             `json:"status,omitempty"`
	Ordered_Time   time.Time          `json:"ordered_time,omitempty"`
	Pay_Method     string             `json:"pay_method,omitempty"`
	From_Time      time.Time          `json:"from_time,omitempty"`
	To_Time        time.Time          `json:"to_time,omitempty"`
	From_Where     string             `json:"from_where,omitempty"`
	To_Where       string             `json:"to_where,omitempty"`
	Price_Per_Day  float64            `json:"price_per_day,omitempty"`
	Total          float64            `json:"total,omitempty"`
}
