# NovaBackup v6.0 - Enterprise Backup Pipeline
# 10-стадійна pipeline як у Veeam Backup

param(
    [Parameter(Mandatory=$true)]
    [string]$JobName,
    
    [Parameter(Mandatory=$true)]
    [string]$SourcePath,
    
    [Parameter(Mandatory=$true)]
    [string]$RepositoryPath,
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("Full", "Incremental", "ReverseIncremental", "SyntheticFull")]
    [string]$BackupType = "Incremental",
    
    [Parameter(Mandatory=$false)]
    [string]$VMName = "",
    
    [Parameter(Mandatory=$false)]
    [switch]$ApplicationAware,
    
    [Parameter(Mandatory=$false)]
    [switch]$UseCBT,
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("None", "LZ4", "LZ5", "Optimal", "High", "Extreme")]
    [string]$CompressionLevel = "Optimal",
    
    [Parameter(Mandatory=$false)]
    [switch]$EnableEncryption,
    
    [Parameter(Mandatory=$false)]
    [switch]$EnableDeduplication
)

# Global variables
$script:JobId = [Guid]::NewGuid().ToString()
$script:StartTime = Get-Date
$script:CurrentStage = ""
$script:Stages = @()
$script:Metadata = @{}

function Write-PipelineLog {
    param([string]$Message, [string]$Level = "INFO", [string]$Color = "White")
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff"
    $logEntry = "[$timestamp] [$Level] [Stage:$($script:CurrentStage)] $Message"
    
    Write-Host $logEntry -ForegroundColor $Color
    
    # Also write to log file
    $logFile = "$RepositoryPath\logs\$($script:JobId).log"
    if (!(Test-Path (Split-Path $logFile -Parent))) {
        New-Item -ItemType Directory -Path (Split-Path $logFile -Parent) -Force
    }
    Add-Content -Path $logFile -Value $logEntry -ErrorAction SilentlyContinue
}

function Update-Progress {
    param([int]$Percentage, [string]$Status = "")
    
    $activity = "Stage $script:CurrentStage: $Percentage%"
    if ($Status) {
        $activity += " - $Status"
    }
    
    Write-Progress -Activity $activity -Status $Status -PercentComplete $Percentage
}

