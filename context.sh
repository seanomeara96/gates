#!/bin/bash

# Define the output file
OUTPUT_FILE="llm_context.txt"

# Initialize the output file
echo "# File Content and Context for LLM" > "$OUTPUT_FILE"
echo "This file contains the content of all files in the directory along with their paths for context." >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Check if .gitignore exists
if [ -f ".gitignore" ]; then
    # Process all files, excluding those specified in .gitignore
    find . -type f -not -path "*/\.*" | grep -v -f <(grep -v '^#' .gitignore | grep -v '^\s*$' | sed 's/^/\.\//') | while read -r file; do
        # Skip the output file itself
        if [ "$file" == "./$OUTPUT_FILE" ]; then
            continue
        fi

        echo "## File: $file" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        cat "$file" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        echo "---" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    done
else
    # Process all files if no .gitignore exists
    find . -type f -not -path "*/\.*" | while read -r file; do
        # Skip the output file itself
        if [ "$file" == "./$OUTPUT_FILE" ]; then
            continue
        fi

        echo "## File: $file" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        cat "$file" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        echo "---" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    done
fi

echo "Generated context file: $OUTPUT_FILE"