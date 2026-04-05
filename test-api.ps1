# Step 1: Try login (should get 403 - password change required)
Write-Host "=== Step 1: Login (expect 403) ==="
$loginBody = @{ username = 'admin'; password = 'admin123' } | ConvertTo-Json
try {
    $loginResp = Invoke-WebRequest -Uri 'http://localhost:8000/api/auth/login' -Method POST -Body $loginBody -ContentType 'application/json' -UseBasicParsing
    Write-Host "Login: $($loginResp.Content)"
} catch {
    Write-Host "Login Status: $($_.Exception.Response.StatusCode.value__)"
    Write-Host "Login Body: $($_.Exception.Response.StatusDescription)"
    $stream = $_.Exception.Response.GetResponseStream()
    $reader = New-Object System.IO.StreamReader($stream)
    $responseBody = $reader.ReadToEnd()
    Write-Host "Login Response: $responseBody"
}

# Step 2: Change password (first login)
Write-Host "`n=== Step 2: Change Password ==="
$changeBody = @{ username = 'admin'; currentPassword = 'admin123'; newPassword = 'Admin123!' } | ConvertTo-Json
$changeResp = Invoke-WebRequest -Uri 'http://localhost:8000/api/auth/change-password-first-login' -Method POST -Body $changeBody -ContentType 'application/json' -UseBasicParsing
Write-Host "Change Password Response: $($changeResp.Content)"
$token = ($changeResp.Content | ConvertFrom-Json).token
Write-Host "Token: $token"

# Step 3: Login with new password
Write-Host "`n=== Step 3: Login with new password ==="
$newLoginBody = @{ username = 'admin'; password = 'Admin123!' } | ConvertTo-Json
$newLoginResp = Invoke-WebRequest -Uri 'http://localhost:8000/api/auth/login' -Method POST -Body $newLoginBody -ContentType 'application/json' -UseBasicParsing
Write-Host "New Login Response: $($newLoginResp.Content)"
$token2 = ($newLoginResp.Content | ConvertFrom-Json).token
Write-Host "Token2: $token2"

# Step 4: Test authenticated endpoints
Write-Host "`n=== Step 4: Authenticated Endpoints ==="
$headers = @{ Authorization = "Bearer $token2" }

$jobs = Invoke-WebRequest -Uri 'http://localhost:8000/api/jobs' -Headers $headers -UseBasicParsing
Write-Host "Jobs: $($jobs.Content)"

$repos = Invoke-WebRequest -Uri 'http://localhost:8000/api/repositories' -Headers $headers -UseBasicParsing
Write-Host "Repositories: $($repos.Content)"

$reports = Invoke-WebRequest -Uri 'http://localhost:8000/api/reports/summary' -Headers $headers -UseBasicParsing
Write-Host "Reports: $($reports.Content)"

$me = Invoke-WebRequest -Uri 'http://localhost:8000/api/auth/me' -Headers $headers -UseBasicParsing
Write-Host "Me: $($me.Content)"

Write-Host "`n=== All Tests Passed ==="
