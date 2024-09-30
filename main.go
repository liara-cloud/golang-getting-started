package main

import (
    "fmt"
    "log"
    "os"
    "strconv"

    "github.com/joho/godotenv"
    gomail "gopkg.in/gomail.v2"
)

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Get environment variables for SMTP settings
    mailHost := os.Getenv("MAIL_HOST")
    mailPortStr := os.Getenv("MAIL_PORT")
    mailUser := os.Getenv("MAIL_USER")
    mailPassword := os.Getenv("MAIL_PASSWORD")
    mailFrom := os.Getenv("MAIL_FROM")

    // Convert mail port from string to integer
    mailPort, err := strconv.Atoi(mailPortStr)
    if err != nil {
        log.Fatalf("Invalid MAIL_PORT: %v", err)
    }

    // Create a new email message
    m := gomail.NewMessage()
    m.SetHeader("From", mailFrom)
    m.SetHeader("To", "recipient@example.com") // Set recipient email here
    m.SetHeader("Subject", "Test Email from Go with HTML")
    m.SetHeader("x-liara-tag", "test-tag") // Custom header for tagging

    // Set HTML body for the email
    m.SetBody("text/html", `
        <h1>This is a test email</h1>
        <p>Sent from Go using <b>gomail</b> and SMTP with TLS.</p>
    `)

    // Create SMTP dialer with TLS
    d := gomail.NewDialer(mailHost, mailPort, mailUser, mailPassword)

    // Send the email
    if err := d.DialAndSend(m); err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    fmt.Println("Test email sent successfully!")
}
