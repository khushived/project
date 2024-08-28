package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"io"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	localModelURL = "http://localhost:8000/summarize"
)

var ctx = context.Background()

// Struct for handling summary request
type SummaryRequest struct {
	Text string `json:"text"`
}

// Struct for handling summary response
type SummaryResponse struct {
	Summary string `json:"summary"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to PostgreSQL database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL not set in environment")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	// Fetch tweets from the past hour
	query := `SELECT content FROM tweets WHERE tweet_time >= NOW() - INTERVAL '1 hour'`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Error fetching tweets: %v", err)
	}
	defer rows.Close()

	var tweets []string
	for rows.Next() {
		var tweet string
		if err := rows.Scan(&tweet); err != nil {
			log.Fatalf("Error scanning tweet: %v", err)
		}
		tweets = append(tweets, tweet)
	}

	if len(tweets) == 0 {
		log.Println("No tweets found in the past hour")
		return
	}

	// Summarize the tweets
	summary, err := summarizeTweetsWithLocalModel(tweets)
	if err != nil {
		log.Fatalf("Error summarizing tweets: %v", err)
	}

	// Fetch subscriber emails
	subscriberQuery := `SELECT email FROM subscribers`
	subscriberRows, err := db.Query(subscriberQuery)
	if err != nil {
		log.Fatalf("Error fetching subscribers: %v", err)
	}
	defer subscriberRows.Close()

	var emails []string
	for subscriberRows.Next() {
		var email string
		if err := subscriberRows.Scan(&email); err != nil {
			log.Fatalf("Error scanning subscriber email: %v", err)
		}
		emails = append(emails, email)
	}

	// Send the summary to all subscribers
	for _, email := range emails {
		if err := sendEmail(email, "Hourly Twitter Summary", summary); err != nil {
			log.Printf("Error sending email to %s: %v", email, err)
		} else {
			log.Printf("Successfully sent summary to %s", email)
		}
	}
}

// Function to summarize tweets using the local model server
func summarizeTweetsWithLocalModel(tweets []string) (string, error) {
    text := fmt.Sprintf("Summarize the following tweets:\n\n%s", strings.Join(tweets, "\n\n"))

    requestBody, err := json.Marshal(SummaryRequest{Text: text})
    if err != nil {
        return "", err
    }

    resp, err := http.Post(localModelURL, "application/json", bytes.NewBuffer(requestBody))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        responseBody, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("failed to get summary, status code: %d, response body: %s", resp.StatusCode, string(responseBody))
    }

    var response SummaryResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return "", err
    }

    return response.Summary, nil
}


// Function to send an email
func sendEmail(to, subject, body string) error {
	from := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
}
