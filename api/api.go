package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"awesomeProject/internal"
	"awesomeProject/users"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	message           = "message"
	nickname          = "nickname"
	firstname         = "firstname"
	lastname          = "lastname"
	email             = "email"
	password          = "password"
	information       = "information"
	databaseName      = "api_users"
	tableName         = "users"
	disconnectMessage = "disconnect"
	usersLimit        = 2
	envReadError      = "error loading .env file:"
)

func Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func GetAllUsers(c echo.Context) error {
	err := godotenv.Load(".env")
	if err != nil {
		return errors.Unwrap(err)
	}
	var mongoUri = fmt.Sprintf("mongodb+srv://%s:%s@%s", os.Getenv("db_account"), os.Getenv("db_password"), os.Getenv("db_url"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return errors.Unwrap(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.WithFields(log.Fields{
				message: disconnectMessage,
			})
		}
	}()

	collection := client.Database(databaseName).Collection(tableName)
	filter := bson.M{}
	page, _ := strconv.ParseInt(c.Param("page"), 10, 64)
	options := options.Find()
	options.SetLimit(usersLimit)
	options.SetSkip(page * usersLimit)
	cursor, err := collection.Find(context.TODO(), filter, options)
	if err != nil {
		return errors.New("users not found")
	}
	var sb strings.Builder
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		var user users.User
		err := cursor.Decode(&user)
		if err != nil {
			return errors.New("error on users returning")
		}
		sb.WriteString(fmt.Sprintf("%s - %s\n", user.Nickname, user.Email))
	}

	return c.String(http.StatusOK, sb.String())
}

func UserRegister(c echo.Context) error {
	user, err := CreateUser(c)
	if err != nil {
		return errors.Unwrap(err)
	}
	err = addUser(*user)
	if err != nil {
		return errors.Unwrap(err)
	}
	return c.String(http.StatusOK, "user added")
}

func CreateUser(c echo.Context) (*users.User, error) {
	user := &users.User{
		FirstName:   c.FormValue(firstname),
		LastName:    c.FormValue(lastname),
		Nickname:    c.FormValue(nickname),
		Email:       strings.ToLower(c.FormValue(email)),
		Password:    c.FormValue(password),
		Information: c.FormValue(information),
		CreatedAt:   time.Now(),
	}

	if !internal.NameValidation(user.FirstName) {
		return nil, errors.New("incorect first name")
	}

	if !internal.NameValidation(user.LastName) {
		return nil, errors.New("incorect last name")
	}

	if !internal.NicknameValidation(user.Nickname) {
		return nil, errors.New("incorect nickname")
	}

	if !internal.EmailValidation(user.Email) {
		return nil, errors.New("incorect email")
	}

	if !internal.PasswordValidation(user.Password) {
		return nil, errors.New("incorect password")
	}

	res, _ := getUserBy(user.Nickname)
	if res != nil {
		return nil, errors.New("user with this email or username exists")
	}

	user.Password = internal.HashString(user.Password)

	return user, nil
}

func UserEdit(c echo.Context) error {
	err := godotenv.Load(".env")
	if err != nil {
		return errors.Unwrap(err)
	}
	var mongoUri = fmt.Sprintf("mongodb+srv://%s:%s@%s", os.Getenv("db_account"), os.Getenv("db_password"), os.Getenv("db_url"))
	if c.FormValue(nickname) == "" {
		return errors.New("nickname not found")
	}
	user, err := getUserBy(c.FormValue(nickname))
	if err != nil {
		return errors.New(err.Error())
	}
	if c.FormValue(firstname) != "" {
		if !internal.NameValidation(c.FormValue(firstname)) {
			return errors.New("incorrect First name")
		}
		user.FirstName = c.FormValue(firstname)
	}
	if c.FormValue(lastname) != "" {
		if !internal.NameValidation(c.FormValue(lastname)) {
			return errors.New("incorrect Last name")
		}
		user.LastName = c.FormValue(lastname)
	}
	if c.FormValue(password) != "" {
		if !internal.PasswordValidation(c.FormValue(password)) {
			return errors.New("incorrect password")
		}
		user.Password = internal.HashString(c.FormValue(password))
	}
	if c.FormValue(information) != "" {
		user.Information = c.FormValue(information)
	}
	user.UpdatedAt = time.Now()

	filter := bson.M{nickname: user.Nickname}

	update := bson.M{
		"$set": bson.M{
			firstname:    user.FirstName,
			lastname:     user.LastName,
			information:  user.Information,
			"updated_at": user.UpdatedAt,
			password:     user.Password,
		},
	}
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return errors.Unwrap(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.WithFields(log.Fields{
				message: disconnectMessage,
			}).Warning(err)
		}
	}()
	collection := client.Database(databaseName).Collection(tableName)
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return errors.New("update error")
	}
	return c.String(http.StatusOK, "user edited")
}

func UserProfile(c echo.Context) error {

	paramNickname := c.Param(nickname)

	user, err := getUserBy(paramNickname)
	if err != nil {
		return errors.New("user not found")
	}

	return c.String(http.StatusOK, fmt.Sprintf("User name: %s\nUser lastname: %s", user.FirstName, user.LastName))
}

func addUser(user users.User) error {
	err := godotenv.Load(".env")
	if err != nil {
		return errors.Unwrap(err)
	}
	var mongoUri = fmt.Sprintf("mongodb+srv://%s:%s@%s", os.Getenv("db_account"), os.Getenv("db_password"), os.Getenv("db_url"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return errors.Unwrap(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.WithFields(log.Fields{
				message: disconnectMessage,
			}).Warning(err)
		}
	}()

	collection := client.Database(databaseName).Collection(tableName)
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		return errors.Unwrap(err)
	}
	return nil
}

func getUserBy(s string) (*users.User, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	var mongoUri = fmt.Sprintf("mongodb+srv://%s:%s@%s", os.Getenv("db_account"), os.Getenv("db_password"), os.Getenv("db_url"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, errors.Unwrap(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.WithFields(log.Fields{
				message: disconnectMessage,
			}).Warning(err)
		}
	}()

	collection := client.Database(databaseName).Collection(tableName)
	filter := bson.M{
		"$or": []bson.M{
			{nickname: s},
		},
	}
	var userResult users.User
	err = collection.FindOne(context.TODO(), filter).Decode(&userResult)

	if err != nil {
		return nil, errors.Unwrap(err)
	}
	return &userResult, nil
}
