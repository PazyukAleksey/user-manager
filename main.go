package main

import (
	"awesomeProject/api"
	"awesomeProject/internal/cash"
	"awesomeProject/internal/db"
	"awesomeProject/internal/token"
	"context"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func main() {
	globalClient, err := db.ConnectToMongo()
	if err != nil {
		log.Panic(fmt.Errorf("mongo connect error: %s", err))
	}
	defer globalClient.Disconnect(context.TODO())

	redisClient, err := cash.InitRedis()
	if err != nil {
		log.Panic(fmt.Errorf("redis connect error: %s", err))
	}
	defer redisClient.Close()

	e := echo.New()
	e.GET("/", api.Hello)
	e.GET("/users/:page", func(c echo.Context) error {
		return api.GetAllUsers(c, globalClient, redisClient)
	})
	e.GET("/users/", func(c echo.Context) error {
		return api.GetAllUsers(c, globalClient, redisClient)
	})

	e.POST("/register/", func(c echo.Context) error {
		return api.UserRegister(c, globalClient)
	})
	e.POST("/log-in/", func(c echo.Context) error {
		return api.Login(c, globalClient)
	})
	e.GET("/rating/:nickname/", func(c echo.Context) error {
		return api.GetRating(c, globalClient, redisClient)
	})

	g := e.Group("/profile")
	g.Use(token.JwtMiddleware)
	g.POST("/edit/", func(c echo.Context) error {
		return api.UserEdit(c, globalClient)
	})
	g.GET("/:nickname/", func(c echo.Context) error {
		return api.UserProfile(c, globalClient, redisClient)
	})
	g.POST("/delete/:nickname/", func(c echo.Context) error {
		return api.Delete(c, globalClient)
	})
	g.POST("/add-rating/:nickname/", func(c echo.Context) error {
		return api.ChangeRating(c, globalClient, true)
	})
	g.POST("/sub-rating/:nickname/", func(c echo.Context) error {
		return api.ChangeRating(c, globalClient, false)
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("port"))))
}
