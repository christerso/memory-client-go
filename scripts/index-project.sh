#!/bin/bash
# Script to index a project directory with memory-client

# Default values
BATCH_SIZE=50
MAX_FILE_SIZE_KB=1024

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --project)
      PROJECT_PATH="$2"
      shift 2
      ;;
    --tag)
      TAG="$2"
      shift 2
      ;;
    --batch-size)
      BATCH_SIZE="$2"
      shift 2
      ;;
    --max-file-size)
      MAX_FILE_SIZE_KB="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 --project <path> --tag <tag> [--batch-size <size>] [--max-file-size <kb>]"
      exit 1
      ;;
  esac
done

# Check required parameters
if [ -z "$PROJECT_PATH" ]; then
  echo "Error: Project path is required"
  echo "Usage: $0 --project <path> --tag <tag> [--batch-size <size>] [--max-file-size <kb>]"
  exit 1
fi

if [ -z "$TAG" ]; then
  echo "Error: Tag is required"
  echo "Usage: $0 --project <path> --tag <tag> [--batch-size <size>] [--max-file-size <kb>]"
  exit 1
fi

# Check if directory exists
if [ ! -d "$PROJECT_PATH" ]; then
  echo "Error: Directory $PROJECT_PATH does not exist"
  exit 1
fi

echo "Indexing project at $PROJECT_PATH with tag: $TAG"

# Create temporary file to store list of files
TEMP_FILE=$(mktemp)

# Find all files in the repository, excluding binary and media files
find "$PROJECT_PATH" -type f \
  -not -path "*/\.*" \
  -not -name "*.jpg" -not -name "*.jpeg" -not -name "*.png" -not -name "*.gif" \
  -not -name "*.mp3" -not -name "*.mp4" -not -name "*.avi" -not -name "*.mov" \
  -not -name "*.zip" -not -name "*.rar" -not -name "*.7z" -not -name "*.exe" \
  -not -name "*.dll" -not -name "*.so" -not -name "*.bin" -not -name "*.dat" \
  -not -name "*.pdf" -not -name "*.doc" -not -name "*.docx" -not -name "*.xls" \
  -not -name "*.xlsx" -not -name "*.ppt" -not -name "*.pptx" \
  -size -"$MAX_FILE_SIZE_KB"k > "$TEMP_FILE"

TOTAL_FILES=$(wc -l < "$TEMP_FILE")
echo "Found $TOTAL_FILES files to index"

# Process files in batches
PROCESSED_COUNT=0
ERROR_COUNT=0

while IFS= read -r file; do
  # Get relative path
  RELATIVE_PATH="${file#$PROJECT_PATH/}"
  
  # Process the file
  if [ -f "$file" ]; then
    # Get file content
    CONTENT=$(cat "$file" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
      # Add the file to the memory client with the tag
      memory-client add --role project --content "Path: $RELATIVE_PATH
Tag: $TAG

$CONTENT"
      
      PROCESSED_COUNT=$((PROCESSED_COUNT + 1))
      
      # Show progress
      PERCENT=$((PROCESSED_COUNT * 100 / TOTAL_FILES))
      if [ $((PROCESSED_COUNT % 10)) -eq 0 ]; then
        echo "Progress: $PERCENT% ($PROCESSED_COUNT/$TOTAL_FILES files)"
      fi
    else
      echo "Error reading file: $file"
      ERROR_COUNT=$((ERROR_COUNT + 1))
    fi
  fi
  
  # Process in batches
  if [ $((PROCESSED_COUNT % BATCH_SIZE)) -eq 0 ]; then
    echo "Processed $PROCESSED_COUNT files so far..."
    sleep 1  # Brief pause between batches
  fi
done < "$TEMP_FILE"

# Clean up
rm "$TEMP_FILE"

echo "Indexing complete. Processed $PROCESSED_COUNT files with $ERROR_COUNT errors."
echo "All files were tagged with: $TAG"
echo "You can search for these files using: memory-client search --tag $TAG"
