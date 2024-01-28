package todo

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title string `json:"text"`
}

func (Todo) TableName() string {
	return "todolist"
}

type TodoHandler struct {
	db *gorm.DB
}

func NewTodoHandler(db *gorm.DB) *TodoHandler {
	return &TodoHandler{db: db}
}

func (handler *TodoHandler) NewTask(context *gin.Context) {
	var todo Todo
	if err := context.ShouldBindJSON(&todo); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if todo.Title == "sleep" {
		transactionId := context.Request.Header.Get("TransactionId")
		audience, _ := context.Get("aud")
		log.Println(transactionId, audience, "not allowed")
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "not allowed",
		})
		return
	}

	result := handler.db.Create(&todo)
	if err := result.Error; err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"ID": todo.Model.ID,
	})
}

func (handler *TodoHandler) List(context *gin.Context) {
	var result []Todo
	r := handler.db.Find(&result)
	if err := r.Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, result)
}

func (handler *TodoHandler) Remove(context *gin.Context) {
	idParam := context.Param("id")

	// parse to int
	id, err := strconv.Atoi(idParam)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	r := handler.db.Delete(&Todo{}, id)
	if err := r.Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
