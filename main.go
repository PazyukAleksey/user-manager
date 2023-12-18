package main

import (
	api "awesomeProject/api"
	"crypto/subtle"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	e := echo.New()

	e.GET("/", api.Hello)
	e.GET("/users/:page", api.GetAllUsers)
	e.GET("/users/", api.GetAllUsers)
	e.POST("/register/", api.UserRegister)

	g := e.Group("/profile")
	g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if subtle.ConstantTimeCompare([]byte(username), []byte("admin")) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte("admin")) == 1 {
			return true, nil
		}
		return false, nil
	}))
	g.POST("/edit/", api.UserEdit)
	g.GET("/:nickname/", api.UserProfile)

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("port"))))
}
