package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wesleybruno/golang-monolito/internal/store"
)

type userKey string

const userCtx userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {

	user := getUserFromCtx(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {

	//TODO(auth): replace with auth
	myId := int64(1)

	user := getUserFromCtx(r)

	err := app.store.Follower.Follow(r.Context(), myId, user.ID)
	if err != nil {

		switch err {
		case store.ErrDuplicateKey:
			app.conflictError(w, r, err)
		default:
			app.internalServerError(w, r, err)
			return
		}

	}

	if err := app.jsonResponseNoData(w, http.StatusCreated); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {

	//TODO(auth): replace with auth
	myId := int64(1)

	user := getUserFromCtx(r)

	err := app.store.Follower.Unfollow(r.Context(), myId, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponseNoData(w, http.StatusCreated); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userId := chi.URLParam(r, "userId")
		id, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := app.store.Users.GetById(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
