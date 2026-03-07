package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var papertrailToken string

func init() {
	godotenv.Load()
	papertrailToken = os.Getenv("PAPERTRAIL_API_TOKEN")

	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().StringP("db", "d", "", "SQLite database path (optional)")
	logsCmd.Flags().StringP("tsv", "t", "output.tsv", "Output TSV file path")
	logsCmd.Flags().StringP("hours", "n", "", "Hours ago (e.g., 24, 1-12)")
	logsCmd.Flags().StringP("date", "D", "", "Specific date (format: YYYY-MM-DD)")
	logsCmd.Flags().StringP("start", "s", "", "Start date for range (format: YYYY-MM-DD)")
	logsCmd.Flags().StringP("end", "e", "", "End date for range (format: YYYY-MM-DD)")
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Download logs from Papertrail",
	Long: `Download logs from Papertrail API.

Three modes:
  1. Hours mode: --hours 24 or --hours 1-12
  2. Single date: --date 2024-01-15
  3. Date range: --start 2024-01-01 --end 2024-01-15

Examples:
  papertail logs --hours 24 --tsv output.tsv
  papertail logs --hours 1-6 --tsv output.tsv
  papertail logs --date 2024-03-01 --tsv output.tsv
  papertail logs --start 2024-03-01 --end 2024-03-07 --tsv output.tsv`,
	Run: runLogs,
}

func runLogs(cmd *cobra.Command, args []string) {
	if papertrailToken == "" {
		log.Fatal("Error: PAPERTRAIL_API_TOKEN environment variable is required")
	}

	dbPath, _ := cmd.Flags().GetString("db")
	tsvPath, _ := cmd.Flags().GetString("tsv")
	hoursStr, _ := cmd.Flags().GetString("hours")
	dateStr, _ := cmd.Flags().GetString("date")
	startDate, _ := cmd.Flags().GetString("start")
	endDate, _ := cmd.Flags().GetString("end")

	mode := detectMode(hoursStr, dateStr, startDate, endDate)

	var err error
	switch mode {
	case "hours":
		startHour, endHour, err := parseHoursRange(hoursStr)
		if err != nil {
			log.Fatal(err)
		}
		err = downloadByHours(tsvPath, startHour, endHour)
	case "date":
		err = downloadByDate(tsvPath, dateStr)
	case "range":
		err = downloadByDateRange(tsvPath, startDate, endDate)
	default:
		log.Fatal("Error: specify --hours, --date, or --start/--end")
	}

	if err != nil {
		log.Fatalf("Error downloading logs: %v", err)
	}

	if dbPath != "" && fileExists(dbPath) {
		if err := importToSQLite(dbPath, tsvPath); err != nil {
			log.Printf("Warning: Failed to import to SQLite: %v", err)
		}
	}

	fmt.Println("Logs downloaded to:", tsvPath)
}

func detectMode(hours, date, start, end string) string {
	if hours != "" {
		return "hours"
	}
	if date != "" {
		return "date"
	}
	if start != "" && end != "" {
		return "range"
	}
	return ""
}

func parseHoursRange(s string) (int, int, error) {
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid hours format: %s", s)
		}
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid start hour: %v", err)
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid end hour: %v", err)
		}
		return start, end, nil
	}
	h, err := strconv.Atoi(s)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hours: %v", err)
	}
	return 1, h, nil
}

