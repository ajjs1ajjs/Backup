// Reports & Analytics - Veeam-style reporting
package reports

import (
	"fmt"
	"novabackup/internal/database"
	"time"
)

// Report types
const (
	ReportDaily   = "daily"
	ReportWeekly  = "weekly"
	ReportMonthly = "monthly"
	ReportCustom  = "custom"
)

// ReportData represents report data
type ReportData struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	GeneratedAt time.Time              `json:"generated_at"`
	PeriodStart time.Time              `json:"period_start"`
	PeriodEnd   time.Time              `json:"period_end"`
	Summary     map[string]interface{} `json:"summary"`
	Charts      []ChartData            `json:"charts"`
	Tables      []TableData            `json:"tables"`
}

// ChartData represents chart data for visualization
type ChartData struct {
	Title  string    `json:"title"`
	Type   string    `json:"type"` // bar, line, pie
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
}

// TableData represents table data
type TableData struct {
	Title   string     `json:"title"`
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// Statistics represents system statistics
type Statistics struct {
	TotalJobs        int       `json:"total_jobs"`
	ActiveJobs       int       `json:"active_jobs"`
	FailedJobs       int       `json:"failed_jobs"`
	TotalSessions    int       `json:"total_sessions"`
	SuccessSessions  int       `json:"success_sessions"`
	FailedSessions   int       `json:"failed_sessions"`
	TotalBackups     int64     `json:"total_backups"`
	TotalRestores    int64     `json:"total_restores"`
	AvgDuration      string    `json:"avg_duration"`
	SuccessRate      float64   `json:"success_rate"`
	StorageUsed      int64     `json:"storage_used"`
	StorageFree      int64     `json:"storage_free"`
	DedupRatio       float64   `json:"dedup_ratio"`
	CompressionRatio float64   `json:"compression_ratio"`
	LastUpdated      time.Time `json:"last_updated"`
}

// ReportEngine generates reports and analytics
type ReportEngine struct {
	db *database.Database
}

// NewReportEngine creates a new report engine
func NewReportEngine(db *database.Database) *ReportEngine {
	return &ReportEngine{
		db: db,
	}
}

// GenerateDailyReport generates a daily report
func (e *ReportEngine) GenerateDailyReport(date time.Time) (*ReportData, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	return e.generateReport(ReportDaily, "Щоденний звіт", start, end)
}

// GenerateWeeklyReport generates a weekly report
func (e *ReportEngine) GenerateWeeklyReport(date time.Time) (*ReportData, error) {
	// Get Monday of the week
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start := time.Date(date.Year(), date.Month(), date.Day()-weekday+1, 0, 0, 0, 0, date.Location())
	end := start.Add(7 * 24 * time.Hour)

	return e.generateReport(ReportWeekly, "Тижневий звіт", start, end)
}

// GenerateMonthlyReport generates a monthly report
func (e *ReportEngine) GenerateMonthlyReport(date time.Time) (*ReportData, error) {
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	end := start.AddDate(0, 1, 0)

	return e.generateReport(ReportMonthly, "Місячний звіт", start, end)
}

// generateReport generates a report for a specific period
func (e *ReportEngine) generateReport(reportType, title string, start, end time.Time) (*ReportData, error) {
	report := &ReportData{
		ID:          fmt.Sprintf("report_%s_%d", reportType, start.Unix()),
		Type:        reportType,
		Title:       title,
		GeneratedAt: time.Now(),
		PeriodStart: start,
		PeriodEnd:   end,
		Summary:     make(map[string]interface{}),
		Charts:      make([]ChartData, 0),
		Tables:      make([]TableData, 0),
	}

	// Load sessions for the period
	sessions, err := e.db.ListSessions()
	if err != nil {
		return nil, err
	}

	// Filter sessions by period
	var filteredSessions []database.Session
	for _, s := range sessions {
		if s.StartTime.After(start) && s.StartTime.Before(end) {
			filteredSessions = append(filteredSessions, s)
		}
	}

	// Calculate statistics
	successCount := 0
	failedCount := 0
	totalBytes := int64(0)
	totalFiles := 0
	totalDuration := time.Duration(0)

	for _, s := range filteredSessions {
		if s.Status == "success" {
			successCount++
		} else {
			failedCount++
		}
		totalBytes += s.BytesWritten
		totalFiles += s.FilesProcessed
		totalDuration += s.EndTime.Sub(s.StartTime)
	}

	// Summary
	report.Summary["total_sessions"] = len(filteredSessions)
	report.Summary["success_sessions"] = successCount
	report.Summary["failed_sessions"] = failedCount
	report.Summary["success_rate"] = float64(0)
	if len(filteredSessions) > 0 {
		report.Summary["success_rate"] = float64(successCount) / float64(len(filteredSessions)) * 100
	}
	report.Summary["total_bytes"] = totalBytes
	report.Summary["total_files"] = totalFiles
	report.Summary["avg_duration"] = ""
	if len(filteredSessions) > 0 {
		report.Summary["avg_duration"] = (totalDuration / time.Duration(len(filteredSessions))).String()
	}

	// Chart: Sessions by day
	chartData := e.generateSessionsByDayChart(filteredSessions, start, end)
	report.Charts = append(report.Charts, chartData)

	// Chart: Storage usage over time
	storageChart := e.generateStorageUsageChart(filteredSessions)
	report.Charts = append(report.Charts, storageChart)

	// Table: Job statistics
	jobTable := e.generateJobStatisticsTable(filteredSessions)
	report.Tables = append(report.Tables, jobTable)

	// Table: Failed sessions
	if failedCount > 0 {
		failedTable := e.generateFailedSessionsTable(filteredSessions)
		report.Tables = append(report.Tables, failedTable)
	}

	return report, nil
}

// generateSessionsByDayChart generates chart data for sessions by day
func (e *ReportEngine) generateSessionsByDayChart(sessions []database.Session, start, end time.Time) ChartData {
	days := int(end.Sub(start).Hours() / 24)
	labels := make([]string, days)
	data := make([]float64, days)

	for i := 0; i < days; i++ {
		dayStart := start.AddDate(0, 0, i)
		dayEnd := dayStart.Add(24 * time.Hour)

		labels[i] = dayStart.Format("02.01")

		count := 0
		for _, s := range sessions {
			if s.StartTime.After(dayStart) && s.StartTime.Before(dayEnd) {
				count++
			}
		}
		data[i] = float64(count)
	}

	return ChartData{
		Title:  "Сесії по днях",
		Type:   "bar",
		Labels: labels,
		Data:   data,
	}
}

// generateStorageUsageChart generates chart data for storage usage
func (e *ReportEngine) generateStorageUsageChart(sessions []database.Session) ChartData {
	labels := make([]string, 0)
	data := make([]float64, 0)

	cumulative := int64(0)
	for _, s := range sessions {
		cumulative += s.BytesWritten
		labels = append(labels, s.StartTime.Format("02.01 15:04"))
		data = append(data, float64(cumulative)/1024/1024/1024) // GB
	}

	return ChartData{
		Title:  "Використання сховища (GB)",
		Type:   "line",
		Labels: labels,
		Data:   data,
	}
}

// generateJobStatisticsTable generates table data for job statistics
func (e *ReportEngine) generateJobStatisticsTable(sessions []database.Session) TableData {
	// Group by job
	jobStats := make(map[string][]database.Session)
	for _, s := range sessions {
		jobStats[s.JobName] = append(jobStats[s.JobName], s)
	}

	headers := []string{"Завдання", "Всього сесій", "Успішно", "Невдало", "Останній запуск", "Середній розмір"}
	rows := make([][]string, 0)

	for jobName, jobSessions := range jobStats {
		success := 0
		failed := 0
		totalBytes := int64(0)
		lastRun := time.Time{}

		for _, s := range jobSessions {
			if s.Status == "success" {
				success++
			} else {
				failed++
			}
			totalBytes += s.BytesWritten
			if s.StartTime.After(lastRun) {
				lastRun = s.StartTime
			}
		}

		avgSize := ""
		if len(jobSessions) > 0 {
			avgSize = formatBytes(totalBytes / int64(len(jobSessions)))
		}

		rows = append(rows, []string{
			jobName,
			fmt.Sprintf("%d", len(jobSessions)),
			fmt.Sprintf("%d", success),
			fmt.Sprintf("%d", failed),
			lastRun.Format("02.01.2006 15:04"),
			avgSize,
		})
	}

	return TableData{
		Title:   "Статистика по завданнях",
		Headers: headers,
		Rows:    rows,
	}
}

// generateFailedSessionsTable generates table data for failed sessions
func (e *ReportEngine) generateFailedSessionsTable(sessions []database.Session) TableData {
	headers := []string{"Завдання", "Час початку", "Тривалість", "Помилка"}
	rows := make([][]string, 0)

	for _, s := range sessions {
		if s.Status == "failed" {
			duration := s.EndTime.Sub(s.StartTime).String()
			rows = append(rows, []string{
				s.JobName,
				s.StartTime.Format("02.01.2006 15:04"),
				duration,
				s.Error,
			})
		}
	}

	return TableData{
		Title:   "Невдалі сесії",
		Headers: headers,
		Rows:    rows,
	}
}

// GetStatistics returns current system statistics
func (e *ReportEngine) GetStatistics() (*Statistics, error) {
	sessions, err := e.db.ListSessions()
	if err != nil {
		return nil, err
	}

	jobs, err := e.db.ListJobs()
	if err != nil {
		return nil, err
	}

	stats := &Statistics{
		TotalJobs:   len(jobs),
		LastUpdated: time.Now(),
	}

	// Count active jobs
	for _, job := range jobs {
		if job.Enabled {
			stats.ActiveJobs++
		}
	}

	// Calculate session statistics
	for _, s := range sessions {
		if s.Status == "success" {
			stats.SuccessSessions++
			stats.TotalBackups += s.BytesWritten
		} else {
			stats.FailedSessions++
		}
	}

	stats.TotalSessions = len(sessions)
	stats.FailedJobs = stats.FailedSessions

	// Calculate success rate
	if stats.TotalSessions > 0 {
		stats.SuccessRate = float64(stats.SuccessSessions) / float64(stats.TotalSessions) * 100
	}

	// Calculate average duration
	if stats.TotalSessions > 0 {
		totalDuration := time.Duration(0)
		for _, s := range sessions {
			totalDuration += s.EndTime.Sub(s.StartTime)
		}
		stats.AvgDuration = (totalDuration / time.Duration(stats.TotalSessions)).String()
	}

	// Storage stats (simplified)
	stats.StorageUsed = stats.TotalBackups
	stats.StorageFree = 1024 * 1024 * 1024 * 500 // 500GB default
	stats.DedupRatio = 2.5                       // Example ratio
	stats.CompressionRatio = 3.2                 // Example ratio

	return stats, nil
}

// GetTrends returns trend data for charts
func (e *ReportEngine) GetTrends(days int) map[string]interface{} {
	end := time.Now()
	start := end.AddDate(0, 0, -days)

	sessions, _ := e.db.ListSessions()

	// Filter by period
	var filtered []database.Session
	for _, s := range sessions {
		if s.StartTime.After(start) && s.StartTime.Before(end) {
			filtered = append(filtered, s)
		}
	}

	// Group by day
	dailyStats := make(map[string]map[string]int)
	for i := 0; i < days; i++ {
		day := end.AddDate(0, 0, -i)
		dayStr := day.Format("02.01")
		dailyStats[dayStr] = map[string]int{"success": 0, "failed": 0}
	}

	for _, s := range filtered {
		dayStr := s.StartTime.Format("02.01")
		if _, exists := dailyStats[dayStr]; exists {
			if s.Status == "success" {
				dailyStats[dayStr]["success"]++
			} else {
				dailyStats[dayStr]["failed"]++
			}
		}
	}

	return map[string]interface{}{
		"period_days":    days,
		"daily_stats":    dailyStats,
		"total_sessions": len(filtered),
	}
}

// ExportReport exports report to JSON/HTML/PDF
func (e *ReportEngine) ExportReport(report *ReportData, format string) ([]byte, error) {
	switch format {
	case "json":
		return e.exportJSON(report)
	case "html":
		return e.exportHTML(report)
	case "pdf":
		return e.exportPDF(report)
	default:
		return nil, fmt.Errorf("непідтримуваний формат: %s", format)
	}
}

func (e *ReportEngine) exportJSON(report *ReportData) ([]byte, error) {
	// In production, use json.Marshal
	return []byte("{}"), nil
}

func (e *ReportEngine) exportHTML(report *ReportData) ([]byte, error) {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>%s</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		h1 { color: #333; }
		.summary { background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 20px 0; }
		table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
		th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
		th { background: #4CAF50; color: white; }
		tr:nth-child(even) { background: #f2f2f2; }
	</style>
</head>
<body>
	<h1>%s</h1>
	<p>Період: %s - %s</p>
	<p>Згенеровано: %s</p>

	<div class="summary">
		<h2>Підсумки</h2>
		<p>Всього сесій: %v</p>
		<p>Успішно: %v</p>
		<p>Невдало: %v</p>
	</div>
</body>
</html>
`, report.Title, report.Title,
		report.PeriodStart.Format("02.01.2006"),
		report.PeriodEnd.Format("02.01.2006"),
		report.GeneratedAt.Format("02.01.2006 15:04"),
		report.Summary["total_sessions"],
		report.Summary["success_sessions"],
		report.Summary["failed_sessions"])

	return []byte(html), nil
}

func (e *ReportEngine) exportPDF(report *ReportData) ([]byte, error) {
	// In production, use PDF library
	return []byte{}, nil
}

// Helper function
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
