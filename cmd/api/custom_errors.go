package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {

	log.Printf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err)

	writeJsonError(w, http.StatusInternalServerError, "The server encounter a problem")

}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {

	log.Printf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err)

	writeJsonError(w, http.StatusBadRequest, err.Error())

}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {

	log.Printf("conflict error: %s path: %s error: %s", r.Method, r.URL.Path, err)

	writeJsonError(w, http.StatusConflict, err.Error())

}

func (app *application) noContent(w http.ResponseWriter, r *http.Request) {

	log.Printf("bad request error: %s path: %s ", r.Method, r.URL.Path)

	writeJson(w, http.StatusNoContent, nil)

}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found error method: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	writeJsonError(w, http.StatusNotFound, "not found")
}
