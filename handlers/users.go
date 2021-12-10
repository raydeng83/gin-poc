package handlers

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/raydeng83/gin-poc/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type UsersHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewUsersHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *UsersHandler {
	return &UsersHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

func (handler *UsersHandler) ListUsersHandler(c *gin.Context) {
	val, err := handler.redisClient.Get("users").Result()
	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		opts := options.Find().SetProjection(bson.M{"password": 0})
		cur, err := handler.collection.Find(handler.ctx, bson.M{}, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(handler.ctx)

		users := make([]models.User, 0)
		for cur.Next(handler.ctx) {
			var user models.User
			cur.Decode(&user)
			users = append(users, user)
		}

		data, _ := json.Marshal(users)
		handler.redisClient.Set("users", string(data), 0)
		c.JSON(http.StatusOK, users)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to Redis")
		users := make([]models.User, 0)
		json.Unmarshal([]byte(val), &users)
		c.JSON(http.StatusOK, users)
	}
}

func (handler *UsersHandler) NewUserHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new user"})
		return
	}

	log.Println("Remove data from Redis")
	handler.redisClient.Del("users")

	c.JSON(http.StatusOK, user)
}

func (handler *UsersHandler) UpdateUserHandler(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", user.Username},
	}}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User has been updated"})
}

func (handler *UsersHandler) DeleteUserHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User has been deleted"})
}

func (handler *UsersHandler) GetOneUserHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	}, options.FindOne().SetProjection(bson.M{"password": 0}))
	var user models.User
	err := cur.Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
