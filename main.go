package main

import (
	"database/sql"
	"fmt"
	"regexp"

	docs "gin-project1/docs"

	"github.com/go-playground/validator/v10"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "Nam12345"
	DB_NAME     = "mydb"
)

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name" validate:"required"`
	Birthday string `json:"birthday" validate:"required,checkdate"`
	Gender   string `json:"gender" validate:"required,oneof=nam nữ"`
	Email    string `json:"email" validate:"required,email"`
}

type JsonResponse struct {
	Type    string `json:"type"`
	Data    []User `json:"data"`
	Message string `json:"message"`
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func setupDB() *sql.DB {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return db
}

var db *sql.DB

var validate *validator.Validate

func main() {
	db = setupDB()

	validate = validator.New()

	validate.RegisterValidation("checkdate", checkDate)

	router := gin.Default()

	docs.SwaggerInfo.BasePath = "/"

	router.GET("/users", getUsers)
	router.GET("/users/:id", getUser)
	router.POST("/users", createUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.Run(":3000")
}

func checkDate(f1 validator.FieldLevel) bool {
	str := f1.Field().String()

	re := regexp.MustCompile("((19|20)\\d\\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])")

	fmt.Println(re.MatchString(str))

	return re.MatchString(str)
}

// @BasePath /

//GetUsers godoc
//@Summary Lay danh sach User
//@Description Lay danh sach User
//@Tags Users
//@Accept json
//@Produce json
//@Success 200 {object} JsonResponse
//@Router /users [get]
func getUsers(c *gin.Context) {
	fmt.Println("Getting users...")

	rows, err := db.Query("SELECT * FROM users")

	// check errors
	checkErr(err)

	// var response []JsonResponse
	var users []User

	for rows.Next() {
		var id int
		var name string
		var birthday string
		var gender string
		var email string

		err = rows.Scan(&id, &name, &birthday, &gender, &email)

		// check errors
		checkErr(err)

		users = append(users, User{Id: id, Name: name, Birthday: birthday, Gender: gender, Email: email})
	}

	var response = JsonResponse{Type: "success", Data: users}

	c.JSON(200, response)
}

//GetUser godoc
//@Summary Lay User tuong ung
//@Description Lay User tuong ung
//@Tags Users
//@Accept json
//@Produce json
//@Param  id path int true "User ID"
//@Success 200 {object} JsonResponse
//@Router /users/{id} [get]
func getUser(c *gin.Context) {
	userId := c.Param("id")

	var response = JsonResponse{}

	if userId == "" {
		response = JsonResponse{Type: "error", Message: "You are missing userId parameter."}
	} else {
		fmt.Println("getting user " + userId + " from DB...")

		rows, err := db.Query("SELECT * FROM users where id = " + userId)

		// check errors
		checkErr(err)

		var users []User

		for rows.Next() {
			var id int
			var name string
			var birthday string
			var gender string
			var email string

			err = rows.Scan(&id, &name, &birthday, &gender, &email)

			// check errors
			checkErr(err)

			users = append(users, User{Id: id, Name: name, Birthday: birthday, Gender: gender, Email: email})
		}

		response = JsonResponse{Type: "success", Data: users, Message: "Get user " + userId + " successfully!"}
	}

	c.JSON(200, response)
}

//CreateUser godoc
//@Summary  Tao User moi
//@Description Tao User moi
//@Tags Users
//@Accept json
//@Produce json
//@Param  user body User true "Create User"
//@Success 200 {object} JsonResponse
//@Failure 400 {object} JsonResponse
//@Router /users [post]
func createUser(c *gin.Context) {
	var user User

	var response = JsonResponse{}

	//validate
	c.ShouldBindJSON(&user)
	err := validate.Struct(user)
	if err != nil {
		fmt.Println("-----")
		fmt.Println(err)
		fmt.Println("-----")

		var message string

		for _, err := range err.(validator.ValidationErrors) {
			/*
				fmt.Println(err.StructField())
				fmt.Println(err.ActualTag())
				fmt.Println(err.Kind())
				fmt.Println(err.Value())
				fmt.Println(err.Param())
				fmt.Println("---------------")
			*/
			if err.ActualTag() == "required" {
				message = message + "Nhập thiếu thông tin. "
			}
			if err.ActualTag() == "checkdate" {
				message = message + "Nhập sai định dạng ngày tháng (yyyy-mm-dd). "
			}
			if err.ActualTag() == "oneof" {
				message = message + "Nhập sai giới tính (nam hoặc nữ). "
			}
			if err.ActualTag() == "email" {
				message = message + "Nhập sai định dạng email. "
			}
		}

		response = JsonResponse{Type: "fail", Message: message}

		c.JSON(400, response)

		c.Abort()

		return
	}

	fmt.Println("Inserting user into DB")

	fmt.Println("Inserting new user with name: " + user.Name + ", birthday: " + user.Birthday + ", gender: " + user.Gender + ", email: " + user.Email)

	var lastInsertID int
	err = db.QueryRow("INSERT INTO users(name, birthday, gender, email) VALUES($1, $2, $3, $4) returning id;", user.Name, user.Birthday, user.Gender, user.Email).Scan(&lastInsertID)

	// check errors
	fmt.Println(err)

	response = JsonResponse{Type: "success", Message: "The user has been inserted successfully!"}

	c.JSON(200, response)
}

//UpdateUser godoc
//@Summary  Sua thong tin User
//@Description Sua thong tin User
//@Tags Users
//@Accept json
//@Produce json
//@Param  id path int true "User ID"
//@Param  user body User true "Update User"
//@Success 200 {object} JsonResponse
//@Failure 400 {object} JsonResponse
//@Router /users/{id} [put]
func updateUser(c *gin.Context) {
	userId := c.Param("id")

	var user User

	var response = JsonResponse{}

	//validate
	c.ShouldBindJSON(&user)
	err := validate.Struct(user)
	if err != nil {
		fmt.Println("-----")
		fmt.Println(err)
		fmt.Println("-----")

		var message string

		for _, err := range err.(validator.ValidationErrors) {
			/*
				fmt.Println(err.StructField())
				fmt.Println(err.ActualTag())
				fmt.Println(err.Kind())
				fmt.Println(err.Value())
				fmt.Println(err.Param())
				fmt.Println("---------------")
			*/
			if err.ActualTag() == "required" {
				message = message + "Nhập thiếu thông tin. "
			}
			if err.ActualTag() == "checkdate" {
				message = message + "Nhập sai định dạng ngày tháng (yyyy-mm-dd). "
			}
			if err.ActualTag() == "oneof" {
				message = message + "Nhập sai giới tính (nam hoặc nữ). "
			}
			if err.ActualTag() == "email" {
				message = message + "Nhập sai định dạng email. "
			}
		}

		response = JsonResponse{Type: "fail", Message: message}

		c.JSON(400, response)

		c.Abort()

		return
	}

	fmt.Println("updating user " + userId + " from DB...")

	// create the update sql query
	sqlStatement := `UPDATE users SET name=$2, birthday=$3, gender=$4, email=$5 WHERE id=$1`

	// execute the sql statement
	_, err = db.Exec(sqlStatement, userId, user.Name, user.Birthday, user.Gender, user.Email)

	// check errors
	checkErr(err)

	response = JsonResponse{Type: "success", Message: "Update user " + userId + " successfully!"}

	c.JSON(200, response)
}

//DeleteUser godoc
//@Summary  Xoa User
//@Description Xoa User
//@Tags Users
//@Accept json
//@Produce json
//@Param  id path int true "User ID"
//@Success 200 {object} JsonResponse
//@Router /users/{id} [delete]
func deleteUser(c *gin.Context) {
	userId := c.Param("id")

	var response = JsonResponse{}

	if userId == "" {
		response = JsonResponse{Type: "error", Message: "You are missing userId parameter."}
	} else {
		fmt.Println("Deleting user from DB")

		_, err := db.Exec("DELETE FROM users where id = $1", userId)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "User " + userId + " has been deleted successfully!"}
	}

	c.JSON(200, response)
}
