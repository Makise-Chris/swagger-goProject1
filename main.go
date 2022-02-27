package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "Nam12345"
	DB_NAME     = "mydb"
)

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Birthday string `json:"birthday"`
	Gender   string `json:"gender"`
	Email    string `json:"email"`
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

func main() {
	db = setupDB()

	router := gin.Default()

	router.GET("/users", getUsers)
	router.GET("/users/:id", getUser)
	router.POST("/users", createUser)
	router.PUT("/users/:id", updateUser)
	router.DELETE("/users/:id", deleteUser)
	router.Run(":3000")
}

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

func createUser(c *gin.Context) {
	var user User

	err := c.ShouldBindJSON(&user)

	if err != nil {
		checkErr(err)
		return
	}

	var response = JsonResponse{}

	fmt.Println("Inserting user into DB")

	fmt.Println("Inserting new user with name: " + user.Name + ", birthday: " + user.Birthday + ", gender: " + user.Gender + ", email: " + user.Email)

	var lastInsertID int
	err = db.QueryRow("INSERT INTO users(name, birthday, gender, email) VALUES($1, $2, $3, $4) returning id;", user.Name, user.Birthday, user.Gender, user.Email).Scan(&lastInsertID)

	// check errors
	checkErr(err)

	response = JsonResponse{Type: "success", Message: "The user has been inserted successfully!"}

	c.JSON(200, response)
}

func updateUser(c *gin.Context) {
	userId := c.Param("id")

	var user User

	err := c.ShouldBindJSON(&user)

	if err != nil {
		checkErr(err)
		return
	}

	var response = JsonResponse{}

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

		response = JsonResponse{Type: "success", Message: "The movie has been deleted successfully!"}
	}

	c.JSON(200, response)
}
