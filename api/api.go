package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"awesomeProject/internal/hash"
	"awesomeProject/internal/token"
	"awesomeProject/internal/validation"
	"awesomeProject/users"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	nickname        = "nickname"
	firstname       = "firstname"
	lastname        = "lastname"
	email           = "email"
	password        = "password"
	information     = "information"
	role            = "role"
	databaseName    = "api_users"
	tableName       = "users"
	usersLimit      = 3
	cacheExpiration = time.Minute
)

func Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func GetAllUsers(c echo.Context, client *mongo.Client, redisClient *redis.Client) error {
	var usersPage int64
	if c.Param("page") == "" {
		usersPage = 0
	} else {
		page, err := strconv.ParseInt(c.Param("page"), 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("bad request: %s", err))
		}
		usersPage = page
	}
	redisKey := fmt.Sprintf("/users/%d", usersPage)
	cachedData, err := redisClient.Get(context.TODO(), redisKey).Result()
	if err == nil {
		return c.String(http.StatusOK, cachedData)
	}

	collection := client.Database(databaseName).Collection(tableName)
	filter := bson.M{}
	queryOptions := options.Find().SetSort(bson.D{{Key: "Rating", Value: -1}})
	queryOptions.SetLimit(usersLimit)
	queryOptions.SetSkip(usersPage * usersLimit)
	cursor, err := collection.Find(context.TODO(), filter, queryOptions)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("users not found: %s", err))
	}
	var sb strings.Builder
	for cursor.Next(context.TODO()) {
		var user users.User
		err := cursor.Decode(&user)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("user decode error: %s", err))
		}
		if len(sb.String()) > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s - %s", user.Nickname, user.Email))
	}
	cursor.Close(context.TODO())
	err = redisClient.Set(context.TODO(), redisKey, sb.String(), cacheExpiration).Err()
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("redis error: %s", err))
	}
	return c.String(http.StatusOK, sb.String())
}

func UserRegister(c echo.Context, client *mongo.Client) error {
	user, err := CreateUser(c, client)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("create user error: %s", err))
	}
	err = addUser(*user)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("add user error: %s", err))
	}
	return c.String(http.StatusOK, "user added")
}

func CreateUser(c echo.Context, client *mongo.Client) (*users.User, error) {
	user := &users.User{
		FirstName:   c.FormValue(firstname),
		LastName:    c.FormValue(lastname),
		Nickname:    c.FormValue(nickname),
		Email:       strings.ToLower(c.FormValue(email)),
		Password:    c.FormValue(password),
		Information: c.FormValue(information),
		Role:        c.FormValue(role),
		CreatedAt:   time.Now(),
	}

	var validationErr string

	switch {
	case !validation.NameValidation(user.FirstName):
		validationErr = "incorrect first name"
	case !validation.NameValidation(user.LastName):
		validationErr = "incorrect last name"
	case !validation.NicknameValidation(user.Nickname):
		validationErr = "incorrect nickname"
	case !validation.EmailValidation(user.Email):
		validationErr = "incorrect email"
	case !validation.PasswordValidation(user.Password):
		validationErr = "incorrect password"
	}

	if validationErr != "" {
		return nil, errors.New(validationErr)
	}

	res, _ := getUserByEmail(user.Nickname, client)
	if res != nil {
		return nil, errors.New("user with this email or username exists")
	}

	user.Password = hash.HashString(user.Password)

	return user, nil
}