func downloadByHours(tsvPath string, startHour, endHour int) error {
	tmpDir, err := os.MkdirTemp("", "papertail-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := writeHeader(tsvPath); err != nil {
		return err
	}

	file, err := os.OpenFile(tsvPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	client := &http.Client{}
	for i := startHour; i <= endHour; i++ {
		dateStr := time.Now().UTC().Add(-time.Duration(i) * time.Hour).Format("2006-01-02-15")
		fmt.Printf("Downloading: %s\n", dateStr)
		if err := downloadAndExtract(client, dateStr, tmpDir, file); err != nil {
			log.Printf("Warning: failed to download %s: %v", dateStr, err)
		}
	}
	return nil
}

func downloadByDate(tsvPath, dateStr string) error {
	tmpDir, err := os.MkdirTemp("", "papertail-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := writeHeader(tsvPath); err != nil {
		return err
	}

	file, err := os.OpenFile(tsvPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	client := &http.Client{}
	dateWithHour := dateStr + "-00"
	if err := downloadAndExtract(client, dateWithHour, tmpDir, file); err != nil {
		return fmt.Errorf("failed to download %s: %v", dateWithHour, err)
	}
	return nil
}

func downloadByDateRange(tsvPath, startDay, endDay string) error {
	tmpDir, err := os.MkdirTemp("", "papertail-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := writeHeader(tsvPath); err != nil {
		return err
	}

	file, err := os.OpenFile(tsvPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	client := &http.Client{}

	archives, err := fetchArchives(client)
	if err != nil {
		return fmt.Errorf("failed to fetch archives: %w", err)
	}

	dates := filterArchives(archives, startDay, endDay)

	for _, d := range dates {
		fmt.Printf("Downloading: %s\n", d)
		if err := downloadAndExtract(client, d, tmpDir, file); err != nil {
			log.Printf("Warning: failed to download %s: %v", d, err)
		}
	}
	return nil
}

func writeHeader(tsvPath string) error {
	header := "id\tgenerated_at\treceived_at\tsource_id\tsource_name\tsource_ip\tfacility_name\tseverity_name\tprogram\tmessage\n"
	return os.WriteFile(tsvPath, []byte(header), 0644)
}

func fetchArchives(client *http.Client) ([]Archive, error) {
	req, err := http.NewRequest("GET", "https://papertrailapp.com/api/v1/archives.json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-papertrail-Token", papertrailToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch archives: status %d", resp.StatusCode)
	}

	var archives []Archive
	if err := json.NewDecoder(resp.Body).Decode(&archives); err != nil {
		return nil, err
	}
	return archives, nil
}

func filterArchives(archives []Archive, startDay, endDay string) []string {
	startDate, _ := time.Parse("2006-01-02", startDay)
	endDate, _ := time.Parse("2006-01-02", endDay)

	var result []string
	for _, a := range archives {
		dateStr := a.extractDate()
		if dateStr == "" {
			continue
		}
		archDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		if !archDate.Before(startDate) && archDate.Before(endDate) {
			result = append(result, a.Filename)
		}
	}
	return result
}

type Archive struct {
	Filename string `json:"filename"`
}

func (a Archive) extractDate() string {
	parts := strings.Split(a.Filename, "-")
	if len(parts) >= 3 {
		datePart := parts[0] + "-" + parts[1] + "-" + parts[2]
		if len(datePart) == 10 {
			return datePart
		}
	}
	return ""
}

func downloadAndExtract(client *http.Client, dateStr, tmpDir string, output *os.File) error {
	url := fmt.Sprintf("https://papertrailapp.com/api/v1/archives/%s/download", dateStr)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-papertrail-Token", papertrailToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	archivePath := filepath.Join(tmpDir, dateStr+".tsv.gz")
	out, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return err
	}

	return extractToFile(archivePath, output)
}

func extractToFile(archivePath string, output *os.File) error {
	cmd := exec.Command("7z", "x", "-so", archivePath)
	cmd.Stderr = os.Stderr

	pr, pw := io.Pipe()
	cmd.Stdout = pw

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start 7z: %w", err)
	}

	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		if _, err := output.WriteString(scanner.Text() + "\n"); err != nil {
			pw.Close()
			return err
		}
	}

	pw.Close()
	if err := cmd.Wait(); err != nil {
		if !strings.Contains(err.Error(), "No files to process") {
			return err
		}
	}
	return nil
}

func importToSQLite(dbPath, tsvPath string) error {
	cmd := exec.Command("sqlite-utils", "upsert", dbPath, "heroku-logs", "--pk=id", "--alter", "--tsv", tsvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sqlite-utils error: %w - %s", err, string(output))
	}
	fmt.Println("Imported to SQLite:", dbPath)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
