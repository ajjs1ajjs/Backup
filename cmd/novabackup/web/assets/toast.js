/**
 * NovaBackup Enterprise - Toast Notifications
 * Replaces alert() with beautiful toast notifications
 */

// Toast container
let toastContainer = null;

// Initialize toast container
function initToastContainer() {
    if (!toastContainer) {
        toastContainer = document.createElement('div');
        toastContainer.className = 'toast-container';
        document.body.appendChild(toastContainer);
    }
}

// Show toast notification
function showToast(message, type = 'info', duration = 5000) {
    initToastContainer();

    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;

    const icons = {
        success: '✅',
        warning: '⚠️',
        error: '❌',
        info: 'ℹ️'
    };

    const titles = {
        success: 'Успішно!',
        warning: 'Попередження',
        error: 'Помилка',
        info: 'Інформація'
    };

    toast.innerHTML = `
        <span class="toast-icon">${icons[type]}</span>
        <div class="toast-content">
            <div class="toast-title">${titles[type]}</div>
            <div class="toast-message">${escapeHtml(message)}</div>
        </div>
        <button class="toast-close" onclick="closeToast(this.parentElement)">&times;</button>
        <div class="toast-progress">
            <div class="toast-progress-bar" style="animation-duration: ${duration}ms"></div>
        </div>
    `;

    toastContainer.appendChild(toast);

    // Auto-close after duration
    if (duration > 0) {
        setTimeout(() => closeToast(toast), duration);
    }

    return toast;
}

// Close toast
function closeToast(toast) {
    if (!toast) return;

    toast.classList.add('hiding');
    setTimeout(() => {
        if (toast.parentElement) {
            toast.parentElement.removeChild(toast);
        }
    }, 300);
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Convenience methods
function showSuccess(message, duration) {
    return showToast(message, 'success', duration);
}

function showError(message, duration) {
    return showToast(message, 'error', duration);
}

function showWarning(message, duration) {
    return showToast(message, 'warning', duration);
}

function showInfo(message, duration) {
    return showToast(message, 'info', duration);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', initToastContainer);
