package main

import (
	"backend/internal/models"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Welcome to the forum",
		Version: "1.0.0",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := app.DB.GetPosts()

	if err != nil {
		app.errorJSON(w, err)
	}

	_ = app.writeJSON(w, http.StatusOK, posts)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {

	//read json payload
	var requestPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	//validate user against database

	user, err := app.DB.GetUserByUsername(requestPayload.Username)
	if err != nil {
		app.errorJSON(w, errors.New("invalid username"), http.StatusBadRequest)
		return
	}

	// check password
	isValid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !isValid {
		app.errorJSON(w, errors.New("invalid password"), http.StatusBadRequest)
		return
	}

	// create a jwt user
	u := jwtUser{
		ID:       user.ID,
		Username: user.Username,
	}

	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// to show that refresh token is generated
	// log.Println("refresh_tokens:", tokens.RefreshToken)

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	// log.Println("cookies generated:", refreshCookie)	// to show that refresh token is generated
	http.SetCookie(w, refreshCookie)
	// log.Println("cookies set:", refreshCookie) // to show that refresh token is set
	app.writeJSON(w, http.StatusAccepted, tokens)

}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			//parse token to get claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}
			//get User ID from token claims
			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}
			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:       user.ID,
				Username: user.Username,
			}

			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}
			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))
			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

// func (app *application) PostCatalog(w http.ResponseWriter, r *http.Request) {
// 	posts, err := app.DB.GetPosts()

// 	if err != nil {
// 		app.errorJSON(w, err)
// 	}

// 	_ = app.writeJSON(w, http.StatusOK, posts)
// }




// func (app *application) AllPostsByGenre(w http.ResponseWriter, r *http.Request) {
// 	id, err := strconv.Atoi(chi.URLParam(r, "id"))
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}
// 	posts, err := app.DB.GetPosts(id)
// 	if err != nil {
// 		app.errorJSON(w, err)
// 		return
// 	}
// 	app.writeJSON(w, http.StatusOK, posts)
// }


func (app *application) AllPostsByUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	posts, err := app.DB.GetPostsFromUser(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, posts)
}


func (app *application) GetPost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	post, err := app.DB.OnePost(postID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, post)

}

func (app *application) PostForEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	postID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	post, genres, err := app.DB.OnePostForEdit(postID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload = struct {
		Post   *models.Post    `json:"post"`
		Genres []*models.Genre `json:"genres"`
	}{
		post,
		genres,
	}
	_ = app.writeJSON(w, http.StatusOK, payload)

}
