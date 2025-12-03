#!/bin/bash
set -euo pipefail

INPUT="${1:-coverage.txt}"
OUTPUT="${2:-coverage.md}"

if [[ ! -f "$INPUT" ]]; then
    echo "Warning: Coverage file not found at $INPUT" >&2
    exit 0
fi

# Header
{
    echo "### 📊 Test Coverage Report"
    echo
} >"$OUTPUT"

# Extract total coverage percentage
TOTAL_LINE=$(grep -i "total" "$INPUT" || echo "")
PERCENT="N/A"
if [[ -n "$TOTAL_LINE" ]]; then
    PERCENT=$(echo "$TOTAL_LINE" | grep -oP '\d+\.\d+%' | tail -1 || echo "N/A")
fi

{
    echo "> **Overall Coverage**: $PERCENT"
    echo
} >>"$OUTPUT"

# Start collapsible table
{
    echo "<details>"
    echo "<summary>📁 Full Coverage Table</summary>"
    echo
    echo '| Filename | Coverage |'
    echo '|----------|----------|'
} >>"$OUTPUT"

# Aggregate coverage per file
declare -A file_total
declare -A file_count

while read -r line; do
    [[ -z "$line" ]] && continue
    [[ "$line" =~ ^total: ]] && continue

    # Extract filename and coverage
    filename=$(echo "$line" | cut -d: -f1 | sed 's|switchtube-downloader||g')
    coverage_str=$(echo "$line" | awk '{print $NF}')
    coverage_num=${coverage_str//%/}

    file_total["$filename"]=$(awk -v sum="${file_total["$filename"]:-0}" -v cov="$coverage_num" 'BEGIN { print sum + cov }')
    file_count["$filename"]=$((${file_count["$filename"]:-0} + 1))
done <"$INPUT"

# Collect rows, sort by coverage descending
rows=()

for filename in "${!file_total[@]}"; do
    sum=${file_total["$filename"]}
    count=${file_count["$filename"]}
    avg=$(awk -v s="$sum" -v c="$count" 'BEGIN { printf "%.1f", s / c }')
    coverage="${avg}%"
    pct_num="$avg"
    rows+=("${pct_num}|${filename}|${coverage}")
done

if [[ ${#rows[@]} -gt 0 ]]; then
    sorted_rows=$(printf "%s\n" "${rows[@]}" | sort -t '|' -k1,1nr)
else
    sorted_rows=""
fi

# Emit sorted rows
if [[ -n "$sorted_rows" ]]; then
    while IFS='|' read -r pct filename coverage; do
        [[ -z "$filename" ]] && continue
        echo "| $filename | $coverage |" >>"$OUTPUT"
    done <<<"$sorted_rows"
fi

# Close collapsible + instructions
{
    echo
    echo "</details>"
    echo
    echo "<details>"
    echo "<summary>📝 How to view locally</summary>"
    echo
    echo '```bash'
    echo "go test -coverprofile=coverage.out ./internal/..."
    echo "go tool cover -html=coverage.out"
    echo '```'
    echo "</details>"
    echo
    echo "_Generated on $(date -u '+%Y-%m-%d %H:%M:%S UTC')_"
} >>"$OUTPUT"
