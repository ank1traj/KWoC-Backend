package controllers

import (
	"encoding/json"
	"fmt"
	"kwoc-backend/middleware"
	"kwoc-backend/utils"
	"net/http"

	"kwoc-backend/models"

	"gorm.io/gorm"
)

type RegisterStudentReqFields struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	College  string `json:"college"`
}

func RegisterStudent(w http.ResponseWriter, r *http.Request) {
	app := r.Context().Value(middleware.APP_CTX_KEY).(*middleware.App)
	db := app.Db
	var reqFields = RegisterStudentReqFields{}

	err := json.NewDecoder(r.Body).Decode(&reqFields)
	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error decoding JSON body.", http.StatusBadRequest)
		return
	}

	// Check if the JWT login username is the same as the student's given username
	login_username := r.Context().Value(middleware.LOGIN_CTX_USERNAME_KEY).(string)

	if reqFields.Username != login_username {
		utils.LogWarn(
			r,
			fmt.Sprintf(
				"POSSIBLE SESSION HIJACKING\nJWT Username: %s, Given Username: %s",
				login_username,
				reqFields.Username,
			),
		)

		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Login username and given username do not match.")
		return
	}

	// Check if the student already exists in the db
	student := models.Student{}
	tx := db.
		Table("students").
		Where("username = ?", reqFields.Username).
		First(&student)

	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		utils.LogErrAndRespond(r, w, err, "Database error.", http.StatusInternalServerError)
		return
	}

	student_exists := student.Username == reqFields.Username

	if student_exists {
		utils.LogWarnAndRespond(
			r,
			w,
			fmt.Sprintf("Student `%s` already exists.", student.Username),
			http.StatusBadRequest,
		)
		return
	}

	// Create a db entry if the student doesn't exist
	tx = db.Create(&models.Student{
		Username: reqFields.Username,
		Name:     reqFields.Name,
		Email:    reqFields.Email,
		College:  reqFields.College,
	})

	if tx.Error != nil {
		utils.LogErrAndRespond(r, w, err, "Database error.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Student registration successful.")
}