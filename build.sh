#!/bin/bash
# This script processes PNG, JPEG, WebP, and HEIC images into optimized WebP assets using ImageMagick.
# It also moves the original source images to a backup directory.

# Source directory for the original images
SOURCE_DIR="new_images/"

# Destination directory for the converted WebP images
DEST_DIR="static/images/"

# Backup directory for the original images
BACKUP_DIR="backup/original_images/"

# Ensure the destination directory and the backup directory exists
mkdir -p "$DEST_DIR"
mkdir -p "$BACKUP_DIR"

# Image dimensions and quality balance recipe-page clarity against faster card loading.
FULL_MAX_DIMENSION="1600"
FULL_QUALITY="80"
CARD_DIMENSION="600"
CARD_QUALITY="75"
CONVERSION_FAILED=0

# Generate both assets before backing up an original so a failed thumbnail conversion never loses source material.
while IFS= read -r -d '' FILE; do
  FILENAME=$(basename "$FILE")
  FILENAME_WITHOUT_EXT="${FILENAME%.*}"
  FULL_OUTPUT_FILE="$DEST_DIR/$FILENAME_WITHOUT_EXT.webp"
  CARD_OUTPUT_FILE="$DEST_DIR/$FILENAME_WITHOUT_EXT-card.webp"

  echo "Processing: $FILE -> $FULL_OUTPUT_FILE and $CARD_OUTPUT_FILE"
  if magick "$FILE" -strip -resize "${FULL_MAX_DIMENSION}x${FULL_MAX_DIMENSION}>" -quality "$FULL_QUALITY" "$FULL_OUTPUT_FILE" && \
    magick "$FILE" -strip -resize "${CARD_DIMENSION}x${CARD_DIMENSION}^" -gravity center -extent "${CARD_DIMENSION}x${CARD_DIMENSION}" -quality "$CARD_QUALITY" "$CARD_OUTPUT_FILE"; then
    # Only move the original after both published assets exist.
    mv "$FILE" "$BACKUP_DIR"
    echo "Successfully created recipe and card images for: $FILENAME"
  else
    echo "Error creating recipe and card images for: $FILE"
    CONVERSION_FAILED=1
  fi
# Match extensions case-insensitively because supplied camera exports may use uppercase extensions.
# ImageMagick's HEIC decoder normalizes camera HEIC files into the same WebP output pipeline as other sources.
done < <(find "$SOURCE_DIR" -maxdepth 1 -type f \( -iname "*.png" -o -iname "*.jpeg" -o -iname "*.jpg" -o -iname "*.webp" -o -iname "*.heic" \) -print0)

if [ "$CONVERSION_FAILED" -ne 0 ]; then
  echo "Image processing failed; original files remain in $SOURCE_DIR."
  exit 1
fi

echo "Image processing complete. Original images have been backed up."

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
