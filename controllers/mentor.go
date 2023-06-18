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

type RegisterMentorReqFields struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func (dbHandler *DBHandler) RegisterMentor(w http.ResponseWriter, r *http.Request) {
	db := dbHandler.db
	var reqFields = RegisterMentorReqFields{}

	err := json.NewDecoder(r.Body).Decode(&reqFields)
	if err != nil {
		utils.LogErrAndRespond(r, w, err, "Error decoding JSON body.", http.StatusBadRequest)
		utils.LogErr(r, err, "Error decoding JSON body.")

		return
	}

	// Check if the JWT login username is the same as the mentor's given username
	login_username := r.Context().Value(middleware.LoginCtxKey(middleware.LOGIN_CTX_USERNAME_KEY)).(string)

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

	// Check if the mentor already exists in the db
	mentor := models.Mentor{}
	tx := db.
		Table("mentors").
		Where("username = ?", reqFields.Username).
		First(&mentor)

	if tx.Error != nil && tx.Error != gorm.ErrRecordNotFound {
		utils.LogErrAndRespond(r, w, tx.Error, "Database error.", http.StatusInternalServerError)
		return
	}

	mentor_exists := mentor.Username == reqFields.Username

	if mentor_exists {
		utils.LogErrAndRespond(r, w, gorm.ErrRecordNotFound, "Error: Mentor already exists.", http.StatusBadRequest)

		return
	}

	// Create a db entry if the mentor doesn't exist
	tx = db.Create(&models.Mentor{
		Username: reqFields.Username,
		Name:     reqFields.Name,
		Email:    reqFields.Email,
	})

	if tx.Error != nil {
		utils.LogErrAndRespond(r, w, tx.Error, "Database error.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Mentor registration successful.")
}
