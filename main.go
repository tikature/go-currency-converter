package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	API_KEY   = "99af1e52e8b504f480478eda"
	BASE_URL  = "https://v6.exchangerate-api.com/v6/%s/pair"
	CODES_URL = "https://v6.exchangerate-api.com/v6/%s/codes"
	PORT      = ":8080"
)

// Struct untuk response mata uang
type CurrencyCodesResponse struct {
	Result        string     `json:"result"`
	Documentation string     `json:"documentation"`
	TermsOfUse    string     `json:"terms_of_use"`
	TimeLastUpdateUtc string `json:"time_last_update_utc"`
	TimeNextUpdateUtc string `json:"time_next_update_utc"`
	SupportedCodes [][]string `json:"supported_codes"`
}

// Struct untuk response konversi
type ConversionResponse struct {
	Result           string  `json:"result"`
	Documentation    string  `json:"documentation"`
	TermsOfUse       string  `json:"terms_of_use"`
	TimeLastUpdateUtc string `json:"time_last_update_utc"`
	TimeNextUpdateUtc string `json:"time_next_update_utc"`
	BaseCode         string  `json:"base_code"`
	TargetCode       string  `json:"target_code"`
	ConversionRate   float64 `json:"conversion_rate"`
	ConversionResult float64 `json:"conversion_result"`
}

// Struct untuk error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Struct untuk success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Middleware untuk enable CORS
func enableCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	
	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return false
	}
	return true
}

// Middleware untuk logging
func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}

// Handler untuk mendapatkan semua mata uang
func getCurrenciesHandler(w http.ResponseWriter, r *http.Request) {
	if !enableCORS(w, r) {
		return
	}

	log.Println("📋 Request: Get All Currencies")
	
	url := fmt.Sprintf(CODES_URL, API_KEY)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Error fetching currencies: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "FETCH_ERROR",
			Message: "Gagal mengambil data mata uang dari server",
		})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error reading response body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "READ_ERROR", 
			Message: "Gagal membaca data dari server",
		})
		return
	}

	var data CurrencyCodesResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "PARSE_ERROR",
			Message: "Gagal memproses data mata uang",
		})
		return
	}

	if data.Result != "success" {
		log.Printf("❌ API returned error: %s", data.Result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "API_ERROR",
			Message: "API gagal memberikan data mata uang",
		})
		return
	}

	log.Printf("✅ Successfully fetched %d currencies", len(data.SupportedCodes))
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
		Message: fmt.Sprintf("Berhasil mendapatkan %d mata uang", len(data.SupportedCodes)),
	})
}

// Handler untuk konversi mata uang
func convertHandler(w http.ResponseWriter, r *http.Request) {
	if !enableCORS(w, r) {
		return
	}

	// Parse query parameters
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")

	log.Printf("💱 Request: Convert %s %s to %s", amountStr, from, to)

	// Validasi input
	if from == "" || to == "" || amountStr == "" {
		log.Println("❌ Missing required parameters")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "MISSING_PARAMETERS",
			Message: "Parameter from, to, dan amount wajib diisi",
		})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Printf("❌ Invalid amount: %s", amountStr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INVALID_AMOUNT",
			Message: "Jumlah yang dimasukkan tidak valid",
		})
		return
	}

	if amount <= 0 {
		log.Printf("❌ Amount must be positive: %f", amount)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INVALID_AMOUNT",
			Message: "Jumlah harus lebih besar dari 0",
		})
		return
	}

	// Call exchange rate API
	url := fmt.Sprintf(BASE_URL+"/%s/%s/%.2f", API_KEY, from, to, amount)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Error calling conversion API: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "CONVERSION_ERROR",
			Message: "Gagal melakukan konversi mata uang",
		})
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Error reading conversion response: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "READ_ERROR",
			Message: "Gagal membaca hasil konversi",
		})
		return
	}

	var data ConversionResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("❌ Error parsing conversion JSON: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "PARSE_ERROR",
			Message: "Gagal memproses hasil konversi",
		})
		return
	}

	if data.Result != "success" {
		log.Printf("❌ Conversion API error: %s", data.Result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "CONVERSION_FAILED",
			Message: "Konversi mata uang gagal, periksa kode mata uang",
		})
		return
	}

	log.Printf("✅ Conversion successful: %.2f %s = %.2f %s (rate: %.4f)", 
		amount, from, data.ConversionResult, to, data.ConversionRate)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Success: true,
		Data:    data,
		Message: fmt.Sprintf("Konversi berhasil: %.2f %s = %.2f %s", 
			amount, from, data.ConversionResult, to),
	})
}

// Handler untuk health check
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if !enableCORS(w, r) {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "Currency Converter API",
	})
}

// Handler untuk serve static files
func staticHandler(w http.ResponseWriter, r *http.Request) {
	// Set proper content type for HTML
	if r.URL.Path == "/" || r.URL.Path == "/index.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	
	// Serve static files from current directory
	http.FileServer(http.Dir("./")).ServeHTTP(w, r)
}

// Main function
func main() {
	// Print startup banner
	fmt.Println("🏆 ===============================================")
	fmt.Println("💰 GoldEx - Premium Currency Converter Server")
	fmt.Println("🏆 ===============================================")
	fmt.Println("🚀 Starting server...")
	fmt.Printf("📡 Server will run on http://localhost%s\n", PORT)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("   📄 GET  /                    - Landing Page")
	fmt.Println("   💱 GET  /api/convert         - Currency Conversion")
	fmt.Println("   📋 GET  /api/currencies      - Get All Currencies")
	fmt.Println("   ❤️  GET  /api/health         - Health Check")
	fmt.Println("🏆 ===============================================")

	// Setup routes dengan middleware logging
	http.HandleFunc("/api/currencies", logRequest(getCurrenciesHandler))
	http.HandleFunc("/api/convert", logRequest(convertHandler))
	http.HandleFunc("/api/health", logRequest(healthHandler))
	
	// Serve static files (HTML, CSS, JS)
	http.HandleFunc("/", staticHandler)

	// Start server
	log.Printf("🚀 Server starting on port%s", PORT)
	log.Printf("📄 Landing page: http://localhost%s", PORT)
	log.Printf("📡 API Base URL: http://localhost%s/api", PORT)
	
	if err := http.ListenAndServe(PORT, nil); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}