package main

import (
    "database/sql"
    "log"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"

    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
    // Load environment variables from .env file
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Connect to PostgreSQL database
    var err error
    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatalf("Failed to connect to the database: %v", err)
    }
    defer db.Close()

    // Run tests
    os.Exit(m.Run())
}

func TestSubscribeHandler(t *testing.T) {
    req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(subscribeHandler)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expected := "Subscribe to Tweet Monitoring"
    if !strings.Contains(rr.Body.String(), expected) {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}

func TestProcessSubscriptionHandler(t *testing.T) {
    form := url.Values{}
    form.Add("email", "test@example.com")
    req, err := http.NewRequest("POST", "/subscribe", strings.NewReader(form.Encode()))
    if err != nil {
        t.Fatal(err)
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    rr := httptest.NewRecorder()
    handler := http.HandlerFunc(processSubscriptionHandler)
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusSeeOther {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusSeeOther)
    }

    // Check if the email was inserted into the database
    var email string
    err = db.QueryRow("SELECT email FROM subscribers WHERE email = $1", "test@example.com").Scan(&email)
    if err != nil {
        t.Errorf("Failed to find inserted email: %v", err)
    }
    if email != "test@example.com" {
        t.Errorf("Inserted email does not match: got %v want %v", email, "test@example.com")
    }
}

func TestSendEmailNotification(t *testing.T) {
    err := sendEmailNotification("test@example.com", "Test Subject", "Test Body")
    if err != nil {
        t.Errorf("Failed to send email: %v", err)
    }
}
