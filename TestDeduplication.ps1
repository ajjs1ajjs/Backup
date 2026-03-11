# Test NovaBackup Deduplication Engine

Write-Host "=== TESTING NOVABACKUP DEDUPLICATION ENGINE ===" -ForegroundColor Cyan
Write-Host ""

# Create test data with duplicates
$testData = @(
    @{ "Id" = 1; "Data" = "AAAAAA"; "Size" = 6 },
    @{ "Id" = 2; "Data" = "BBBBBB"; "Size" = 6 },
    @{ "Id" = 3; "Data" = "CCCCCC"; "Size" = 6 },
    @{ "Id" = 4; "Data" = "AAAAAA"; "Size" = 6 },  # Duplicate
    @{ "Id" = 5; "Data" = "DDDDDD"; "Size" = 6 },
    @{ "Id" = 6; "Data" = "BBBBBB"; "Size" = 6 },  # Duplicate
    @{ "Id" = 7; "Data" = "EEEEEE"; "Size" = 6 },
    @{ "Id" = 8; "Data" = "FFFFFF"; "Size" = 6 }
)

Write-Host "Test Data: 8 blocks (with duplicates)" -ForegroundColor Gray
Write-Host ""

# Initialize deduplication engine
$globalIndex = @{}
$uniqueBlocks = @{}
$dedupRatio = 1.0

# Stage 1: Hash Calculation
Write-Host "1. Hash Calculation (SHA256)" -ForegroundColor Yellow
$hashedBlocks = @()

foreach ($block in $testData) {
    $hash = [System.BitConverter]::ToString(
        [System.Security.Cryptography.SHA256]::Create().ComputeHash(
            [System.Text.Encoding]::UTF8.GetBytes($block.Data)
        )
    ).Replace("-", "").Substring(0, 16)
    
    $hashedBlocks += @{
        "Id" = $block.Id
        "Data" = $block.Data
        "Size" = $block.Size
        "Hash" = $hash
    }
}

Write-Host "   Calculated SHA256 hashes for all blocks" -ForegroundColor Gray

# Stage 2: Deduplication
Write-Host "2. Deduplication Process" -ForegroundColor Yellow
$finalBlocks = @()
$duplicateCount = 0

foreach ($block in $hashedBlocks) {
    if ($globalIndex.ContainsKey($block.Hash)) {
        # Duplicate found
        $finalBlocks += @{
            "Id" = $block.Id
            "Reference" = $globalIndex[$block.Hash].Id
            "Hash" = $block.Hash
            "Duplicate" = $true
        }
        $duplicateCount++
        Write-Host "   Duplicate found: Block $($block.Id) → Reference to Block $($globalIndex[$block.Hash].Id)" -ForegroundColor Red
    } else {
        # Unique block
        $globalIndex[$block.Hash] = $block
        $uniqueBlocks[$block.Hash] = $block
        $finalBlocks += @{
            "Id" = $block.Id
            "Data" = $block.Data
            "Hash" = $block.Hash
            "Duplicate" = $false
        }
        Write-Host "   Unique block: Block $($block.Id) → Hash $($block.Hash)" -ForegroundColor Green
    }
}

# Calculate results
$originalSize = ($testData | Measure-Object -Property Size).Sum
$uniqueSize = ($uniqueBlocks.Values | Measure-Object -Property Size).Sum
$dedupRatio = [math]::Round($originalSize / $uniqueSize, 2)

Write-Host ""
Write-Host "=== DEDUPLICATION RESULTS ===" -ForegroundColor Green
Write-Host "Original Blocks: $($testData.Count)" -ForegroundColor White
Write-Host "Unique Blocks: $($uniqueBlocks.Count)" -ForegroundColor White
Write-Host "Duplicate Blocks: $duplicateCount" -ForegroundColor White
Write-Host "Original Size: $originalSize bytes" -ForegroundColor White
Write-Host "Dedup Size: $uniqueSize bytes" -ForegroundColor White
Write-Host "Deduplication Ratio: $dedupRatio:1" -ForegroundColor Green
Write-Host "Storage Savings: $([math]::Round((1 - 1/$dedupRatio) * 100, 1))%" -ForegroundColor Green

Write-Host ""
Write-Host "=== ARCHITECTURE DIAGRAM ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "                Web UI" -ForegroundColor Yellow
Write-Host "                   │" -ForegroundColor Gray
Write-Host "                API Server" -ForegroundColor Yellow
Write-Host "                   │" -ForegroundColor Gray
Write-Host "            Job Scheduler" -ForegroundColor Yellow
Write-Host "                   │" -ForegroundColor Gray
Write-Host "            Backup Orchestrator" -ForegroundColor Yellow
Write-Host "                   │" -ForegroundColor Gray
Write-Host "        ┌──────────┼──────────┐" -ForegroundColor Gray
Write-Host "        │          │          │" -ForegroundColor Gray
Write-Host "   Snapshot    Data Reader   CBT" -ForegroundColor Yellow
Write-Host "        │" -ForegroundColor Gray
Write-Host "   Data Pipeline" -ForegroundColor Yellow
Write-Host "        │" -ForegroundColor Gray
Write-Host "Compression → Dedup → Encryption" -ForegroundColor Green
Write-Host "        │" -ForegroundColor Gray
Write-Host "   Storage Engine" -ForegroundColor Yellow
Write-Host "        │" -ForegroundColor Gray
Write-Host "Repository / Cloud / Tape" -ForegroundColor Green
Write-Host ""
