package controller

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/switcherapi/switcher-gitops/src/config"
	"github.com/switcherapi/switcher-gitops/src/utils"
)

func DefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		next.ServeHTTP(w, r)
	})
}

func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validateToken(r) {
			next.ServeHTTP(w, r)
		} else {
			utils.ResponseJSON(w, ErrorResponse{
				Error: "Invalid token",
			}, http.StatusUnauthorized)
		}
	})
}

func validateToken(r *http.Request) bool {
	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		return false
	}

	parts := strings.Split(authToken, " ")
	if len(parts) != 2 {
		return false
	}

	tokenStr := parts[1]
	aoiKey := config.GetEnv("SWITCHER_API_JWT_SECRET")

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(aoiKey), nil
	})

	if err != nil {
		return false
	}

	return token.Valid
}
