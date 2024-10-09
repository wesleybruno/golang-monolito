package main

import (
	"net/http"
)

func (app *application) healthCheckerHandler(w http.ResponseWriter, r *http.Request) {

	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": version,
	}

	if err := writeJson(w, http.StatusOK, data); err != nil {
		writeJsonError(w, http.StatusBadGateway, err.Error())
	}

}
