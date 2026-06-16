package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	// Ensure no real env vars leak into the defaults under test.
	for _, k := range []string{
		"APP_ENV", "SERVER_PORT", "DB_HOST", "DB_PORT", "DB_USER",
		"DB_PASSWORD", "DB_NAME", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS",
		"DB_CONN_MAX_LIFETIME_MINUTES", "CORS_ALLOWED_ORIGINS",
		"BACKEND_BASE_URL", "UPLOAD_DIR",
	} {
		t.Setenv(k, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want development", cfg.AppEnv)
	}
	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort = %q, want 8080", cfg.ServerPort)
	}
	if cfg.DBName != "iclassroom" {
		t.Errorf("DBName = %q, want iclassroom", cfg.DBName)
	}
	if cfg.DBMaxOpenConns != 25 {
		t.Errorf("DBMaxOpenConns = %d, want 25", cfg.DBMaxOpenConns)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:5173" {
		t.Errorf("CORSAllowedOrigins = %v, want [http://localhost:5173]", cfg.CORSAllowedOrigins)
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() = true, want false for development")
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DB_NAME", "classroom_prod")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example.com, https://b.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if !cfg.IsProduction() {
		t.Error("IsProduction() = false, want true")
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("ServerPort = %q, want 9090", cfg.ServerPort)
	}
	if got, want := len(cfg.CORSAllowedOrigins), 2; got != want {
		t.Fatalf("len(CORSAllowedOrigins) = %d, want %d", got, want)
	}
	if cfg.CORSAllowedOrigins[1] != "https://b.example.com" {
		t.Errorf("CORSAllowedOrigins[1] = %q, want https://b.example.com (trimmed)", cfg.CORSAllowedOrigins[1])
	}
}

func TestValidateRejectsEmptyDBName(t *testing.T) {
	t.Setenv("DB_NAME", "")
	// getEnv falls back to default for empty values, so force a failure path
	// directly through validate to confirm it guards required fields.
	cfg := &Config{
		ServerPort:         "8080",
		DBUser:             "root",
		DBName:             "",
		DBMaxOpenConns:     10,
		DBMaxIdleConns:     10,
		CORSAllowedOrigins: []string{"http://localhost:5173"},
		BackendBaseURL:     "http://localhost:8080",
		UploadDir:          "./uploads",
	}
	if err := cfg.validate(); err == nil {
		t.Error("validate() = nil, want error for empty DBName")
	}
}

func TestDBDSN(t *testing.T) {
	cfg := &Config{
		DBHost: "127.0.0.1", DBPort: "3306",
		DBUser: "root", DBPassword: "secret", DBName: "iclassroom",
		BackendBaseURL: "http://localhost:8080",
		UploadDir:      "./uploads",
	}
	want := "root:secret@tcp(127.0.0.1:3306)/iclassroom?charset=utf8mb4&parseTime=true&loc=UTC"
	if got := cfg.DBDSN(); got != want {
		t.Errorf("DBDSN() = %q, want %q", got, want)
	}
}
