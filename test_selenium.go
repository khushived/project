package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"crypto/sha256"
	"net/url"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
	_ "github.com/lib/pq"
)

const (
	chromeDriverPath = "C:/Users/HP/Downloads/chromedriver-win64/chromedriver-win64/chromedriver.exe"
	port             = 9515
	twitterSearchURL = "https://twitter.com/search?q="
)

var ctx = context.Background()

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdb.Close()

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


	// Prompt the user to enter the search term
	fmt.Print("Enter the hashtag or keyword to search for (without #): ")
	var searchTerm string
	fmt.Scanln(&searchTerm)

	// Construct the search URL
	searchURL := twitterSearchURL + url.QueryEscape("#"+searchTerm) + "&f=live"
	log.Printf("Navigating to search URL: %s", searchURL)

	// Start the Chromedriver service
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port)
	if err != nil {
		log.Fatalf("Error starting the Chromedriver service: %v", err)
	}
	defer func() {
		if err := service.Stop(); err != nil {
			log.Printf("Error stopping Chromedriver service: %v", err)
		}
	}()

	// Connect to the WebDriver instance
	var wd selenium.WebDriver
	retryAttempts := 5
	for i := 0; i < retryAttempts; i++ {
		wd, err = selenium.NewRemote(selenium.Capabilities{"browserName": "chrome"}, fmt.Sprintf("http://localhost:%d/wd/hub", port))
		if err == nil {
			break
		}
		log.Printf("Error connecting to the WebDriver instance (attempt %d/%d): %v", i+1, retryAttempts, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to the WebDriver instance after %d attempts: %v", retryAttempts, err)
	}
	defer func() {
		if err := wd.Quit(); err != nil {
			log.Printf("Error quitting WebDriver: %v", err)
		}
	}()

	// Open the Twitter login page
	log.Println("Navigating to Twitter login page...")
	if err := wd.Get("https://twitter.com/login"); err != nil {
		log.Fatalf("Error navigating to Twitter login page: %v", err)
	}

	// Prompt the user to log in manually and check if the login was successful
	fmt.Println("Please log in to Twitter manually, then press Enter to continue...")
	fmt.Scanln()

	if err := checkLogin(wd); err != nil {
		log.Fatal("Login check failed. Ensure you have logged in before pressing Enter.")
	}

	// Navigate to the Twitter search page for the given search term
	log.Printf("Navigating to Twitter search page: %s", searchURL)
	if err := wd.Get(searchURL); err != nil {
		log.Fatalf("Error navigating to Twitter search page: %v", err)
	}

	// Scroll and wait for the page to load new tweets
	scrollAndWait(wd, 10) // Adjust the number of scrolls as needed

	// Check for new tweets
	checkTweets(wd, rdb, db, searchTerm)
}

// Check if the user is logged in
func checkLogin(wd selenium.WebDriver) error {
	// Implement login check logic
	return nil
}

// Scroll the page and wait for new tweets to load
func scrollAndWait(wd selenium.WebDriver, scrolls int) {
	for i := 0; i < scrolls; i++ {
		wd.ExecuteScript("window.scrollTo(0, document.body.scrollHeight);", nil)
		time.Sleep(2 * time.Second)
	}
}

// Check and save new tweets
func checkTweets(wd selenium.WebDriver, rdb *redis.Client, db *sql.DB, searchTerm string) {
	// Find tweet elements on the page
	tweetElements, err := wd.FindElements(selenium.ByCSSSelector, "article div[lang]")
	if err != nil {
		log.Fatalf("Error finding tweet elements: %v", err)
	}

	for _, tweetElement := range tweetElements {
		// Extract the tweet text
		tweetText, err := tweetElement.Text()
		if err != nil {
			log.Printf("Error extracting tweet text: %v", err)
			continue
		}

		// Generate a unique key for the tweet to check for duplicates
		tweetKey := fmt.Sprintf("tweet:%x", sha256.Sum256([]byte(tweetText)))

		// Check if the tweet is already in Redis (or any other method to avoid duplicates)
		exists, err := rdb.Exists(ctx, tweetKey).Result()
		if err != nil {
			log.Printf("Error checking if tweet exists in Redis: %v", err)
			continue
		}

		if exists > 0 {
			// Tweet already processed
			log.Printf("Tweet already processed: %s", tweetText)
			continue
		}

		// Store the tweet in Redis with an expiration time (optional)
		err = rdb.Set(ctx, tweetKey, true, 24*time.Hour).Err()
		if err != nil {
			log.Printf("Error saving tweet to Redis: %v", err)
		}

		// Insert the tweet into the PostgreSQL database
		_, err = db.Exec(`INSERT INTO tweets (content, search_term, tweet_time) VALUES ($1, $2, NOW())`, tweetText, searchTerm)
		if err != nil {
			log.Printf("Error inserting tweet into database: %v", err)
			continue
		}

		log.Printf("New tweet saved: %s", tweetText)
	}
}


