#!/bin/bash

# Script to generate a Markdown coverage report from Go coverage output
# Usage: ./generate_coverage_report.sh [input_file] [output_file]
# Defaults: input_file=coverage.txt, output_file=coverage.md

INPUT="${1:-coverage.txt}"
OUTPUT="${2:-coverage.md}"

if [[ ! -f "$INPUT" ]]; then
  echo "Warning: Coverage file not found at $INPUT" >&2
  exit 0
fi

# Extract total coverage percentage
TOTAL_LINE=$(grep -i "total" "$INPUT" || echo "")
PERCENT="N/A"
if [[ -n "$TOTAL_LINE" ]]; then
  PERCENT=$(echo "$TOTAL_LINE" | grep -oP '\d+\.\d+%' | tail -1 || echo "N/A")
fi

# Calculate per-file coverage
echo "### 📊 Test Coverage Report" > "$OUTPUT"
echo "" >> "$OUTPUT"
echo '```' >> "$OUTPUT"

# Remove prefix and aggregate by file
sed 's|switchtube-downloader||g' "$INPUT" | \
  awk '
    # Skip total line
    /total:/ { next }

    NF >= 2 {
      # The format is: file:line:function coverage%
      # Last field is coverage, second-to-last is function name
      coverage = $NF

      # Get the file path (first field up to the second colon)
      split($1, parts, ":")
      file = parts[1]

      # Remove % sign and convert to number
      gsub(/%/, "", coverage)
      cov_num = coverage + 0

      # Accumulate coverage per file
      file_total[file] += cov_num
      file_count[file]++
    }

    END {
      # Calculate and print average coverage per file
      for (file in file_total) {
        avg = file_total[file] / file_count[file]
        printf "%-60s %6.1f%%\n", file, avg
      }
    }
  ' | sort >> "$OUTPUT"

echo '```' >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo "> **Overall Coverage**: $PERCENT" >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo "_Generated on $(date -u '+%Y-%m-%d %H:%M:%S UTC')_" >> "$OUTPUT"

echo "" >> "$OUTPUT"
echo "<details>" >> "$OUTPUT"
echo "<summary>📝 How to view locally</summary>" >> "$OUTPUT"
echo "" >> "$OUTPUT"
echo '```bash' >> "$OUTPUT"
echo "go test -coverprofile=coverage.out ./internal/..." >> "$OUTPUT"
echo "go tool cover -html=coverage.out" >> "$OUTPUT"
echo '```' >> "$OUTPUT"
echo "</details>" >> "$OUTPUT"
