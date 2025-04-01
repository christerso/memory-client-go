param (
    [string]$tag = "reign-repo"
)

$repoPath = "g:\repos\reign"
$outputPath = "indexed_files.txt"

# Check if directory exists
if (-not (Test-Path $repoPath -PathType Container)) {
    Write-Error "Directory $repoPath does not exist"
    exit 1
}

Write-Host "Indexing repository at $repoPath with tag: $tag"

# Get all files in the repository, excluding binary and media files
$files = Get-ChildItem -Path $repoPath -Recurse -File | Where-Object {
    $extension = $_.Extension.ToLower()
    -not ($extension -in @(".jpg", ".jpeg", ".png", ".gif", ".mp3", ".mp4", ".avi", ".mov", 
                          ".zip", ".rar", ".7z", ".exe", ".dll", ".so", ".bin", ".dat", 
                          ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx")) -and
    -not ($_.Name.StartsWith(".")) -and
    ($_.Length -lt 1MB)
}

$totalFiles = $files.Count
Write-Host "Found $totalFiles files to index"

# Create a file with the list of files to index
$files | ForEach-Object { $_.FullName } | Out-File -FilePath $outputPath

# Process files in batches
$batchSize = 50
$processedCount = 0
$errorCount = 0

for ($i = 0; $i -lt $totalFiles; $i += $batchSize) {
    $end = [Math]::Min($i + $batchSize, $totalFiles)
    $batch = $files[$i..($end-1)]
    
    foreach ($file in $batch) {
        $relativePath = $file.FullName.Substring($repoPath.Length).TrimStart("\")
        
        # Get file content and add to memory client with tag
        try {
            $content = Get-Content -Path $file.FullName -Raw
            
            # Create a temporary file with the content
            $tempFile = [System.IO.Path]::GetTempFileName()
            $content | Out-File -FilePath $tempFile -Encoding utf8
            
            # Add the file to the memory client with the tag
            $addArgs = @(
                "add",
                "--role", "project",
                "--content", "Path: $relativePath`nTag: $tag`n`n$content"
            )
            
            & .\memory-client.exe $addArgs
            
            # Clean up temp file
            Remove-Item -Path $tempFile -Force
            
            $processedCount++
            
            # Show progress
            $percent = [Math]::Floor(($processedCount * 100) / $totalFiles)
            Write-Progress -Activity "Indexing files" -Status "$processedCount of $totalFiles files processed" -PercentComplete $percent
            
            if ($processedCount % 10 -eq 0) {
                Write-Host "Progress: $percent% ($processedCount/$totalFiles files)"
            }
        }
        catch {
            Write-Error "Error processing file $($file.FullName): $_"
            $errorCount++
        }
    }
}

Write-Host "Indexing complete. Processed $processedCount files with $errorCount errors."
Write-Host "All files were tagged with: $tag"
