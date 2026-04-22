package main

import (
	"errors"
	"fmt"
	"log-service/data"
	"net/http"
	"time"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if requestPayload.Name == "" {
		app.errorJSON(w, errors.New("name is required"))
		return
	}

	event := data.LogEntry{
		Name:      requestPayload.Name,
		Data:      requestPayload.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = app.Models.LogEntry.Insert(event)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("insert error: %w", err))
		return
	}

	resp := jsonResponse{
		Error:   false,
		Message: "logged",
	}

	app.writeJSON(w, http.StatusOK, resp)
}