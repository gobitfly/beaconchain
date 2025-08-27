package testUtils

import (
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = "test/testEnv/.env"
	}

	copyIfMissing(envFile)

	err := godotenv.Load(envFile)
	if err != nil {
		log.Fatalf("❌ Error loading %s: %v", envFile, err)
	}

	baseURL := os.Getenv("BASE_URL")
	log.Printf("✅ Loaded env file: %s", envFile)
	log.Printf("🔍 BASE_URL=%s", baseURL)
}

func GetBaseURL() string {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://beaconcha.in"
	}
	return baseURL
}

func GetExternalGRPCUrl() string {
	baseURL := GetBaseURL()
	port := os.Getenv("EXTERNAL_GRPC_PORT")
	if port == "" {
		port = "9091"
	}
	return net.JoinHostPort(baseURL, port)
}

func GetInternalGRPCUrl() string {
	baseURL := GetBaseURL()
	port := os.Getenv("INTERNAL_GRPC_PORT")
	if port == "" {
		port = "9090"
	}
	return net.JoinHostPort(baseURL, port)
}

func IsValidBaseURL() (bool, string) {
	raw := os.Getenv("BASE_URL")
	if raw == "" {
		return false, "BASE_URL is empty"
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return false, "BASE_URL is not a valid URL"
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false, "BASE_URL must start with http or https"
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return false, "BASE_URL has no hostname"
	}

	if _, err := net.LookupHost(hostname); err != nil {
		return false, "BASE_URL hostname cannot be resolved: " + hostname
	}

	return true, ""
}

func GetDashboardID() int {
	dashboardIDStr := os.Getenv("DASHBOARD_ID")
	if dashboardIDStr == "" {
		log.Fatal("❌ DASHBOARD_ID not set in environment")
	}
	dashboardID, err := strconv.Atoi(dashboardIDStr)
	if err != nil {
		log.Fatalf("❌ Invalid DASHBOARD_ID: %v", err)
	}
	return dashboardID
}

func copyIfMissing(envFile string) {
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		dir := filepath.Dir(envFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0750); err != nil {
				log.Fatalf("❌ Failed to create env directory: %v", err)
			}
		}

		exampleFile := "test/.env.example"
		input, err := os.Open(exampleFile)
		if err != nil {
			log.Fatalf("❌ Example env file not found: %s", exampleFile)
		}

		output, err := os.Create(envFile) // #nosec G304 — safe internal usage
		if err != nil {
			_ = input.Close()
			log.Fatalf("❌ Failed to create target env file: %s", envFile)
		}

		if _, err = io.Copy(output, input); err != nil {
			_ = input.Close()
			_ = output.Close()
			log.Fatalf("❌ Failed to copy env file: %v", err)
		}

		_ = input.Close()
		_ = output.Close()

		log.Printf("📦 Copied .env.example → %s", envFile)
	}
}
