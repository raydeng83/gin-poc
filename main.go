package main

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/raydeng83/gin-poc/handlers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

var authHandler *handlers.AuthHandler
var usersHandler *handlers.UsersHandler

func init() {
	users := map[string]string{
		"admin": "pwd123",
		"packt": "pwd123",
		"ldeng": "pwd123",
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUrl))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	collectionUsers := client.Database(MongoDbName).Collection("users")

	count, err := collectionUsers.CountDocuments(ctx, bson.D{})

	if count == 0 {
		for username, password := range users {
			bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
			if err != nil {
				panic(err)
			}

			log.Println("Inserting user: " + username)

			collectionUsers.InsertOne(ctx, bson.M{
				"username": username,
				"password": string(bytes),
			})
		}
	} else {
		log.Println("Found users in Db. Skip inserting.")
	}

	// redis init
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping()
	fmt.Println(status)

	collectionUsers = client.Database(MongoDbName).Collection("users")
	usersHandler = handlers.NewUsersHandler(ctx, collectionUsers, redisClient)
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
}

func main() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3001"},
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	store, _ := redisStore.NewStore(10, "tcp", "localhost:6379", "", []byte(RedisKey))
	router.Use(sessions.Sessions("gin_poc", store))

	authorized := router.Group("/")

	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/signout", authHandler.SignOutHandler)
	router.GET("/users", usersHandler.ListUsersHandler)

	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/users", usersHandler.NewUserHandler)
		authorized.PUT("/users/:id", usersHandler.UpdateUserHandler)
		authorized.DELETE("/users/:id", usersHandler.DeleteUserHandler)
		authorized.GET("/users/:id", usersHandler.GetOneUserHandler)
	}

	router.Run()
}
