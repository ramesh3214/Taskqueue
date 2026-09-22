package controller

import (
	"encoding/json"
	"net/http"

	"github.com/ramesh3214/taskflow/dto"
	"github.com/ramesh3214/taskflow/service"
	"github.com/ramesh3214/taskflow/util"
	"github.com/ramesh3214/taskflow/validator"
)

type Authcontroller struct {
	services *service.AuthService
}

func Newcontroller(services *service.AuthService) *Authcontroller {
	return &Authcontroller{
		services: services,
	}
}

func (a Authcontroller) Login(w http.ResponseWriter, r *http.Request) {

	var logindata dto.Login

	err := json.NewDecoder(r.Body).Decode(&logindata)

	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = validator.SigninValidator(logindata)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	result, err := a.services.Login(
		ctx,
		logindata.Email,
		logindata.Password,
	)

	if err != nil {
		http.Error(
			w,
			"invalid username or password",
			http.StatusUnauthorized,
		)
		return
	}

	token, err := util.GenerateJwtToken(ctx, result.Id)

	if err != nil {
		http.Error(
			w,
			"failed to generate token",
			http.StatusInternalServerError,
		)
		return
	}

	result.JwtToken = token

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(result)
}

func (a Authcontroller) Signup(w http.ResponseWriter, r *http.Request) {

	var signupdto dto.Signup

	err := json.NewDecoder(r.Body).Decode(&signupdto)

	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = validator.SignupValidator(signupdto)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	result, err := a.services.Signup(ctx, signupdto)

	if err != nil {
		http.Error(
			w,
			"failed to create account",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(result)
}

func (a Authcontroller) DeleteAccount(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	message, err := a.services.DeleteProfile(r.Context(), userID)

	if err != nil {
		http.Error(
			w,
			"failed to delete account",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(message)
}

func (a Authcontroller) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var dto dto.Updateprofile
	userID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&dto)

	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := a.services.UpdateProfile(r.Context(), userID, dto)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "applicatio/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(result)

}

func (a Authcontroller) GetProfile(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("userID").(uint)

	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	result, err := a.services.GetProfile(r.Context(), userID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "applicatio/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(result)

}
