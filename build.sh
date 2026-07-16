#!/bin/bash
# This script converts images from PNG and JPEG formats to WebP format using ImageMagick.
# It also moves the original images to a backup directory.

# Source directory for the original images
SOURCE_DIR="new_images/"

# Destination directory for the converted WebP images
DEST_DIR="static/images/"

# Backup directory for the original images
BACKUP_DIR="backup/original_images/"

# Ensure the destination directory and the backup directory exists
mkdir -p "$DEST_DIR"
mkdir -p "$BACKUP_DIR"

# Quality setting for WebP conversion (adjust as needed, lower value = more compression, lower quality)
QUALITY="80"

# Find all PNG and JPEG files in the source directory
find "$SOURCE_DIR" -maxdepth 1 -type f -name "*.png" -o -name "*.jpeg" -o -name "*.jpg" | while IFS= read -r FILE; do
  # Extract the filename (without extension)
  FILENAME=$(basename "$FILE")
  FILENAME_WITHOUT_EXT="${FILENAME%.*}"

  # Construct the output filename in the destination directory (WebP format)
  OUTPUT_FILE="$DEST_DIR/$FILENAME_WITHOUT_EXT.webp"

  # Convert the image to WebP using ImageMagick's magick command
  echo "Processing: $FILE -> $OUTPUT_FILE"
  magick "$FILE" -strip -quality "$QUALITY" "$OUTPUT_FILE"

  # Check if the conversion was successful
  if [ $? -eq 0 ]; then
    echo "Successfully converted: $OUTPUT_FILE"
  else
    echo "Error converting: $FILE"
  fi
done

# Move all the new images to the backup directory
mv "$SOURCE_DIR"*.png "$BACKUP_DIR"
mv "$SOURCE_DIR"*.jpeg "$BACKUP_DIR"
mv "$SOURCE_DIR"*.jpg "$BACKUP_DIR"

echo "Image processing complete. New images have been backed up."

# Compile the Go application.
go build -o bin/app src/main.go

# Generate the committed content and docs artifacts without requiring the optional local server.
if ! ./bin/app --build; then
  echo "Site generation failed."
  exit 1
fi

echo "Build successful."

# Ask the user if they want to run the application.
while true; do
    read -p "Do you want to run the application? (y/n): " RUN_APP
    if [[ "$RUN_APP" == "y" || "$RUN_APP" == "Y" ]]; then
      ./bin/app
      break
    elif [[ "$RUN_APP" == "n" || "$RUN_APP" == "N" ]]; then
      echo "Application not run."
      break
    else
      echo "Invalid input. Please enter 'y' or 'n'."
    fi
  done
# End of script
