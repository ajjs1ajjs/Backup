// Ransomware Detection - Detect suspicious backup patterns
package backup

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// RansomwareAlert represents a ransomware detection alert
type RansomwareAlert struct {
	Level       string                 `json:"level"` // critical, high, medium, low
	Score       int                    `json:"score"` // 0-100
	Indicators  []string               `json:"indicators"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	Recommended []string               `json:"recommended"`
}

// RansomwareDetector analyzes backups for ransomware signs
type RansomwareDetector struct {
	PreviousBackup *BackupSession
	CurrentBackup  *BackupSession
	Thresholds     RansomwareThresholds
}

// RansomwareThresholds defines detection thresholds
type RansomwareThresholds struct {
	EncryptedExtensionCount  int     // Number of files with encrypted extensions
	ChangedFilesPercent      float64 // Percentage of changed files
	EntropyIncrease          float64 // Average entropy increase
	DeletedFilesPercent      float64 // Percentage of deleted files
	FileSizeReductionPercent float64 // Average file size reduction
}

// DefaultThresholds returns default detection thresholds
func DefaultThresholds() RansomwareThresholds {
	return RansomwareThresholds{
		EncryptedExtensionCount:  100,
		ChangedFilesPercent:      50.0,
		EntropyIncrease:          0.3,
		DeletedFilesPercent:      30.0,
		FileSizeReductionPercent: 20.0,
	}
}

// Known ransomware file extensions
var ransomwareExtensions = map[string]bool{
	".encrypted": true,
	".locked":    true,
	".crypto":    true,
	".crypt":     true,
	".lockbit":   true,
	".conti":     true,
	".revil":     true,
	".wannacry":  true,
	".petya":     true,
	".ryuk":      true,
}

// Analyze analyzes current backup for ransomware indicators
func (d *RansomwareDetector) Analyze() *RansomwareAlert {
	alert := &RansomwareAlert{
		Level:      "low",
		Score:      0,
		Indicators: make([]string, 0),
		Details:    make(map[string]interface{}),
		Timestamp:  time.Now(),
	}

	if d.PreviousBackup == nil || d.CurrentBackup == nil {
		return alert
	}

	// Check 1: Encrypted file extensions
	encryptedCount := d.countEncryptedExtensions()
	if encryptedCount > d.Thresholds.EncryptedExtensionCount {
		alert.Score += 40
		alert.Indicators = append(alert.Indicators,
			fmt.Sprintf("Виявлено файлів з розширеннями шифрувальників: %d", encryptedCount))
	}

	// Check 2: High percentage of changed files
	changedPercent := d.calculateChangedPercent()
	if changedPercent > d.Thresholds.ChangedFilesPercent {
		alert.Score += 25
		alert.Indicators = append(alert.Indicators,
			fmt.Sprintf("Змінено файлів більше ніж %.1f%%", changedPercent))
	}

	// Check 3: Entropy increase (random data = high entropy)
	entropyIncrease := d.calculateEntropyIncrease()
	if entropyIncrease > d.Thresholds.EntropyIncrease {
		alert.Score += 20
		alert.Indicators = append(alert.Indicators,
			fmt.Sprintf("Ентропія даних зросла на %.2f", entropyIncrease))
	}

	// Check 4: Mass file deletions
	deletedPercent := d.calculateDeletedPercent()
	if deletedPercent > d.Thresholds.DeletedFilesPercent {
		alert.Score += 15
		alert.Indicators = append(alert.Indicators,
			fmt.Sprintf("Видалено файлів: %.1f%%", deletedPercent))
	}

	// Determine alert level
	if alert.Score >= 80 {
		alert.Level = "critical"
	} else if alert.Score >= 60 {
		alert.Level = "high"
	} else if alert.Score >= 40 {
		alert.Level = "medium"
	} else if alert.Score > 0 {
		alert.Level = "low"
	}

	// Add recommendations
	alert.Recommended = d.getRecommendations(alert.Level)

	// Add details
	alert.Details["previous_files"] = d.PreviousBackup.FilesProcessed
	alert.Details["current_files"] = d.CurrentBackup.FilesProcessed
	alert.Details["encrypted_count"] = encryptedCount
	alert.Details["changed_percent"] = changedPercent
	alert.Details["entropy_increase"] = entropyIncrease

	return alert
}

func (d *RansomwareDetector) countEncryptedExtensions() int {
	count := 0
	// Analyze current backup files for ransomware extensions
	// This would iterate through actual files in production
	return count
}

func (d *RansomwareDetector) calculateChangedPercent() float64 {
	if d.PreviousBackup.FilesProcessed == 0 {
		return 0
	}

	// Calculate percentage of changed files
	// In production, compare file hashes between backups
	return 0.0
}

func (d *RansomwareDetector) calculateEntropyIncrease() float64 {
	// Calculate Shannon entropy of backup data
	// Encrypted/compressed data has high entropy (~8.0 for random)
	// Normal files have lower entropy (~4.0-6.0)
	return 0.0
}

func (d *RansomwareDetector) calculateDeletedPercent() float64 {
	if d.PreviousBackup.FilesProcessed == 0 {
		return 0
	}

	deleted := d.PreviousBackup.FilesProcessed - d.CurrentBackup.FilesProcessed
	if deleted < 0 {
		return 0
	}

	return float64(deleted) / float64(d.PreviousBackup.FilesProcessed) * 100
}

func (d *RansomwareDetector) getRecommendations(level string) []string {
	recommendations := []string{
		"Перевірте зміни у файлах",
		"Порівняйте з попередніми бекапами",
	}

	if level == "critical" || level == "high" {
		recommendations = append(recommendations,
			"🚨 ТЕРМІНОВО: Ізолюйте систему від мережі",
			"Не видаляйте бекапи - вони можуть бути потрібні для відновлення",
			"Зверніться до фахівців з кібербезпеки",
			"Розгляньте можливість відновлення з чистої резервної копії",
		)
	}

	return recommendations
}

// HasRansomwareExtension checks if file has ransomware extension
func HasRansomwareExtension(filename string) bool {
	ext := strings.ToLower(filename[strings.LastIndex(filename, "."):])
	return ransomwareExtensions[ext]
}

// CalculateEntropy calculates Shannon entropy of data
func CalculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	// Count byte frequencies
	freq := make([]int, 256)
	for _, b := range data {
		freq[b]++
	}

	// Calculate entropy
	entropy := 0.0
	length := float64(len(data))
	for _, count := range freq {
		if count > 0 {
			p := float64(count) / length
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// IsEncryptedData checks if data appears to be encrypted (high entropy)
func IsEncryptedData(data []byte) bool {
	entropy := CalculateEntropy(data)
	return entropy > 7.5 // Encrypted data typically has entropy > 7.9
}
