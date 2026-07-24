package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ezvacss.xyz/db"
	"ezvacss.xyz/shortener"
)

type ShortenRequest struct {
	URL            string `json:"url"`
	TurnstileToken string `json:"turnstile_token"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// HandleShorten accepts a POST request with a URL and returns a short code.
func HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Clean, prefix, and parse URL format
	rawURL := strings.TrimSpace(req.URL)
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" || (!strings.Contains(u.Host, ".") && u.Host != "localhost") {
		http.Error(w, "Invalid URL format. Please enter a valid web address (e.g. google.com).", http.StatusBadRequest)
		return
	}

	// Prevent redirect loops targeting our own service host
	if u.Host == r.Host || u.Host == "ezvacss.xyz" {
		http.Error(w, "Cannot shorten links targeting our own domain.", http.StatusBadRequest)
		return
	}

	// Verify target domain DNS existence to block dead / non-existent domains
	host := u.Hostname()
	if host != "localhost" && host != "127.0.0.1" {
		lookupCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		var resolver net.Resolver
		addrs, err := resolver.LookupHost(lookupCtx, host)
		if err != nil || len(addrs) == 0 {
			http.Error(w, "Target domain does not exist or cannot be reached. Please check the URL.", http.StatusBadRequest)
			return
		}
	}

	req.URL = rawURL

	userID := GetUserIDFromRequest(r)

	// Verify Turnstile Captcha for guest users (skip for registered/authenticated users)
	if userID == 0 {
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = r.RemoteAddr
		}
		ip = cleanIP(ip)

		valid, err := VerifyTurnstileToken(req.TurnstileToken, ip)
		if err != nil || !valid {
			log.Printf("Captcha verification failed for shorten: valid=%t, err=%v", valid, err)
			http.Error(w, "Captcha verification failed. Please try again.", http.StatusBadRequest)
			return
		}
	}

	ctx := context.Background()

	// Use random code generation for simplicity. Alternatively, insert then get ID and base62 encode it.
	shortCode := shortener.GenerateRandomCode(7)

	// Insert into urls table
	var dbErr error
	if userID > 0 {
		_, dbErr = db.Pool.Exec(ctx,
			"INSERT INTO urls (original_url, short_code, user_id) VALUES ($1, $2, $3)",
			req.URL, shortCode, userID)
	} else {
		_, dbErr = db.Pool.Exec(ctx,
			"INSERT INTO urls (original_url, short_code) VALUES ($1, $2)",
			req.URL, shortCode)
	}

	if dbErr != nil {
		log.Printf("Failed to insert URL: %v", dbErr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert into logs
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	_, err = db.Pool.Exec(ctx,
		"INSERT INTO logs (short_code, action, ip_address, user_agent) VALUES ($1, $2, $3, $4)",
		shortCode, "shorten", ip, userAgent)
	if err != nil {
		// Log but don't fail the request
		log.Printf("Failed to log shorten action: %v", err)
	}

	// Construct full short URL (detect HTTPS from proxy or force it for production domain)
	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	} else if strings.Contains(r.Host, "ezvacss.xyz") {
		scheme = "https"
	}
	domain := scheme + "://" + r.Host
	resp := ShortenResponse{
		ShortURL: domain + "/" + shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleRedirect handles GET requests to short codes and redirects to the original URL.
func HandleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract short code from path, e.g., "/K1jx3Cde1"
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Find original URL
	var originalURL string
	err := db.Pool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE short_code = $1", shortCode).Scan(&originalURL)

	if err != nil {
		// e.g. pgx.ErrNoRows
		log.Printf("Short code not found: %s", shortCode)
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Log the redirect asynchronously or synchronously
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	LogRedirectAsync(shortCode, "redirect", ip, userAgent)

	// Redirect to the original URL
	http.Redirect(w, r, originalURL, http.StatusFound) // 302 redirect
}

// cleanIP strips ports and extracts the client IP address from proxy headers.
func cleanIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if strings.Contains(ip, ",") {
		parts := strings.Split(ip, ",")
		ip = strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(ip)
	if err == nil {
		return host
	}
	return ip
}

// ResolveIPLocation fetches location from a free API on the backend.
func ResolveIPLocation(ip string) string {
	ip = cleanIP(ip)
	if ip == "" || ip == "::1" || ip == "127.0.0.1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.16.") {
		return "Localhost (Dev)"
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://freeipapi.com/api/json/" + ip)
	if err != nil {
		log.Printf("Failed to fetch geo IP: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var geo struct {
		CityName    string `json:"cityName"`
		CountryName string `json:"countryName"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return ""
	}

	if geo.CityName == "" && geo.CountryName == "" {
		return ""
	}

	city := geo.CityName
	if city == "" {
		city = "Unknown City"
	}
	country := geo.CountryName
	if country == "" {
		country = "Unknown Country"
	}

	return city + ", " + country
}

