// Universal auth headers helper
// Include this script in all HTML pages that need authentication
function getAuthHeaders() {
    const token = localStorage.getItem("authToken");
    const headers = { "Content-Type": "application/json" };
    if (token) {
        headers.Authorization = "Bearer " + token;
    }
    return headers;
}

// Check if user is authenticated
function checkAuth() {
    const token = localStorage.getItem("authToken");
    if (!token) {
        window.location.href = "login.html";
        return false;
    }
    return true;
}

// Logout function
function logout() {
    localStorage.removeItem("authToken");
    localStorage.removeItem("user");
    window.location.href = "login.html";
}
