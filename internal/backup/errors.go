// Enhanced Error Handling - Detailed error messages with solutions
package backup

import (
	"fmt"
	"os"
	"strings"
)

// ErrorCode represents specific error codes
type ErrorCode string

const (
	ErrDiskFull           ErrorCode = "DISK_FULL"
	ErrPermissionDenied   ErrorCode = "PERMISSION_DENIED"
	ErrSourceNotFound     ErrorCode = "SOURCE_NOT_FOUND"
	ErrNetworkError       ErrorCode = "NETWORK_ERROR"
	ErrEncryptionFailed   ErrorCode = "ENCRYPTION_FAILED"
	ErrCompressionFailed  ErrorCode = "COMPRESSION_FAILED"
	ErrDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrTimeout            ErrorCode = "TIMEOUT"
	ErrInvalidConfig      ErrorCode = "INVALID_CONFIG"
	ErrRansomwareDetected ErrorCode = "RANSOMWARE_DETECTED"
)

// BackupError represents a detailed backup error
type BackupError struct {
	Code        ErrorCode         `json:"code"`
	Message     string            `json:"message"`
	Details     map[string]string `json:"details,omitempty"`
	Solution    string            `json:"solution"`
	Severity    string            `json:"severity"` // critical, warning, info
	Recoverable bool              `json:"recoverable"`
}

func (e *BackupError) Error() string {
	return fmt.Sprintf("[%s] %s - %s", e.Code, e.Message, e.Solution)
}

// NewDiskFullError creates a disk full error
func NewDiskFullError(path string, required, available int64) *BackupError {
	return &BackupError{
		Code:    ErrDiskFull,
		Message: fmt.Sprintf("Недостатньо місця на диску %s", path),
		Details: map[string]string{
			"required":  FormatBytes(required),
			"available": FormatBytes(available),
			"path":      path,
		},
		Solution:    "Звільніть місце на диску або оберіть інше сховище з достатнім вільним місцем",
		Severity:    "critical",
		Recoverable: true,
	}
}

// NewPermissionError creates a permission denied error
func NewPermissionError(path string, err error) *BackupError {
	return &BackupError{
		Code:    ErrPermissionDenied,
		Message: fmt.Sprintf("Відмовлено в доступі до %s", path),
		Details: map[string]string{
			"path":  path,
			"error": err.Error(),
		},
		Solution:    "Перевірте права доступу до файлу/папки. Запустіть від імені адміністратора або надайте відповідні права",
		Severity:    "critical",
		Recoverable: true,
	}
}

// NewSourceNotFoundError creates a source not found error
func NewSourceNotFoundError(path string) *BackupError {
	return &BackupError{
		Code:    ErrSourceNotFound,
		Message: fmt.Sprintf("Джерело не знайдено: %s", path),
		Details: map[string]string{
			"path": path,
		},
		Solution:    "Перевірте чи існує вказаний шлях та чи є він доступним",
		Severity:    "critical",
		Recoverable: false,
	}
}

// NewNetworkError creates a network error
func NewNetworkError(host string, port int, err error) *BackupError {
	return &BackupError{
		Code:    ErrNetworkError,
		Message: fmt.Sprintf("Помилка мережі при підключенні до %s:%d", host, port),
		Details: map[string]string{
			"host":  host,
			"port":  fmt.Sprintf("%d", port),
			"error": err.Error(),
		},
		Solution:    "Перевірте мережеве підключення, чи доступний сервер, та чи відкритий необхідний порт",
		Severity:    "critical",
		Recoverable: true,
	}
}

// WrapError wraps a standard error with additional context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Check for specific system errors
	if sysErr, ok := err.(*os.PathError); ok {
		// Check error string for common patterns (cross-platform)
		errStr := sysErr.Err.Error()
		if containsAny(errStr, []string{"disk full", "no space", "not enough space"}) {
			return NewDiskFullError(sysErr.Path, 0, 0)
		}
		if containsAny(errStr, []string{"access denied", "permission denied"}) {
			return NewPermissionError(sysErr.Path, err)
		}
		if containsAny(errStr, []string{"file not found", "no such file"}) {
			return NewSourceNotFoundError(sysErr.Path)
		}
	}

	// Generic error with context
	return &BackupError{
		Code:        "GENERAL_ERROR",
		Message:     fmt.Sprintf("%s: %v", context, err),
		Solution:    "Перевірте логи для отримання додаткової інформації",
		Severity:    "warning",
		Recoverable: true,
	}
}

func containsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(strings.ToLower(s), sub) {
			return true
		}
	}
	return false
}

// IsRecoverable checks if error is recoverable
func IsRecoverable(err error) bool {
	if be, ok := err.(*BackupError); ok {
		return be.Recoverable
	}
	return false
}

// GetSeverity returns error severity
func GetSeverity(err error) string {
	if be, ok := err.(*BackupError); ok {
		return be.Severity
	}
	return "warning"
}

// FormatBytes formats bytes to human readable format
func FormatBytes(bytes int64) string {
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

// ErrorToJSON converts error to JSON representation
func ErrorToJSON(err error) map[string]interface{} {
	if be, ok := err.(*BackupError); ok {
		return map[string]interface{}{
			"code":        be.Code,
			"message":     be.Message,
			"details":     be.Details,
			"solution":    be.Solution,
			"severity":    be.Severity,
			"recoverable": be.Recoverable,
		}
	}
	return map[string]interface{}{
		"message": err.Error(),
	}
}
