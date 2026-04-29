package main

import (
	"errors"
	"log"
	"net/http"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
    // Define request structure
    var requestPayload struct {
        From    string `json:"from"`
        To      string `json:"to"`
        Subject string `json:"subject"`
        Message string `json:"message"`
    }

    // Parse JSON request body
    err := app.readJSON(w, r, &requestPayload)
    if err != nil {
        log.Println(err)
        app.errorJSON(w, err, http.StatusBadRequest)
        return
    }

    // Validate required fields
    if requestPayload.From == "" {
        app.errorJSON(w, errors.New("from address is required"), http.StatusBadRequest)
        return
    }

    if requestPayload.To == "" {
        app.errorJSON(w, errors.New("recipient address is required"), http.StatusBadRequest)
        return
    }

    if requestPayload.Subject == "" {
        app.errorJSON(w, errors.New("subject is required"), http.StatusBadRequest)
        return
    }

    if requestPayload.Message == "" {
        app.errorJSON(w, errors.New("message is required"), http.StatusBadRequest)
        return
    }

    // Create message object
    msg := Message{
        From:    requestPayload.From,
        To:      requestPayload.To,
        Subject: requestPayload.Subject,
        Data:    requestPayload.Message,
    }

    // Send the email
    err = app.Mailer.SendSMTPMessage(msg)
    if err != nil {
        log.Println(err)
        app.errorJSON(w, err, http.StatusInternalServerError)
        return
    }

    // Success response
    response := jsonResponse{
        Error:   false,
        Message: "sent to " + requestPayload.To,
    }

    app.writeJSON(w, http.StatusOK, response)
}