@echo off
echo Indexing g:\repos\reign with tag "reign-repo"

REM First, check if the directory exists
if not exist g:\repos\reign (
  echo Error: Directory g:\repos\reign does not exist
  exit /b 1
)

echo Starting indexing process...
echo This may take some time depending on the size of the repository.

REM Use the memory-client to add each file with the tag
for /r g:\repos\reign %%f in (*) do (
  set "filepath=%%f"
  set "ext=%%~xf"
  
  REM Skip binary and media files
  echo %%f | findstr /i ".jpg .jpeg .png .gif .mp3 .mp4 .avi .mov .zip .rar .7z .exe .dll .so .bin .dat .pdf .doc .docx .xls .xlsx .ppt .pptx" > nul
  if errorlevel 1 (
    REM Skip files larger than 1MB (1048576 bytes)
    for %%s in ("%%f") do (
      if %%~zs LSS 1048576 (
        REM Add the file to the memory client with the tag
        echo Adding: %%f
        memory-client.exe add --role project --content "TAG: reign-repo^PATH: %%f^CONTENT: " --file "%%f"
      )
    )
  )
)

echo Indexing complete!
echo All files have been tagged with "reign-repo"
