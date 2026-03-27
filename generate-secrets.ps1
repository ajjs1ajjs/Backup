#!/usr/bin/env pwsh
<#
.SYNOPSIS
    NovaBackup Secrets Generator
    
.DESCRIPTION
    Generates secure random keys and passwords for NovaBackup configuration.
    All generated values are suitable for production use.
    
.EXAMPLE
    .\generate-secrets.ps1
    .\generate-secrets.ps1 -All
    .\generate-secrets.ps1 -MasterKey
    .\generate-secrets.ps1 -JwtSecret
    .\generate-secrets.ps1 -ApiKey
    .\generate-secrets.ps1 -Password
    
.NOTES
    Generated secrets are cryptographically secure using System.Security.Cryptography
#>

[CmdletBinding()]
param(
    [switch]$All,
    [switch]$MasterKey,
    [switch]$JwtSecret,
    [switch]$ApiKey,
    [switch]$Password,
    [switch]$PostgresPassword,
    [switch]$UpdateEnv
)

# Generate cryptographically secure random string
function Get-SecureRandomString {
    param(
        [int]$Length = 32,
        [string]$CharacterSet = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    )
    
    $random = New-Object System.Security.Cryptography.RNGCryptoServiceProvider
    $bytes = New-Object byte[] $Length
    $random.GetBytes($bytes)
    
    $result = ""
    foreach ($byte in $bytes) {
        $result += $CharacterSet[$byte % $CharacterSet.Length]
    }
    
    return $result
}

# Generate hex string (for keys)
function Get-SecureHexKey {
    param([int]$Length = 32)
    $bytes = New-Object byte[] ($Length / 2)
    $random = New-Object System.Security.Cryptography.RNGCryptoServiceProvider
    $random.GetBytes($bytes)
    return [System.BitConverter]::ToString($bytes).Replace('-', '').ToLower()
}

# Generate master key (32 bytes = 64 hex chars)
function New-MasterKey {
    param([int]$Length = 64)
    Write-Host "Master Key ($Length chars):" -ForegroundColor Cyan
    $key = Get-SecureHexKey -Length $Length
    Write-Host $key -ForegroundColor Green
    Write-Host ""
    return $key
}

# Generate JWT secret (base64 encoded random bytes)
function New-JwtSecret {
    param([int]$Length = 64)
    Write-Host "JWT Secret ($Length chars):" -ForegroundColor Cyan
    $secret = Get-SecureRandomString -Length $Length
    Write-Host $secret -ForegroundColor Green
    Write-Host ""
    return $secret
}

# Generate API key
function New-ApiKey {
    param([int]$Length = 32)
    Write-Host "API Key ($Length chars):" -ForegroundColor Cyan
    $key = Get-SecureRandomString -Length $Length
    Write-Host $key -ForegroundColor Green
    Write-Host ""
    return $key
}

# Generate secure password
function New-SecurePassword {
    param(
        [int]$Length = 20,
        [switch]$IncludeSpecial
    )
    
    $chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
    if ($IncludeSpecial) {
        $chars += '!@#$%^&*()_+-=[]{}|;:,.<>?'
    }
    
    Write-Host "Secure Password ($Length chars):" -ForegroundColor Cyan
    $password = Get-SecureRandomString -Length $Length -CharacterSet $chars
    Write-Host $password -ForegroundColor Green
    Write-Host ""
    return $password
}

# Generate PostgreSQL password
function New-PostgresPassword {
    param([int]$Length = 24)
    Write-Host "PostgreSQL Password ($Length chars):" -ForegroundColor Cyan
    # PostgreSQL passwords shouldn't contain certain special chars
    $chars = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-'
    $password = Get-SecureRandomString -Length $Length -CharacterSet $chars
    Write-Host $password -ForegroundColor Green
    Write-Host ""
    return $password
}

# Update .env file with new secrets
function Update-EnvFile {
    param(
        [string]$MasterKey,
        [string]$JwtSecret,
        [string]$ApiKey,
        [string]$PostgresPassword
    )
    
    $envFile = Join-Path $PSScriptRoot '.env'
    $envExample = Join-Path $PSScriptRoot '.env.example'
    
    if (-not (Test-Path $envExample)) {
        Write-Host "Error: .env.example not found" -ForegroundColor Red
        return $false
    }
    
    # Copy example to .env if it doesn't exist
    if (-not (Test-Path $envFile)) {
        Copy-Item $envExample $envFile
        Write-Host "Created .env from template" -ForegroundColor Green
    }
    
    # Read current content
    $content = Get-Content $envFile -Raw
    
    # Replace secrets
    if ($MasterKey) {
        $content = $content -replace '(NOVABACKUP_MASTER_KEY=).*', "`$1$MasterKey"
    }
    if ($JwtSecret) {
        $content = $content -replace '(NOVABACKUP_JWT_SECRET=).*', "`$1$JwtSecret"
    }
    if ($ApiKey) {
        $content = $content -replace '(NOVABACKUP_API_KEY=).*', "`$1$ApiKey"
    }
    if ($PostgresPassword) {
        $content = $content -replace '(POSTGRES_PASSWORD=).*', "`$1$PostgresPassword"
    }
    
    # Write back
    Set-Content $envFile -Value $content -NoNewline
    Write-Host "Updated .env file with new secrets" -ForegroundColor Green
    Write-Host "Location: $envFile" -ForegroundColor Gray
    Write-Host ""
    
    return $true
}

# Main logic
Write-Host "=== NovaBackup Secrets Generator ===" -ForegroundColor Cyan
Write-Host ""

$generatedSomething = $false

if ($All -or $MasterKey) {
    $masterKey = New-MasterKey -NoNewline
    $generatedSomething = $true
}

if ($All -or $JwtSecret) {
    $jwtSecret = New-JwtSecret -NoNewline
    $generatedSomething = $true
}

if ($All -or $ApiKey) {
    $apiKey = New-ApiKey -NoNewline
    $generatedSomething = $true
}

if ($All -or $Password) {
    $password = New-SecurePassword -Length 20 -IncludeSpecial -NoNewline
    $generatedSomething = $true
}

if ($All -or $PostgresPassword) {
    $postgresPassword = New-PostgresPassword -NoNewline
    $generatedSomething = $true
}

if ($UpdateEnv) {
    Update-EnvFile `
        -MasterKey $masterKey `
        -JwtSecret $jwtSecret `
        -ApiKey $apiKey `
        -PostgresPassword $postgresPassword
}

if (-not $generatedSomething) {
    Write-Host "Usage:" -ForegroundColor Yellow
    Write-Host "  .\generate-secrets.ps1 -All              # Generate all secrets" -ForegroundColor Gray
    Write-Host "  .\generate-secrets.ps1 -MasterKey        # Generate master encryption key" -ForegroundColor Gray
    Write-Host "  .\generate-secrets.ps1 -JwtSecret        # Generate JWT secret" -ForegroundColor Gray
    Write-Host "  .\generate-secrets.ps1 -ApiKey           # Generate API key" -ForegroundColor Gray
    Write-Host "  .\generate-secrets.ps1 -Password         # Generate secure password" -ForegroundColor Gray
    Write-Host "  .\generate-secrets.ps1 -PostgresPassword # Generate PostgreSQL password" -ForegroundColor Gray
    Write-Host "  .\generate-secrets.ps1 -UpdateEnv        # Update .env file with new secrets" -ForegroundColor Gray
    Write-Host ""
}

Write-Host "⚠️  IMPORTANT: Store these secrets securely!" -ForegroundColor Yellow
Write-Host "   Never commit .env file to version control" -ForegroundColor Gray
Write-Host ""
