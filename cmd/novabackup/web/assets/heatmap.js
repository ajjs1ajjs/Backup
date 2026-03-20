/**
 * NovaBackup Enterprise - Heatmap Calendar
 * GitHub-style contribution graph for backup sessions
 */

// Generate heatmap data from sessions
function generateHeatmapData(sessions) {
    const last365Days = [];
    const today = new Date();

    // Initialize last 365 days
    for (let i = 364; i >= 0; i--) {
        const date = new Date(today);
        date.setDate(date.getDate() - i);
        last365Days.push({
            date: date.toISOString().split('T')[0],
            count: 0,
            sessions: []
        });
    }

    // Count sessions per day
    sessions.forEach(session => {
        const date = new Date(session.start_time).toISOString().split('T')[0];
        const dayData = last365Days.find(d => d.date === date);
        if (dayData) {
            dayData.count++;
            dayData.sessions.push(session);
        }
    });

    return last365Days;
}

// Get color level based on count
function getLevel(count) {
    if (count === 0) return 0;
    if (count === 1) return 1;
    if (count <= 3) return 2;
    if (count <= 5) return 3;
    return 4;
}

// Render heatmap
function renderHeatmap(containerId, sessions) {
    const container = document.getElementById(containerId);
    if (!container) return;

    const heatmapData = generateHeatmapData(sessions);
    const months = ['Січ', 'Лют', 'Бер', 'Кві', 'Тра', 'Чер', 'Лип', 'Сер', 'Вер', 'Жов', 'Лис', 'Гру'];

    // Group data by weeks
    const weeks = [];
    let currentWeek = [];

    // Add padding for first week (start from Sunday)
    const firstDate = new Date(heatmapData[0].date);
    const startDay = firstDate.getDay();
    for (let i = 0; i < startDay; i++) {
        currentWeek.push(null);
    }

    heatmapData.forEach((day, index) => {
        currentWeek.push(day);
        if (currentWeek.length === 7 || index === heatmapData.length - 1) {
            weeks.push(currentWeek);
            currentWeek = [];
        }
    });

    // Calculate total sessions
    const totalSessions = heatmapData.reduce((sum, day) => sum + day.count, 0);
    const maxStreak = calculateMaxStreak(heatmapData);
    const currentStreak = calculateCurrentStreak(heatmapData);

    // Build HTML
    let html = `
        <div class="heatmap-container">
            <div class="heatmap-header">
                <div>
                    <h3 class="heatmap-title">📅 Активність бекапів (365 днів)</h3>
                    <p style="font-size: 0.875rem; color: var(--text-secondary); margin-top: 0.25rem;">
                        Всього: <strong style="color: var(--text-primary)">${totalSessions}</strong> сесій |
                        Поточна серія: <strong style="color: var(--accent-green)">${currentStreak}</strong> днів |
                        Макс серія: <strong style="color: var(--accent-blue)">${maxStreak}</strong> днів
                    </p>
                </div>
                <div class="heatmap-legend">
                    <span>Менше</span>
                    <div class="heatmap-legend-colors">
                        <div class="heatmap-legend-item" style="background: var(--bg-secondary)"></div>
                        <div class="heatmap-legend-item" style="background: rgba(16, 185, 129, 0.3)"></div>
                        <div class="heatmap-legend-item" style="background: rgba(16, 185, 129, 0.5)"></div>
                        <div class="heatmap-legend-item" style="background: rgba(16, 185, 129, 0.7)"></div>
                        <div class="heatmap-legend-item" style="background: rgba(16, 185, 129, 0.9)"></div>
                    </div>
                    <span>Більше</span>
                </div>
            </div>
            <div class="heatmap-grid">
    `;

    // Render month labels
    let lastMonth = -1;
    weeks.forEach((week, weekIndex) => {
        const firstDay = week.find(d => d !== null);
        if (firstDay) {
            const month = new Date(firstDay.date).getMonth();
            if (month !== lastMonth) {
                html += `<div class="heatmap-month-label" style="grid-column: ${weekIndex + 1}; font-size: 0.625rem; color: var(--text-secondary); text-align: center;">${months[month]}</div>`;
                lastMonth = month;
            }
        }
    });

    // Render days
    weeks.forEach((week, weekIndex) => {
        week.forEach((day, dayIndex) => {
            if (day === null) {
                html += `<div class="heatmap-cell" style="visibility: hidden"></div>`;
            } else {
                const level = getLevel(day.count);
                const dateObj = new Date(day.date);
                const dateStr = dateObj.toLocaleDateString('uk-UA', {
                    day: '2-digit',
                    month: '2-digit',
                    year: 'numeric'
                });

                html += `
                    <div class="heatmap-cell level-${level}"
                         data-date="${day.date}"
                         data-count="${day.count}"
                         title="${dateStr}: ${day.count} сесій">
                        <div class="heatmap-tooltip">
                            <strong>${dateStr}</strong><br/>
                            ${day.count} сесій${day.sessions.length > 0 ? '<br/>' + day.sessions.map(s => `• ${s.job_name || 'Manual'}`).join('') : ''}
                        </div>
                    </div>
                `;
            }
        });
    });

    html += `
            </div>
        </div>
    `;

    container.innerHTML = html;
}

// Calculate max streak
function calculateMaxStreak(heatmapData) {
    let maxStreak = 0;
    let currentStreak = 0;

    heatmapData.forEach(day => {
        if (day.count > 0) {
            currentStreak++;
            maxStreak = Math.max(maxStreak, currentStreak);
        } else {
            currentStreak = 0;
        }
    });

    return maxStreak;
}

// Calculate current streak
function calculateCurrentStreak(heatmapData) {
    let streak = 0;

    // Count backwards from today
    for (let i = heatmapData.length - 1; i >= 0; i--) {
        if (heatmapData[i].count > 0) {
            streak++;
        } else {
            break;
        }
    }

    return streak;
}

// Auto-update heatmap when sessions load
function initHeatmap() {
    // Will be called from index.html after loading sessions
    if (typeof loadDashboard === 'function') {
        const originalLoadDashboard = loadDashboard;
        window.loadDashboard = async function() {
            await originalLoadDashboard();

            // Find sessions and render heatmap
            const sessionsRes = await fetch("/api/backup/sessions", {
                headers: { Authorization: localStorage.getItem("authToken") }
            });
            const sessions = await sessionsRes.json();

            if (sessions.sessions) {
                renderHeatmap('heatmap-container', sessions.sessions);
            }
        };
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', initHeatmap);
