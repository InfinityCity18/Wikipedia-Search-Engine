#!/bin/bash

TARGET_DIR="./PlainTextWikipedia/processed"

echo "Scanning for JSON files with empty text fields..."

for FILE in "$TARGET_DIR"/*.json; do
    [ -e "$FILE" ] || continue

    if jq -e '.text == "" or .text == null' "$FILE" >/dev/null 2>&1; then
        echo "️Removing: $FILE"
        rm "$FILE"
    fi
done

echo "Cleanup complete!"