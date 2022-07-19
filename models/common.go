package models

import (
	"server/configs"

	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")
var carCollection *mongo.Collection = configs.GetCollection(configs.DB, "cars")
var orderCollection *mongo.Collection = configs.GetCollection(configs.DB, "orders")
var reviewCollection *mongo.Collection = configs.GetCollection(configs.DB, "reviews")
var cartypeCollection *mongo.Collection = configs.GetCollection(configs.DB, "cartypes")