func UserEdit(c echo.Context, client *mongo.Client) error {

	if c.FormValue(nickname) == "" {
		return c.String(http.StatusBadRequest, fmt.Sprintf("nickname not found"))
	}
	user, err := getUserByEmail(c.FormValue(nickname), client)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("get user by email: %s", err))
	}
	if c.FormValue(firstname) != "" {
		if !validation.NameValidation(c.FormValue(firstname)) {
			return c.String(http.StatusBadRequest, fmt.Sprintf("incorrect First name: %s", err))
		}
		user.FirstName = c.FormValue(firstname)
	}
	if c.FormValue(lastname) != "" {
		if !validation.NameValidation(c.FormValue(lastname)) {
			return c.String(http.StatusBadRequest, fmt.Sprintf("incorrect Last name: %s", err))
		}
		user.LastName = c.FormValue(lastname)
	}
	if c.FormValue(password) != "" {
		if !validation.PasswordValidation(c.FormValue(password)) {
			return c.String(http.StatusBadRequest, fmt.Sprintf("incorrect password: %s", err))
		}
		user.Password = hash.HashString(c.FormValue(password))
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
	collection := client.Database(databaseName).Collection(tableName)
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("error update user: %s", err))
	}
	return c.String(http.StatusOK, "user edited")
}

func UserProfile(c echo.Context, client *mongo.Client, redisClient *redis.Client) error {
	paramNickname := c.Param(nickname)

	redisKey := fmt.Sprintf("user-profile-%s", paramNickname)
	cachedData, err := redisClient.Get(context.TODO(), redisKey).Result()
	if err == nil {
		return c.String(http.StatusOK, cachedData)
	}

	user, err := getUserByNickname(paramNickname, client)
	if err != nil {
		return c.String(http.StatusNotFound, "user not found")
	}

	result := fmt.Sprintf("User name: %s\nUser lastname: %s\nUser id: %s", user.FirstName, user.LastName, user.ID)
	err = redisClient.Set(context.TODO(), redisKey, result, cacheExpiration).Err()
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("redis error: %s", err))
	}

	return c.String(http.StatusOK, result)
}

func GetRating(c echo.Context, client *mongo.Client, redisClient *redis.Client) error {
	paramNickname := c.Param(nickname)

	redisKey := fmt.Sprintf("user-rating-%s", paramNickname)
	cachedData, err := redisClient.Get(context.TODO(), redisKey).Result()
	if err == nil {
		return c.String(http.StatusOK, cachedData)
	}

	user, err := getUserByNickname(paramNickname, client)
	if err != nil {
		return c.String(http.StatusNotFound, fmt.Sprintf("incorect user nickname: %s", err))
	}

	result := fmt.Sprintf("user has %d", user.Rating)
	err = redisClient.Set(context.TODO(), redisKey, result, cacheExpiration).Err()
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("redis error: %s", err))
	}

	return c.String(http.StatusOK, result)
}

func ChangeRating(c echo.Context, client *mongo.Client, ratingFlag bool) error {
	paramNickname := c.Param(nickname)
	tokenString := c.Request().Header["Authorization"][0]
	var rating int
	ratingDeleteFlag := false
	ratingDeleteIndex := 0
	if ratingFlag {
		rating = 1
	} else {
		rating = -1
	}
	userNickname, err := token.GetUserNicknameFromToken(tokenString)
	if err != nil {
		fmt.Println(err)
	}
	senderUser, err := getUserByNickname(userNickname, client)
	if err != nil {
		return fmt.Errorf("can't find sender %s", err)
	}
	if !senderUser.VotedAt.IsZero() && time.Now().Sub(senderUser.VotedAt) > 1 {
		return c.String(http.StatusBadRequest, fmt.Sprintf("you can voted once per 1 hour"))
	}
	senderUser.VotedAt = time.Now()
	user, err := getUserByNickname(paramNickname, client)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("incorect user nickname: %s", err))
	}
	if user.Nickname == senderUser.Nickname {
		return c.String(http.StatusBadRequest, fmt.Sprintf("user can't vote by himselves"))
	}
	var ratingList []users.UserRatingList
	if len(user.UserRatingList) > 0 {
		err = json.Unmarshal([]byte(user.UserRatingList), &ratingList)
		if err != nil {
			return c.String(http.StatusBadRequest, fmt.Sprintf("cant find user vote list: %s", err))
		}
		for index, voteItem := range ratingList {
			if voteItem.VotedNickname == senderUser.Nickname {
				if rating == voteItem.VotedRating {
					return c.String(http.StatusBadRequest, fmt.Sprintf("User can't vote twice for one person"))
				} else {
					ratingDeleteFlag = true
					ratingDeleteIndex = index
				}
			}
		}
	}

	user.Rating += rating
	if ratingDeleteFlag {
		ratingList = append(ratingList[:ratingDeleteIndex], ratingList[ratingDeleteIndex+1:]...)
	} else {
		ratingList = append(ratingList, users.UserRatingList{VotedRating: rating, VotedDate: time.Now(), VotedNickname: senderUser.Nickname})
	}
	ratingListByte, err := json.Marshal(ratingList)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("can't code user vote list: %s", err))
	}
	user.UserRatingList = string(ratingListByte)

	err = updateRatedUser(user, client)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("can't update user: %s", err))
	}

	err = updateRatedUser(senderUser, client)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("can't update user: %s", err))
	}

	return c.String(http.StatusOK, fmt.Sprintf("user has %d", user.Rating))
}