// LogRedirectAsync resolves geolocation and logs redirects in a non-blocking background goroutine.
func LogRedirectAsync(shortCode, action, ip, userAgent string) {
	go func() {
		location := ResolveIPLocation(ip)
		ctx := context.Background()
		_, err := db.Pool.Exec(ctx,
			"INSERT INTO logs (short_code, action, ip_address, user_agent, location) VALUES ($1, $2, $3, $4, $5)",
			shortCode, action, cleanIP(ip), userAgent, location)
		if err != nil {
			log.Printf("Failed to log redirect: %v", err)
		}
	}()
}

type LinkStat struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	Clicks      int    `json:"clicks"`
	CreatedAt   string `json:"created_at"`
}

type LogEntry struct {
	ID        int    `json:"id"`
	ShortCode string `json:"short_code"`
	Action    string `json:"action"`
	IPAddress string `json:"ip_address"`
	Location  string `json:"location"`
	UserAgent string `json:"user_agent"`
	CreatedAt string `json:"created_at"`
}

type StatsResponse struct {
	TotalLinks  int        `json:"total_links"`
	TotalClicks int        `json:"total_clicks"`
	Links       []LinkStat `json:"links"`
	RecentLogs  []LogEntry `json:"recent_logs"`
}

// HandleStats returns metrics and recent logs from the database for the dashboard.
func HandleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	var totalLinks int
	err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM urls WHERE user_id = $1", userID).Scan(&totalLinks)
	if err != nil {
		log.Printf("Failed to count URLs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var totalClicks int
	err = db.Pool.QueryRow(ctx, `
		SELECT COUNT(l.id) 
		FROM logs l
		JOIN urls u ON l.short_code = u.short_code
		WHERE u.user_id = $1 AND l.action = 'redirect'
	`, userID).Scan(&totalClicks)
	if err != nil {
		log.Printf("Failed to count logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch top links with click counts for this user
	rows, err := db.Pool.Query(ctx, `
		SELECT u.short_code, u.original_url, u.created_at, COUNT(l.id) as clicks
		FROM urls u
		LEFT JOIN logs l ON u.short_code = l.short_code AND l.action = 'redirect'
		WHERE u.user_id = $1
		GROUP BY u.short_code, u.original_url, u.created_at
		ORDER BY clicks DESC, u.created_at DESC
		LIMIT 20
	`, userID)
	if err != nil {
		log.Printf("Failed to fetch link stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var links []LinkStat
	for rows.Next() {
		var l LinkStat
		var t time.Time
		if err := rows.Scan(&l.ShortCode, &l.OriginalURL, &t, &l.Clicks); err != nil {
			log.Printf("Failed to scan link stat: %v", err)
			continue
		}
		l.CreatedAt = t.Format(time.RFC3339)
		links = append(links, l)
	}

	// Fetch recent action logs for this user's links
	logRows, err := db.Pool.Query(ctx, `
		SELECT l.id, l.short_code, l.action, l.ip_address, l.user_agent, l.location, l.created_at
		FROM logs l
		JOIN urls u ON l.short_code = u.short_code
		WHERE u.user_id = $1
		ORDER BY l.created_at DESC
		LIMIT 1000
	`, userID)
	if err != nil {
		log.Printf("Failed to fetch logs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer logRows.Close()

	var logsList []LogEntry
	for logRows.Next() {
		var entry LogEntry
		var t time.Time
		var ip, ua, loc *string
		if err := logRows.Scan(&entry.ID, &entry.ShortCode, &entry.Action, &ip, &ua, &loc, &t); err != nil {
			log.Printf("Failed to scan log: %v", err)
			continue
		}
		if ip != nil {
			entry.IPAddress = *ip
		}
		if ua != nil {
			entry.UserAgent = *ua
		}
		if loc != nil {
			entry.Location = *loc
		}
		entry.CreatedAt = t.Format(time.RFC3339)
		logsList = append(logsList, entry)
	}

	resp := StatsResponse{
		TotalLinks:  totalLinks,
		TotalClicks: totalClicks,
		Links:       links,
		RecentLogs:  logsList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type UnshortenResponse struct {
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}

// HandleUnshorten lets visitors preview the destination of a short code without redirecting.
func HandleUnshorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code parameter is required", http.StatusBadRequest)
		return
	}

	// Verify Turnstile Captcha for guest users if token provided
	userID := GetUserIDFromRequest(r)
	if userID == 0 {
		token := r.URL.Query().Get("turnstile_token")
		if token != "" {
			ip := r.Header.Get("CF-Connecting-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}
			ip = cleanIP(ip)

			valid, err := VerifyTurnstileToken(token, ip)
			if err != nil || !valid {
				log.Printf("Captcha verification failed for unshorten: valid=%t, err=%v", valid, err)
				http.Error(w, "Captcha verification failed. Please try again.", http.StatusBadRequest)
				return
			}
		}
	}

	// Clean code if it's a full URL
	code = strings.TrimSpace(code)
	code = strings.TrimSuffix(code, "/")
	if strings.Contains(code, "/") {
		parts := strings.Split(code, "/")
		code = parts[len(parts)-1]
	}

	ctx := context.Background()
	var originalURL string
	var createdAt time.Time

	err := db.Pool.QueryRow(ctx,
		"SELECT original_url, created_at FROM urls WHERE short_code = $1",
		code).Scan(&originalURL, &createdAt)

	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	resp := UnshortenResponse{
		OriginalURL: originalURL,
		CreatedAt:   createdAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type ReportRequest struct {
	Code           string `json:"code"`
	Reason         string `json:"reason"`
	ReporterName   string `json:"reporter_name"`
	ReporterEmail  string `json:"reporter_email"`
	ReportType     string `json:"report_type"`
	TurnstileToken string `json:"turnstile_token"`
}

// VerifyTurnstileToken validates the Cloudflare Turnstile token with Cloudflare API.
func VerifyTurnstileToken(token, clientIP string) (bool, error) {
	secret := os.Getenv("TURNSTILE_SECRET_KEY")
	if secret == "" {
		log.Println("TURNSTILE_SECRET_KEY not set. Skipping Turnstile verification.")
		return true, nil
	}

	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
		"secret":   {secret},
		"response": {token},
		"remoteip": {clientIP},
	})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var result struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Success, nil
}

// HandleReport logs an abuse report for a malicious shortened URL.
func HandleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.Reason == "" || req.ReportType == "" {
		http.Error(w, "Code, reason, and report type are required", http.StatusBadRequest)
		return
	}

	// Verify Turnstile Captcha
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	ip = cleanIP(ip)

	valid, err := VerifyTurnstileToken(req.TurnstileToken, ip)
	if err != nil || !valid {
		log.Printf("Captcha verification failed: valid=%t, err=%v", valid, err)
		http.Error(w, "Captcha verification failed. Please try again.", http.StatusBadRequest)
		return
	}

	// Clean code if it's a full URL
	code := req.Code
	if strings.Contains(code, "/") {
		parts := strings.Split(code, "/")
		code = parts[len(parts)-1]
	}

	ctx := context.Background()

	// Verify short URL exists
	var exists bool
	err = db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)", code).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	// Insert report into Postgres
	_, err = db.Pool.Exec(ctx,
		"INSERT INTO reports (short_code, reason, reporter_name, reporter_email, report_type) VALUES ($1, $2, $3, $4, $5)",
		code, req.Reason, req.ReporterName, req.ReporterEmail, req.ReportType)

	if err != nil {
		log.Printf("Failed to insert report: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send detailed email notification to admin (asynchronously)
	SendDetailedReportEmail(code, req.Reason, req.ReporterName, req.ReporterEmail, req.ReportType)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Report submitted successfully"})
}

// BackfillLocations finds any logs with missing locations on server startup and resolves them.
func BackfillLocations() {
	go func() {
		// Wait a few seconds for database initialization to completely settle
		time.Sleep(3 * time.Second)
		ctx := context.Background()

		// Retrieve all log entries with missing location fields
		rows, err := db.Pool.Query(ctx, "SELECT id, ip_address FROM logs WHERE location IS NULL OR location = ''")
		if err != nil {
			log.Printf("Failed to select logs for backfill: %v", err)
			return
		}
		defer rows.Close()

		type LogJob struct {
			ID        int
			IPAddress string
		}
		var jobs []LogJob
		for rows.Next() {
			var job LogJob
			if err := rows.Scan(&job.ID, &job.IPAddress); err == nil {
				jobs = append(jobs, job)
			}
		}

		if len(jobs) == 0 {
			return
		}

		log.Printf("Backfilling locations for %d database log entries...", len(jobs))
		for _, job := range jobs {
			if job.IPAddress == "" {
				continue
			}
			location := ResolveIPLocation(job.IPAddress)
			if location != "" {
				_, err = db.Pool.Exec(ctx, "UPDATE logs SET location = $1 WHERE id = $2", location, job.ID)
				if err != nil {
					log.Printf("Failed to update log location: %v", err)
				}
				// Sleep to respect freeipapi.com free tier rate limits
				time.Sleep(1200 * time.Millisecond)
			}
		}
		log.Println("Database locations backfill completed!")
	}()
}
