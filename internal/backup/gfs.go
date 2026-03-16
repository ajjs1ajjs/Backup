// GFS Retention Policy - Grandfather-Father-Son
package backup

import "time"

// GFSRetention represents GFS retention policy
type GFSRetention struct {
	Daily     int `json:"daily"`     // Keep daily backups for N days
	Weekly    int `json:"weekly"`    // Keep weekly backups for N weeks
	Monthly   int `json:"monthly"`   // Keep monthly backups for N months
	Quarterly int `json:"quarterly"` // Keep quarterly backups for N quarters
	Yearly    int `json:"yearly"`    // Keep yearly backups for N years
}

// BackupCopy represents backup copy configuration
type BackupCopy struct {
	Enabled       bool   `json:"enabled"`
	DestinationID string `json:"destination_id"`
	DelayHours    int    `json:"delay_hours"`    // Delay before copying (0-168)
	RetentionDays int    `json:"retention_days"` // Retention for copies
	Encryption    bool   `json:"encryption"`     // Encrypt copy
}

// ApplyGFSRetention applies GFS policy to backup retention
func ApplyGFSRetention(backupTime time.Time, gfs GFSRetention) map[string]bool {
	retention := make(map[string]bool)
	now := time.Now()

	// Daily - keep backups from last N days
	for i := 0; i < gfs.Daily; i++ {
		day := now.AddDate(0, 0, -i)
		if isSameDay(backupTime, day) {
			retention["daily"] = true
		}
	}

	// Weekly - keep backups from last N weeks (Sunday)
	for i := 0; i < gfs.Weekly; i++ {
		week := now.AddDate(0, 0, -i*7)
		if isSunday(backupTime) && isSameWeek(backupTime, week) {
			retention["weekly"] = true
		}
	}

	// Monthly - keep backups from 1st of month
	for i := 0; i < gfs.Monthly; i++ {
		month := now.AddDate(0, -i, 0)
		if backupTime.Day() == 1 && backupTime.Month() == month.Month() && backupTime.Year() == month.Year() {
			retention["monthly"] = true
		}
	}

	// Quarterly - keep backups from 1st of Jan, Apr, Jul, Oct
	quarters := []time.Month{time.January, time.April, time.July, time.October}
	for _, qMonth := range quarters {
		if backupTime.Month() == qMonth && backupTime.Day() == 1 {
			retention["quarterly"] = true
		}
	}

	// Yearly - keep backups from Jan 1st
	if backupTime.Month() == time.January && backupTime.Day() == 1 {
		retention["yearly"] = true
	}

	return retention
}

func isSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

func isSunday(t time.Time) bool {
	return t.Weekday() == time.Sunday
}

func isSameWeek(t1, t2 time.Time) bool {
	_, w1 := t1.ISOWeek()
	_, w2 := t2.ISOWeek()
	return t1.Year() == t2.Year() && w1 == w2
}

// ShouldRetainBackup determines if backup should be retained based on GFS
func ShouldRetainBackup(backupTime time.Time, gfs GFSRetention) bool {
	retention := ApplyGFSRetention(backupTime, gfs)
	return len(retention) > 0
}

// GetRetentionType returns the type of retention for a backup
func GetRetentionType(backupTime time.Time, gfs GFSRetention) string {
	retention := ApplyGFSRetention(backupTime, gfs)

	if retention["yearly"] {
		return "yearly"
	}
	if retention["quarterly"] {
		return "quarterly"
	}
	if retention["monthly"] {
		return "monthly"
	}
	if retention["weekly"] {
		return "weekly"
	}
	if retention["daily"] {
		return "daily"
	}

	return "none"
}
