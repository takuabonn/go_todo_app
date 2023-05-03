package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Login struct {
	Service   LoginService
	Validator *validator.Validate
}

func (l *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var body struct {
		UserName string `json:"user_name" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	err := l.Validator.Struct(body)
	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	jwt, err := l.Service.Login(ctx, body.UserName, body.Password)
	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	rsp := struct {
		AcccessToken string `json:"access_token"`
	}{
		AcccessToken: jwt,
	}

	RespondJSON(ctx, w, rsp, http.StatusOK)
}
