package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"vibeCodingLinkShortener/db"
	"vibeCodingLinkShortener/handlers"
	"vibeCodingLinkShortener/middleware"
)

func main() {
	// Initialize context
	ctx := context.Background()

	// In a real application, you might want to load .env file here
	// using something like github.com/joho/godotenv
	// For now, we assume DATABASE_URL is set in the environment.
	if os.Getenv("DATABASE_URL") == "" {
		log.Println("DATABASE_URL not set. Please set it before running the server.")
		// We'll set a default for local testing if not set
		os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/linkshortener?sslmode=disable")
	}

	// Initialize database
	if err := db.InitDB(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()
	log.Println("Successfully connected to the database")

	// Backfill missing log locations asynchronously
	handlers.BackfillLocations()

	// Set up router using standard library ServeMux
	mux := http.NewServeMux()

	// Register endpoints
	mux.HandleFunc("/shorten", handlers.HandleShorten)
	mux.HandleFunc("/api/stats", handlers.HandleStats)
	mux.HandleFunc("/api/register", handlers.HandleRegister)
	mux.HandleFunc("/api/login", handlers.HandleLogin)
	mux.HandleFunc("/api/logout", handlers.HandleLogout)
	mux.HandleFunc("/api/user", handlers.HandleUser)
	mux.HandleFunc("/api/unshorten", handlers.HandleUnshorten)
	mux.HandleFunc("/api/report", handlers.HandleReport)

	// Root handler to catch all other paths (which should be static files or short codes)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		// Serve index.html for root
		if path == "/" || path == "" {
			http.ServeFile(w, r, "./public/index.html")
			return
		}
		
		// Redirect /dashboard -> /dashboard/
		if path == "/dashboard" {
			http.Redirect(w, r, "/dashboard/", http.StatusMovedPermanently)
			return
		}
		if path == "/dashboard/" {
			http.ServeFile(w, r, "./public/dashboard.html")
			return
		}

		// Redirect /login -> /login/
		if path == "/login" {
			http.Redirect(w, r, "/login/", http.StatusMovedPermanently)
			return
		}
		if path == "/login/" {
			http.ServeFile(w, r, "./public/login.html")
			return
		}

		// Redirect /register -> /register/
		if path == "/register" {
			http.Redirect(w, r, "/register/", http.StatusMovedPermanently)
			return
		}
		if path == "/register/" {
			http.ServeFile(w, r, "./public/register.html")
			return
		}

		// Redirect /privacy -> /privacy/
		if path == "/privacy" {
			http.Redirect(w, r, "/privacy/", http.StatusMovedPermanently)
			return
		}
		if path == "/privacy/" {
			http.ServeFile(w, r, "./public/privacy.html")
			return
		}

		// Redirect /tos -> /tos/
		if path == "/tos" {
			http.Redirect(w, r, "/tos/", http.StatusMovedPermanently)
			return
		}
		if path == "/tos/" {
			http.ServeFile(w, r, "./public/tos.html")
			return
		}

		// Redirect /contact -> /contact/
		if path == "/contact" {
			http.Redirect(w, r, "/contact/", http.StatusMovedPermanently)
			return
		}
		if path == "/contact/" {
			http.ServeFile(w, r, "./public/contact.html")
			return
		}

		// Redirect /report -> /report/
		if path == "/report" {
			http.Redirect(w, r, "/report/", http.StatusMovedPermanently)
			return
		}
		if path == "/report/" {
			http.ServeFile(w, r, "./public/report.html")
			return
		}

		// Redirect /unshorten -> /unshorten/
		if path == "/unshorten" {
			http.Redirect(w, r, "/unshorten/", http.StatusMovedPermanently)
			return
		}
		if path == "/unshorten/" {
			http.ServeFile(w, r, "./public/unshorten.html")
			return
		}

		// Check if file exists in public directory
		filePath := "./public" + path
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}

		// Otherwise, attempt to redirect
		handlers.HandleRedirect(w, r)
	})

	// Wrap mux with logging middleware
	loggedMux := middleware.Logger(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, loggedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
