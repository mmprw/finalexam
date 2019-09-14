package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
	"log"
	"fmt"
	"database/sql"
	"strconv"
	_"github.com/lib/pq"

)
var db 	*sql.DB
var err error
type Customer struct {
	ID  int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Status string `json:"status"`
}

func authMiddleware(c *gin.Context)  {
	fmt.Println("this is a middle ware")
	token := c.GetHeader("Authorization")
	if token != "token2019"{
		c.JSON(http.StatusUnauthorized,gin.H{"error":"unautorized."})
		c.Abort()
		return
	} 
	c.Next()
	fmt.Println("after in middleware")
}

func createCustomer(c *gin.Context) {
	var t Customer
	var id int
	err := c.ShouldBindJSON(&t) 

	if  err != nil {
		c.JSON(http.StatusBadRequest,err.Error())
		return
	}
	row := db.QueryRow("INSERT INTO Customer (name, email,status) values ($1, $2,$3) RETURNING id ", t.Name,t.Email,t.Status)
	err = row.Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't scan id"}) 
		return
	}
	t.ID = id
	c.JSON(http.StatusCreated,t) 
	//c.JSON(http.StatusOK,ct)
}
func onerowCustomer(c *gin.Context){
	id := c.Param("id") 
	rowId,err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't convert string to int"})
		return
	}

	stmt,err := db.Prepare("SELECT id,name,email,status FROM Customer where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't prepare query one row statment"})
		return
	}
	
	row := stmt.QueryRow(rowId)
	var cust Customer

	err = row.Scan(&cust.ID,&cust.Name,&cust.Email,&cust.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't Scan row into varibles"})
		return
	}
	c.JSON(http.StatusOK,cust)

}
func allCustomer(c *gin.Context) {
	stmt,err := db.Prepare("SELECT id,name,email,status FROM Customer")
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status": "can't prepare query all todos statment"})
		return
	}
	row,err := stmt.Query()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status": "Error!!! Query"})
		return
	}
	var cust []Customer
	for row.Next(){
		var id int
		var name ,email,status string
		err := row.Scan(&id,&name,&email,&status)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"status":"can't scan row into variable"})
			return
		}
		cust = append(cust,Customer{id,name,email,status})
	}	
	c.JSON(http.StatusOK,cust)	
}

func updateCustomer(c *gin.Context) {
	id := c.Param("id") 
	rowId,err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't convert string to int"})
		return
	}

	var ctm Customer
	err = c.ShouldBindJSON(&ctm)

	stmt ,err := db.Prepare("UPDATE Customer SET name=$2,email=$3 ,status=$4 where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't prepare statment update"})
		return

	}
	if _,err := stmt.Exec(rowId,ctm.Name,ctm.Email,ctm.Status); err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error execute update"})
		return
	}
	c.JSON(http.StatusOK,ctm)


}
func daleteCustomer(c *gin.Context) {
	id := c.Param("id") 
	rowId,err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't convert string to int"})
		return
	}
	var ctm Customer
	err = c.ShouldBindJSON(&ctm)
	stmt,err := db.Prepare("DELETE FROM Customer where id=$1")
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"can't prepare query one row statment"})
		c.Abort()
		return
	}
	if _,err := stmt.Exec(rowId); err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"status":"error execute delete"})
		return
	}
	c.JSON(http.StatusOK,map[string]string{"message":"customer deleted"})
}

func main() {

	db,err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error",err)
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF  NOT EXISTS Customer (    
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
	`  
	_,err= db.Exec(createTb) 
	if err != nil {
		log.Fatal("Connect to database error",err)
	}

	r := gin.Default()
	r.Use(authMiddleware)
	r.POST("/customers",createCustomer)
	r.GET("/customers/:id",onerowCustomer)
	r.GET("/customers",allCustomer)
	r.PUT("/customers/:id",updateCustomer)
	r.DELETE("/customers/:id",daleteCustomer)
	r.Run(":2019")
}