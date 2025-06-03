package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Shakshor/blog_project_go/internal/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create user:%v", err))
		return
	}

	// respondWithJSON(w, 200, user)
	respondWithJSON(w, 201, databaseUserToUser(user))
}

func (apiCfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, 200, user)
}

func (apiCfg *apiConfig) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	// secret key
	var jwt_key = []byte(os.Getenv(("JWT_SECRET")))

	// get the user data
	// decoding the request body
	type parameters struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing json: %v", err))
	}

	// get the db user data
	dbUser, err := apiCfg.DB.GetUserByName(r.Context(), params.Name)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get user: %v", err))
		return
	}

	// compare btn requested data and db data
	if params.Name == dbUser.Name {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_name": params.Name,
			"exp":       time.Now().Add(time.Hour * 24 * 30).Unix(),
		})

		tokenString, err := token.SignedString(jwt_key)
		if err != nil {
			respondWithError(w, 400, fmt.Sprintf("Failed to create token: %v", err))
			return
		}

		fmt.Println("generated token", tokenString, err)

		respondWithJSON(w, 200, tokenString)

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "token",
				Value:    tokenString,
				Path:     "",
				Domain:   "",
				Expires:  time.Now().Add(time.Hour * 24 * 30),
				Secure:   false,
				HttpOnly: true,
			})
	} else {
		respondWithError(w, 400, fmt.Sprintf("Invalid credentials: %v", err))
		return
	}
}

func (apiCfg *apiConfig) handlerGetPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	posts, err := apiCfg.DB.GetPostsForUser(r.Context(), database.
		GetPostsForUserParams{
		UserID: user.ID,
		Limit:  10,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get posts %v", err))
	}

	respondWithJSON(w, 200, databasePostsToPosts(posts))
}
