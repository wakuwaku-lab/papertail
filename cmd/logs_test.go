package cmd

import (
	"compress/gzip"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestImportToSQLite(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "papertail-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	tsvPath := filepath.Join(tmpDir, "test.tsv")

	// 1. Create a mock TSV file
	// Header: id, generated_at, received_at, source_id, source_name, source_ip, facility_name, severity_name, program, message
	header := "id\tgenerated_at\treceived_at\tsource_id\tsource_name\tsource_ip\tfacility_name\tseverity_name\tprogram\tmessage\n"
	row1 := "12345\t2024-03-08T10:00:00Z\t2024-03-08T10:00:01Z\t101\tserver-1\t192.168.1.1\tUser\tInfo\tapp\tHello World\n"
	row2 := "12346\t2024-03-08T10:05:00Z\t2024-03-08T10:05:01Z\t102\tserver-2\t192.168.1.2\tSystem\tError\tauth\tLogin failed\n"
	
	tsvContent := header + row1 + row2
	if err := os.WriteFile(tsvPath, []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write mock TSV: %v", err)
	}

	// 2. Run the import function
	if err := importToSQLite(dbPath, tsvPath); err != nil {
		t.Fatalf("importToSQLite failed: %v", err)
	}

	// 3. Verify the data in SQLite
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open resulting DB: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM \"heroku-logs\"").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query DB: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}

	// Verify specific content
	var message string
	err = db.QueryRow("SELECT message FROM \"heroku-logs\" WHERE id = 12346").Scan(&message)
	if err != nil {
		t.Fatalf("Failed to query specific row: %v", err)
	}

	if message != "Login failed" {
		t.Errorf("Expected 'Login failed', got '%s'", message)
	}

	// 4. Test "Upsert" (Replace) logic
	row2Updated := "12346\t2024-03-08T10:05:00Z\t2024-03-08T10:05:01Z\t102\tserver-2\t192.168.1.2\tSystem\tError\tauth\tLogin failed - RETRY\n"
	tsvContent2 := header + row2Updated
	if err := os.WriteFile(tsvPath, []byte(tsvContent2), 0644); err != nil {
		t.Fatalf("Failed to write updated mock TSV: %v", err)
	}

	if err := importToSQLite(dbPath, tsvPath); err != nil {
		t.Fatalf("Second importToSQLite failed: %v", err)
	}

	err = db.QueryRow("SELECT message FROM \"heroku-logs\" WHERE id = 12346").Scan(&message)
	if err != nil {
		t.Fatalf("Failed to query updated row: %v", err)
	}

	if message != "Login failed - RETRY" {
		t.Errorf("Expected 'Login failed - RETRY' after upsert, got '%s'", message)
	}
}

func TestExtractToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "papertail-gzip-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := "test content for gzip"
	gzipPath := filepath.Join(tmpDir, "test.gz")
	outputPath := filepath.Join(tmpDir, "output.txt")

	// Create gzip file
	f, err := os.Create(gzipPath)
	if err != nil {
		t.Fatalf("Failed to create gzip file: %v", err)
	}
	gw := gzip.NewWriter(f)
	if _, err := gw.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to gzip: %v", err)
	}
	gw.Close()
	f.Close()

	// Extract
	outF, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("Failed to create output file: %v", err)
	}
	defer outF.Close()

	if err := extractToFile(gzipPath, outF); err != nil {
		t.Fatalf("extractToFile failed: %v", err)
	}
	outF.Close()

	// Verify
	extracted, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(extracted) != content {
		t.Errorf("Expected '%s', got '%s'", content, string(extracted))
	}
}
