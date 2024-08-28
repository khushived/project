package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
)

const chromeDriverPath = "C:\\Users\\HP\\Downloads\\chromedriver-win64\\chromedriver.exe" // Adjust the path as needed

func main() {
    // Check if the file exists at the specified path
    if _, err := os.Stat(chromeDriverPath); os.IsNotExist(err) {
        log.Fatalf("File not found: %s", chromeDriverPath)
    }

    // Read the file
    data, err := ioutil.ReadFile(chromeDriverPath)
    if err != nil {
        log.Fatalf("Error reading file: %v", err)
    }

    // Print a success message with the size of the file
    fmt.Printf("Successfully read file: %s (size: %d bytes)\n", chromeDriverPath, len(data))
}
