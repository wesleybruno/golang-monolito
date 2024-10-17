package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/wesleybruno/golang-monolito/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/auth/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		// Role: store.Role{
		// 	Name: "user",
		// },
	}

	// // hash the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	// hash the token for storage but keep the plain token for email
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	// activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	// isProdEnv := app.config.env == "production"
	// vars := struct {
	// 	Username      string
	// 	ActivationURL string
	// }{
	// 	Username:      user.Username,
	// 	ActivationURL: activationURL,
	// }

	// // send mail
	// status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	// if err != nil {
	// 	app.logger.Errorw("error sending welcome email", "error", err)

	// 	// rollback user creation if email fails (SAGA pattern)
	// 	if err := app.store.Users.Delete(ctx, user.ID); err != nil {
	// 		app.logger.Errorw("error deleting user", "error", err)
	// 	}

	// 	app.internalServerError(w, r, err)
	// 	return
	// }

	// app.logger.Infow("Email sent", "status code", status)

	if err := app.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err)
	}
}

// activateUserHandler godoc
//
//	@Summary		Activate a user
//	@Description	Activate a user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Invitation Token"
//	@Success		204		{object}	string	"User activated"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/user/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerError(w, r, err)
	}

}