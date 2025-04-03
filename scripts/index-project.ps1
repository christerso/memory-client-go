param (
    [Parameter(Mandatory=$true)]
    [string]$ProjectPath,
    
    [Parameter(Mandatory=$true)]
    [string]$Tag,
    
    [int]$BatchSize = 50,
    
    [int]$MaxFileSizeKB = 1024
)

$outputPath = "indexed_files.txt"

# Check if directory exists
if (-not (Test-Path $ProjectPath -PathType Container)) {
    Write-Error "Directory $ProjectPath does not exist"
    exit 1
}

Write-Host "Indexing project at $ProjectPath with tag: $Tag"

# Get all files in the repository, excluding binary and media files
$files = Get-ChildItem -Path $ProjectPath -Recurse -File | Where-Object {
    $extension = $_.Extension.ToLower()
    -not ($extension -in @(".jpg", ".jpeg", ".png", ".gif", ".mp3", ".mp4", ".avi", ".mov",
                          ".zip", ".rar", ".7z", ".exe", ".dll", ".so", ".bin", ".dat",
                          ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx")) -and
    -not ($_.Name.StartsWith(".")) -and
    ($_.Length -lt ($MaxFileSizeKB * 1KB))
}

$totalFiles = $files.Count
Write-Host "Found $totalFiles files to index"

# Create a file with the list of files to index
$files | ForEach-Object { $_.FullName } | Out-File -FilePath $outputPath

# Process files in batches
$processedCount = 0
$errorCount = 0

for ($i = 0; $i -lt $totalFiles; $i += $BatchSize) {
    $end = [Math]::Min($i + $BatchSize, $totalFiles)
    $batch = $files[$i..($end-1)]

    foreach ($file in $batch) {
        $relativePath = $file.FullName.Substring($ProjectPath.Length).TrimStart("\")

        # Get file content and add to memory client with tag
        try {
            $content = Get-Content -Path $file.FullName -Raw -ErrorAction SilentlyContinue
            if ($null -eq $content) {
                Write-Warning "Could not read content from $($file.FullName), skipping"
                continue
            }

            # Add the file to the memory client with the tag
            $addArgs = @(
                "add",
                "--role", "project",
                "--content", "Path: $relativePath`nTag: $Tag`n`n$content"
            )

            & memory-client $addArgs

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
Write-Host "All files were tagged with: $Tag"
Write-Host "You can search for these files using: memory-client search --tag $Tag"