func Login(c echo.Context, client *mongo.Client) error {
	currentNickname := c.FormValue(nickname)
	currentPassword := c.FormValue(password)
	if len(currentNickname) == 0 || len(currentPassword) == 0 {
		return c.String(http.StatusBadRequest, "email or password not exist")
	}
	currentUser, err := getUserByNickname(currentNickname, client)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("user not found: %s", err))
	}

	if currentUser.Password != hash.HashString(currentPassword) {
		return c.String(http.StatusBadRequest, "password incorect")
	}

	token, err := token.GenerateToken(currentUser)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("token: %s", err))
	}

	return c.String(http.StatusOK, token)
}

func Delete(c echo.Context, client *mongo.Client) error {
	user, err := getUserByNickname(c.Param("nickname"), client)
	if err != nil {
		return c.String(http.StatusOK, "user not found")
	}

	filter := bson.M{nickname: user.Nickname}

	update := bson.M{
		"$set": bson.M{
			"deleted_ad": time.Now(),
		},
	}
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("error conncection wit mongodb: %s", err))
	}
	collection := client.Database(databaseName).Collection(tableName)
	_, err = collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("error delete user: %s", err))
	}
	return c.String(http.StatusOK, "user deleted")
}

func addUser(user users.User) error {
	err := godotenv.Load(".env")
	if err != nil {
		return errors.Unwrap(err)
	}
	var mongoUri = fmt.Sprintf("mongodb+srv://%s:%s@%s", os.Getenv("db_account"), os.Getenv("db_password"), os.Getenv("db_url"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return fmt.Errorf("mongo connect error: %s", err)
	}
	collection := client.Database(databaseName).Collection(tableName)
	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		return fmt.Errorf("insert error: %s", err)
	}
	return nil
}

func getUserByEmail(s string, client *mongo.Client) (*users.User, error) {
	collection := client.Database(databaseName).Collection(tableName)
	filter := bson.M{
		"$or": []bson.M{
			{email: s},
		},
	}
	var userResult users.User
	err := collection.FindOne(context.TODO(), filter).Decode(&userResult)
	if err != nil {
		return nil, fmt.Errorf("find one error: %s", err)
	}
	return &userResult, nil
}

func getUserByNickname(s string, client *mongo.Client) (*users.User, error) {
	collection := client.Database(databaseName).Collection(tableName)
	filter := bson.M{
		"$or": []bson.M{
			{nickname: s},
		},
	}
	var userResult users.User
	err := collection.FindOne(context.TODO(), filter).Decode(&userResult)
	if err != nil {
		return nil, fmt.Errorf("find one error: %s", err)
	}
	return &userResult, nil
}

func updateRatedUser(user *users.User, client *mongo.Client) error {
	filter := bson.M{nickname: user.Nickname}

	update := bson.M{
		"$set": bson.M{
			"VotedAt":        user.VotedAt,
			"Rating":         user.Rating,
			"UserRatingList": user.UserRatingList,
		},
	}
	collection := client.Database(databaseName).Collection(tableName)
	_, err := collection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		return err
	}
	return nil
}