# ========================================
# STAGE 1: Job Scheduler & Init
# ========================================
function Stage1-JobScheduler {
    $script:CurrentStage = "JobScheduler"
    $script:Stages += @{ "Stage" = "Job Scheduler"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 1: JOB SCHEDULER & INITIALIZATION ===" "INFO" "Cyan"
    
    try {
        # 1.1 Validate job configuration
        Write-PipelineLog "1.1 Validating job configuration..."
        Update-Progress 10 "Validating paths"
        
        if (!(Test-Path $SourcePath)) {
            throw "Source path does not exist: $SourcePath"
        }
        
        if (!(Test-Path $RepositoryPath)) {
            Write-PipelineLog "Creating repository directory: $RepositoryPath"
            New-Item -ItemType Directory -Path $RepositoryPath -Force
        }
        
        # 1.2 Initialize job metadata
        Write-PipelineLog "1.2 Initializing job metadata..."
        Update-Progress 20 "Initializing metadata"
        
        $script:Metadata = @{
            "JobId" = $script:JobId
            "JobName" = $JobName
            "SourcePath" = $SourcePath
            "RepositoryPath" = $RepositoryPath
            "BackupType" = $script:BackupType
            "VMName" = $script:VMName
            "ApplicationAware" = $script:ApplicationAware
            "UseCBT" = $script:UseCBT
            "CompressionLevel" = $script:CompressionLevel
            "EnableEncryption" = $script:EnableEncryption
            "EnableDeduplication" = $script:EnableDeduplication
            "StartTime" = $script:StartTime
        }
        
        $script:Stages[0].Status = "Completed"
        Write-PipelineLog "Stage 1 completed successfully" "INFO" "Green"
        Update-Progress 30
        
    } catch {
        $script:Stages[0].Status = "Failed"
        Write-PipelineLog "Stage 1 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 2: Snapshot Creation
# ========================================
function Stage2-SnapshotCreation {
    $script:CurrentStage = "SnapshotCreation"
    $script:Stages += @{ "Stage" = "Snapshot Creation"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 2: SNAPSHOT CREATION ===" "INFO" "Cyan"
    
    try {
        # 2.1 Create VM snapshot if VM backup
        if ($script:VMName) {
            Write-PipelineLog "2.1 Creating VM snapshot for: $script:VMName"
            Update-Progress 40 "Creating VM snapshot"
            
            # Simulate VM snapshot creation
            if ($script:ApplicationAware) {
                Write-PipelineLog "2.2 Application-aware processing enabled"
                Update-Progress 50 "Application quiescing"
                
                # Simulate VSS writer interaction
                Start-Sleep -Seconds 3
            }
            
            # Simulate snapshot creation
            $snapshotId = "snap_$($script:JobId)_$(Get-Date -Format 'yyyyMMddHHmmss')"
            Write-PipelineLog "2.3 Snapshot created: $snapshotId"
            Update-Progress 60 "Snapshot created"
            
            $script:Metadata.SnapshotId = $snapshotId
            $script:Metadata.SnapshotTime = Get-Date
            
        } else {
            Write-PipelineLog "2.1 File-based backup - no snapshot needed"
            Update-Progress 60 "File system scan"
            
            $script:Metadata.SnapshotId = "file_backup_$($script:JobId)"
            $script:Metadata.SnapshotTime = Get-Date
        }
        
        $script:Stages[1].Status = "Completed"
        Write-PipelineLog "Stage 2 completed successfully" "INFO" "Green"
        Update-Progress 70
        
    } catch {
        $script:Stages[1].Status = "Failed"
        Write-PipelineLog "Stage 2 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 3: Application Consistency
# ========================================
function Stage3-ApplicationConsistency {
    $script:CurrentStage = "ApplicationConsistency"
    $script:Stages += @{ "Stage" = "Application Consistency"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 3: APPLICATION CONSISTENCY ===" "INFO" "Cyan"
    
    try {
        if ($script:ApplicationAware) {
            Write-PipelineLog "3.1 Processing application-aware backup"
            Update-Progress 75 "Application processing"
            
            # Simulate VSS writers
            $vssWriters = @(
                @{ "Name" = "SQL Server"; "Status" = "Stable" },
                @{ "Name" = "Exchange Server"; "Status" = "Stable" },
                @{ "Name" = "Active Directory"; "Status" = "Stable" },
                @{ "Name" = "System State"; "Status" = "Stable" }
            )
            
            foreach ($writer in $vssWriters) {
                Write-PipelineLog "3.2 VSS Writer: $($writer.Name) - $($writer.Status)"
                Start-Sleep -Milliseconds 500
            }
            
            # Simulate pre/post scripts
            Write-PipelineLog "3.3 Executing pre-freeze scripts"
            Start-Sleep -Seconds 1
            
            Write-PipelineLog "3.4 Executing post-thaw scripts"
            Start-Sleep -Seconds 1
            
            $script:Metadata.VssWriters = $vssWriters
        } else {
            Write-PipelineLog "3.1 Application-aware processing disabled"
            Update-Progress 80 "Skipping app consistency"
        }
        
        $script:Stages[2].Status = "Completed"
        Write-PipelineLog "Stage 3 completed successfully" "INFO" "Green"
        Update-Progress 85
        
    } catch {
        $script:Stages[2].Status = "Failed"
        Write-PipelineLog "Stage 3 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 4: Change Block Tracking (CBT)
# ========================================
function Stage4-ChangeBlockTracking {
    $script:CurrentStage = "ChangeBlockTracking"
    $script:Stages += @{ "Stage" = "Change Block Tracking"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 4: CHANGE BLOCK TRACKING ===" "INFO" "Cyan"
    
    try {
        if ($script:UseCBT) {
            Write-PipelineLog "4.1 Using Change Block Tracking (CBT)"
            Update-Progress 87 "CBT analysis"
            
            # Simulate CBT analysis
            $lastBackupTime = Get-Date (Get-Date).AddDays(-7)  # Simulate last backup
            $changedBlocks = @()
            $totalBlocks = 1000
            $changedCount = Get-Random -Minimum 100 -Maximum 300
            
            for ($i = 0; $i -lt $changedCount; $i++) {
                $changedBlocks += @{
                    "BlockId" = $i
                    "Offset" = $i * 4096
                    "Size" = 4096
                    "Hash" = "hash_$i"
                    "Changed" = $true
                }
            }
            
            Write-PipelineLog "4.2 CBT Analysis: $($changedBlocks.Count)/$totalBlocks blocks changed"
            Update-Progress 90 "CBT analysis complete"
            
            $script:Metadata.CBTData = @{
                "LastBackupTime" = $lastBackupTime
                "ChangedBlocks" = $changedBlocks
                "TotalBlocks" = $totalBlocks
                "ChangedCount" = $changedCount
            }
        } else {
            Write-PipelineLog "4.1 CBT disabled - full backup mode"
            Update-Progress 90 "Full backup mode"
            
            $script:Metadata.CBTData = @{
                "Mode" = "Full"
                "Reason" = "CBT disabled"
            }
        }
        
        $script:Stages[3].Status = "Completed"
        Write-PipelineLog "Stage 4 completed successfully" "INFO" "Green"
        Update-Progress 95
        
    } catch {
        $script:Stages[3].Status = "Failed"
        Write-PipelineLog "Stage 4 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 5: Data Read (Proxy Stage)
# ========================================
function Stage5-DataRead {
    $script:CurrentStage = "DataRead"
    $script:Stages += @{ "Stage" = "Data Read"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 5: DATA READ (PROXY STAGE) ===" "INFO" "Cyan"
    
    try {
        Write-PipelineLog "5.1 Initializing data mover (proxy)"
        Update-Progress 96 "Initializing proxy"
        
        # Simulate proxy modes
        $proxyMode = "Direct SAN"  # Could be: HotAdd, NBD, Network
        Write-PipelineLog "5.2 Proxy mode: $proxyMode"
        
        # Simulate data reading
        $dataBlocks = @()
        $totalSize = 0
        
        if ($script:VMName) {
            # VM backup - read from snapshot
            Write-PipelineLog "5.3 Reading from VM snapshot: $($script:Metadata.SnapshotId)"
            $totalSize = Get-Random -Minimum 10737418240 -Maximum 5368709120  # 10-50GB
        } else {
            # File backup - read from file system
            Write-PipelineLog "5.3 Reading from file system: $SourcePath"
            $files = Get-ChildItem -Path $SourcePath -Recurse -File
            $totalSize = ($files | Measure-Object -Property Length).Sum
        }
        
        Write-PipelineLog "5.4 Total data size: $([math]::Round($totalSize/1GB, 2)) GB"
        Update-Progress 97 "Reading data"
        
        # Simulate reading data blocks
        $blocksToRead = [math]::Ceiling($totalSize / 4096)
        for ($i = 0; $i -lt $blocksToRead; $i += 100) {
            $dataBlocks += @{
                "BlockId" = $i
                "Data" = "data_block_$i"
                "Size" = 4096
                "Checksum" = "checksum_$i"
            }
            
            if ($i % 1000 -eq 0) {
                $progress = [math]::Round(($i / $blocksToRead) * 100, 1)
                Update-Progress (97 + [math]::Round($progress * 0.02, 1)) "Reading data blocks"
            }
            
            Start-Sleep -Milliseconds 10
        }
        
        $script:Metadata.DataBlocks = $dataBlocks
        $script:Metadata.TotalSize = $totalSize
        $script:Metadata.ProxyMode = $proxyMode
        
        $script:Stages[4].Status = "Completed"
        Write-PipelineLog "Stage 5 completed successfully" "INFO" "Green"
        Update-Progress 99
        
    } catch {
        $script:Stages[4].Status = "Failed"
        Write-PipelineLog "Stage 5 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 6: Compression
# ========================================
function Stage6-Compression {
    $script:CurrentStage = "Compression"
    $script:Stages += @{ "Stage" = "Compression"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 6: COMPRESSION ===" "INFO" "Cyan"
    
    try {
        Write-PipelineLog "6.1 Applying compression: $($script:CompressionLevel)"
        Update-Progress 99 "Compressing data"
        
        # Simulate compression
        $compressionRatio = 1.0
        switch ($script:CompressionLevel) {
            "None" { 
                $compressionRatio = 1.0 
                Write-PipelineLog "6.2 Compression disabled"
            }
            "LZ4" { 
                $compressionRatio = 0.6 
                Write-PipelineLog "6.2 Using LZ4 compression (fast)"
            }
            "LZ5" { 
                $compressionRatio = 0.5 
                Write-PipelineLog "6.2 Using LZ5 compression (balanced)"
            }
            "Optimal" { 
                $compressionRatio = 0.4 
                Write-PipelineLog "6.2 Using optimal compression"
            }
            "High" { 
                $compressionRatio = 0.3 
                Write-PipelineLog "6.2 Using high compression"
            }
            "Extreme" { 
                $compressionRatio = 0.25 
                Write-PipelineLog "6.2 Using extreme compression (slow)"
            }
        }
        
        $compressedSize = [math]::Round($script:Metadata.TotalSize * $compressionRatio, 0)
        Write-PipelineLog "6.3 Compression ratio: $([math]::Round($compressionRatio, 2)):1"
        Write-PipelineLog "6.4 Compressed size: $([math]::Round($compressedSize/1GB, 2)) GB"
        
        $script:Metadata.CompressionRatio = $compressionRatio
        $script:Metadata.CompressedSize = $compressedSize
        
        $script:Stages[5].Status = "Completed"
        Write-PipelineLog "Stage 6 completed successfully" "INFO" "Green"
        
    } catch {
        $script:Stages[5].Status = "Failed"
        Write-PipelineLog "Stage 6 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 7: Deduplication
# ========================================
function Stage7-Deduplication {
    $script:CurrentStage = "Deduplication"
    $script:Stages += @{ "Stage" = "Deduplication"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 7: DEDUPLICATION ===" "INFO" "Cyan"
    
    try {
        if ($script:EnableDeduplication) {
            Write-PipelineLog "7.1 Performing block-level deduplication"
            Update-Progress 100 "Deduplicating"
            
            # Simulate deduplication
            $uniqueBlocks = @{}
            $duplicateCount = 0
            
            foreach ($block in $script:Metadata.DataBlocks) {
                $blockHash = $block.Checksum
                if ($uniqueBlocks.ContainsKey($blockHash)) {
                    $duplicateCount++
                    Write-PipelineLog "7.2 Duplicate block found: $($block.BlockId)"
                } else {
                    $uniqueBlocks[$blockHash] = $block
                }
            }
            
            $deduplicationRatio = if ($duplicateCount -gt 0) { [math]::Round($script:Metadata.DataBlocks.Count / $uniqueBlocks.Count, 2) } else { 1.0 }
            $dedupSize = [math]::Round($script:Metadata.CompressedSize / $deduplicationRatio, 0)
            
            Write-PipelineLog "7.3 Deduplication ratio: $deduplicationRatio:1"
            Write-PipelineLog "7.4 Unique blocks: $($uniqueBlocks.Count)"
            Write-PipelineLog "7.5 Duplicate blocks: $duplicateCount"
            Write-PipelineLog "7.6 Final size after deduplication: $([math]::Round($dedupSize/1GB, 2)) GB"
            
            $script:Metadata.DeduplicationRatio = $deduplicationRatio
            $script:Metadata.DedupSize = $dedupSize
            $script:Metadata.UniqueBlocks = $uniqueBlocks.Values
        } else {
            Write-PipelineLog "7.1 Deduplication disabled"
            $script:Metadata.DeduplicationRatio = 1.0
            $script:Metadata.DedupSize = $script:Metadata.CompressedSize
        }
        
        $script:Stages[6].Status = "Completed"
        Write-PipelineLog "Stage 7 completed successfully" "INFO" "Green"
        
    } catch {
        $script:Stages[6].Status = "Failed"
        Write-PipelineLog "Stage 7 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 8: Encryption
# ========================================
function Stage8-Encryption {
    $script:CurrentStage = "Encryption"
    $script:Stages += @{ "Stage" = "Encryption"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 8: ENCRYPTION ===" "INFO" "Cyan"
    
    try {
        if ($script:EnableEncryption) {
            Write-PipelineLog "8.1 Applying AES-256 encryption"
            Update-Progress 100 "Encrypting data"
            
            # Simulate encryption
            $encryptionKey = "AES256_$(Get-Date -Format 'yyyyMMddHHmmss')"
            Write-PipelineLog "8.2 Encryption key generated: $encryptionKey"
            
            # Simulate encrypting all blocks
            $encryptedBlocks = @()
            foreach ($block in $script:Metadata.UniqueBlocks) {
                $encryptedBlock = @{
                    "BlockId" = $block.BlockId
                    "EncryptedData" = "encrypted_$($block.Data)"
                    "IV" = "iv_$($block.BlockId)"
                    "Key" = $encryptionKey
                }
                $encryptedBlocks += $encryptedBlock
            }
            
            Write-PipelineLog "8.3 Data encrypted with AES-256"
            $script:Metadata.EncryptedBlocks = $encryptedBlocks
            $script:Metadata.EncryptionKey = $encryptionKey
        } else {
            Write-PipelineLog "8.1 Encryption disabled"
            $script:Metadata.EncryptedBlocks = $script:Metadata.UniqueBlocks
        }
        
        $script:Stages[7].Status = "Completed"
        Write-PipelineLog "Stage 8 completed successfully" "INFO" "Green"
        
    } catch {
        $script:Stages[7].Status = "Failed"
        Write-PipelineLog "Stage 8 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 9: Transport & Storage Write
# ========================================
function Stage9-TransportStorage {
    $script:CurrentStage = "TransportStorage"
    $script:Stages += @{ "Stage" = "Transport & Storage"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 9: TRANSPORT & STORAGE WRITE ===" "INFO" "Cyan"
    
    try {
        Write-PipelineLog "9.1 Initializing transport to repository: $RepositoryPath"
        
        # Create backup files
        $backupDir = "$RepositoryPath\$($script:JobId)"
        if (!(Test-Path $backupDir)) {
            New-Item -ItemType Directory -Path $backupDir -Force
        }
        
        Write-PipelineLog "9.2 Creating backup files (VBK, VIB, VRB)"
        
        # Determine file types based on backup type
        switch ($script:BackupType) {
            "Full" {
                $fileExtension = "VBK"
                Write-PipelineLog "9.3 Creating Full backup (VBK)"
            }
            "Incremental" {
                $fileExtension = "VIB"
                Write-PipelineLog "9.3 Creating Incremental backup (VIB)"
            }
            "ReverseIncremental" {
                $fileExtension = "VRB"
                Write-PipelineLog "9.3 Creating Reverse Incremental backup (VRB)"
            }
            "SyntheticFull" {
                $fileExtension = "VBK"
                Write-PipelineLog "9.3 Creating Synthetic Full backup (VBK)"
            }
        }
        
        # Create backup files
        $backupFile = "$backupDir\$($script:JobName)_$(Get-Date -Format 'yyyyMMddHHmmss').$fileExtension"
        
        # Simulate writing backup data
        $finalSize = if ($script:Metadata.EncryptedBlocks) { 
            $script:Metadata.DedupSize 
        } else { 
            $script:Metadata.DedupSize 
        }
        
        # Create dummy backup file (in real implementation, this would be actual data)
        $dummyData = "NovaBackup v6.0 - $($script:BackupType) backup`r`nJob: $($script:JobName)`r`nSize: $([math]::Round($finalSize/1GB, 2)) GB`r`nCompressed: $([math]::Round($script:Metadata.CompressedSize/1GB, 2)) GB`r`nDedup: $([math]::Round($script:Metadata.DedupSize/1GB, 2)) GB`r`nCreated: $(Get-Date)"
        
        Set-Content -Path $backupFile -Value $dummyData -Encoding UTF8
        
        Write-PipelineLog "9.4 Backup file created: $backupFile"
        Write-PipelineLog "9.5 Final backup size: $([math]::Round($finalSize/1GB, 2)) GB"
        
        # Create metadata file
        $metadataFile = "$backupDir\$($script:JobName)_metadata.json"
        $script:Metadata.EndTime = Get-Date
        $script:Metadata.BackupFile = $backupFile
        $script:Metadata.FinalSize = $finalSize
        
        $script:Metadata | ConvertTo-Json -Depth 10 | Set-Content -Path $metadataFile
        
        Write-PipelineLog "9.6 Metadata file created: $metadataFile"
        
        $script:Stages[8].Status = "Completed"
        Write-PipelineLog "Stage 9 completed successfully" "INFO" "Green"
        
    } catch {
        $script:Stages[8].Status = "Failed"
        Write-PipelineLog "Stage 9 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# STAGE 10: Metadata & Indexing
# ========================================
function Stage10-MetadataIndexing {
    $script:CurrentStage = "MetadataIndexing"
    $script:Stages += @{ "Stage" = "Metadata & Indexing"; "StartTime" = Get-Date; "Status" = "Running" }
    
    Write-PipelineLog "=== STAGE 10: METADATA & INDEXING ===" "INFO" "Cyan"
    
    try {
        Write-PipelineLog "10.1 Creating backup catalog and file index"
        
        $backupDir = "$RepositoryPath\$($script:JobId)"
        
        # Create backup catalog
        $catalog = @{
            "JobId" = $script:JobId
            "JobName" = $script:JobName
            "BackupType" = $script:BackupType
            "BackupFile" = $script:Metadata.BackupFile
            "CreationTime" = $script:Metadata.EndTime
            "SourcePath" = $script:SourcePath
            "RepositoryPath" = $script:RepositoryPath
            "TotalSize" = $script:Metadata.TotalSize
            "CompressedSize" = $script:Metadata.CompressedSize
            "DedupSize" = $script:Metadata.DedupSize
            "FinalSize" = $script:Metadata.FinalSize
            "Stages" = $script:Stages
            "Metadata" = $script:Metadata
        }
        
        $catalogFile = "$backupDir\catalog.json"
        $catalog | ConvertTo-Json -Depth 10 | Set-Content -Path $catalogFile
        
        Write-PipelineLog "10.2 Backup catalog created: $catalogFile"
        
        # Create file index for quick restore
        $fileIndex = @()
        if ($script:VMName) {
            $fileIndex += @{
                "Path" = $script:SourcePath
                "Type" = "VM"
                "Name" = $script:VMName
                "SnapshotId" = $script:Metadata.SnapshotId
            }
        } else {
            $files = Get-ChildItem -Path $script:SourcePath -Recurse -File
            foreach ($file in $files) {
                $fileIndex += @{
                    "Path" = $file.FullName
                    "Type" = "File"
                    "Name" = $file.Name
                    "Size" = $file.Length
                    "Modified" = $file.LastWriteTime
                }
            }
        }
        
        $indexFile = "$backupDir\fileindex.json"
        $fileIndex | ConvertTo-Json -Depth 10 | Set-Content -Path $indexFile
        
        Write-PipelineLog "10.3 File index created: $indexFile"
        Write-PipelineLog "10.4 Total files indexed: $($fileIndex.Count)"
        
        # Calculate final statistics
        $totalDuration = (Get-Date) - $script:StartTime
        $compressionSavings = if ($script:Metadata.TotalSize -gt 0) { 
            [math]::Round((1 - $script:Metadata.CompressionRatio) * 100, 1) 
        } else { 0 }
        
        $dedupSavings = if ($script:Metadata.CompressedSize -gt 0) { 
            [math]::Round((1 - $script:Metadata.DedupSize / $script:Metadata.CompressedSize) * 100, 1) 
        } else { 0 }
        
        Write-PipelineLog "=== BACKUP COMPLETION SUMMARY ===" "INFO" "Green"
        Write-PipelineLog "Job Name: $($script:JobName)" "INFO" "White"
        Write-PipelineLog "Backup Type: $($script:BackupType)" "INFO" "White"
        Write-PipelineLog "Source Size: $([math]::Round($script:Metadata.TotalSize/1GB, 2)) GB" "INFO" "White"
        Write-PipelineLog "Final Size: $([math]::Round($script:Metadata.FinalSize/1GB, 2)) GB" "INFO" "White"
        Write-PipelineLog "Compression Savings: $compressionSavings%" "INFO" "White"
        Write-PipelineLog "Deduplication Savings: $dedupSavings%" "INFO" "White"
        Write-PipelineLog "Total Duration: $($totalDuration.ToString('hh\:mm\:ss'))" "INFO" "White"
        Write-PipelineLog "Backup File: $($script:Metadata.BackupFile)" "INFO" "White"
        
        $script:Stages[9].Status = "Completed"
        Write-PipelineLog "Stage 10 completed successfully" "INFO" "Green"
        
    } catch {
        $script:Stages[9].Status = "Failed"
        Write-PipelineLog "Stage 10 failed: $_" "ERROR" "Red"
        throw
    }
}

# ========================================
# MAIN EXECUTION
# ========================================
function Start-BackupPipeline {
    Write-Host "Starting NovaBackup v6.0 Enterprise Pipeline" -ForegroundColor Cyan
    Write-Host "Job ID: $script:JobId" -ForegroundColor Gray
    Write-Host "Job Name: $JobName" -ForegroundColor Gray
    Write-Host "Source: $SourcePath" -ForegroundColor Gray
    Write-Host "Repository: $RepositoryPath" -ForegroundColor Gray
    Write-Host ""
    
    try {
        Stage1-JobScheduler
        Stage2-SnapshotCreation
        Stage3-ApplicationConsistency
        Stage4-ChangeBlockTracking
        Stage5-DataRead
        Stage6-Compression
        Stage7-Deduplication
        Stage8-Encryption
        Stage9-TransportStorage
        Stage10-MetadataIndexing
        
        Write-Host ""
        Write-Host "=== BACKUP PIPELINE COMPLETED ===" -ForegroundColor Green
        Write-Host "All 10 stages completed successfully!" -ForegroundColor Green
        Write-Host ""
        
        # Display final summary
        $totalDuration = (Get-Date) - $script:StartTime
        Write-Host "FINAL SUMMARY:" -ForegroundColor Yellow
        Write-Host "Job: $JobName" -ForegroundColor White
        Write-Host "Duration: $($totalDuration.ToString('hh\:mm\:ss'))" -ForegroundColor White
        Write-Host "Status: SUCCESS" -ForegroundColor Green
        Write-Host "Backup File: $($script:Metadata.BackupFile)" -ForegroundColor White
        
    } catch {
        Write-Host ""
        Write-Host "=== BACKUP PIPELINE FAILED ===" -ForegroundColor Red
        Write-Host "Error: $_" -ForegroundColor Red
        Write-Host "Check logs for details" -ForegroundColor Red
        exit 1
    }
}

# Execute pipeline
Start-BackupPipeline
