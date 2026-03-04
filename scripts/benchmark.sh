#!/bin/bash

# benchmark.sh - Run Apache Bench and store key metrics to a CSV file
#
# Usage: ./benchmark.sh -u <url_or_port> -n <total_requests> -c <concurrency> -t <tag> [-o <output_file>]
#
# Examples:
#   ./benchmark.sh -u 5000 -n 5000 -c 500 -t "python_server"
#   ./benchmark.sh -u 8000 -n 5000 -c 500 -t "load_balancer"
#   ./benchmark.sh -u http://localhost:8000/ -n 5000 -c 500 -t "load_balancer" -o results.csv

set -euo pipefail

# --------------------------------------------------------------------------- #
# Defaults
# --------------------------------------------------------------------------- #
OUTPUT_FILE="benchmark_results.csv"

# --------------------------------------------------------------------------- #
# Help
# --------------------------------------------------------------------------- #
usage() {
    cat <<EOF
Usage: $(basename "$0") -u <url_or_port> -n <total_requests> -c <concurrency> -t <tag> [-o <output_file>]

Options:
  -u   Target URL or bare port number
         e.g.  8000  →  http://localhost:8000/
         e.g.  http://localhost:8000/path
  -n   Total number of requests to send
  -c   Concurrency level (parallel requests)
  -t   Label/tag for this test run  (e.g. "python_server", "load_balancer")
  -o   Output CSV file  (default: benchmark_results.csv)
  -h   Show this help message

EOF
    exit 1
}

# --------------------------------------------------------------------------- #
# Parse arguments
# --------------------------------------------------------------------------- #
while getopts "u:n:c:t:o:h" opt; do
    case $opt in
        u) URL_OR_PORT="$OPTARG" ;;
        n) TOTAL_REQUESTS="$OPTARG"  ;;
        c) CONCURRENCY="$OPTARG"     ;;
        t) TAG="$OPTARG"             ;;
        o) OUTPUT_FILE="$OPTARG"     ;;
        h) usage ;;
        *) usage ;;
    esac
done

# --------------------------------------------------------------------------- #
# Validate required arguments
# --------------------------------------------------------------------------- #
missing=()
[[ -z "${URL_OR_PORT:-}"    ]] && missing+=("-u <url_or_port>")
[[ -z "${TOTAL_REQUESTS:-}" ]] && missing+=("-n <total_requests>")
[[ -z "${CONCURRENCY:-}"    ]] && missing+=("-c <concurrency>")
[[ -z "${TAG:-}"            ]] && missing+=("-t <tag>")

if [[ ${#missing[@]} -gt 0 ]]; then
    echo "Error: missing required argument(s): ${missing[*]}"
    echo ""
    usage
fi

# --------------------------------------------------------------------------- #
# Build URL — accept bare port numbers or full URLs
# --------------------------------------------------------------------------- #
if [[ "$URL_OR_PORT" =~ ^[0-9]+$ ]]; then
    URL="http://localhost:${URL_OR_PORT}/"
else
    URL="$URL_OR_PORT"
fi

# --------------------------------------------------------------------------- #
# Dependency check
# --------------------------------------------------------------------------- #
if ! command -v ab &>/dev/null; then
    echo "Error: Apache Bench (ab) is not installed."
    echo "       Install it with:  sudo apt install apache2-utils"
    exit 1
fi

# --------------------------------------------------------------------------- #
# Run benchmark
# --------------------------------------------------------------------------- #
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

echo "=========================================="
echo " Apache Bench Benchmark"
echo "=========================================="
echo "  Tag              : $TAG"
echo "  URL              : $URL"
echo "  Total Requests   : $TOTAL_REQUESTS"
echo "  Concurrency      : $CONCURRENCY"
echo "  Output file      : $OUTPUT_FILE"
echo "  Started at       : $TIMESTAMP"
echo "=========================================="
echo ""

AB_STATUS="success"
if ! AB_OUTPUT=$(ab -n "$TOTAL_REQUESTS" -c "$CONCURRENCY" "$URL" 2>&1); then
    AB_STATUS="failed"
    echo "Error: Apache Bench exited with a non-zero status."
    echo "$AB_OUTPUT"
fi

# Print the raw ab output so the user can read it
echo "$AB_OUTPUT"
echo ""

# --------------------------------------------------------------------------- #
# Parse metrics
# --------------------------------------------------------------------------- #

if [[ "$AB_STATUS" == "success" ]]; then
    # Requests per second  →  "Requests per second:    1234.56 [#/sec] (mean)"
    RPS=$(echo "$AB_OUTPUT" \
        | grep "^Requests per second:" \
        | awk '{print $4}')

    # Time per request — ab prints this label twice; first occurrence is the mean
    # for a single request (n/c view), second is mean across all concurrent reqs.
    TPR_MEAN=$(echo "$AB_OUTPUT" \
        | grep "^Time per request:" \
        | head -1 \
        | awk '{print $4}')

    TPR_MEAN_ALL=$(echo "$AB_OUTPUT" \
        | grep "^Time per request:" \
        | tail -1 \
        | awk '{print $4}')

    # Transfer rate  →  "Transfer rate:          5678.90 [Kbytes/sec] received"
    TRANSFER_RATE=$(echo "$AB_OUTPUT" \
        | grep "^Transfer rate:" \
        | awk '{print $3}')

    # Percentile latencies — match lines like "  50%     23"
    extract_percentile() {
        local pct="$1"
        echo "$AB_OUTPUT" \
            | grep -E "^ *${pct}%" \
            | awk '{print $2}'
    }

    P50=$(extract_percentile 50)
    P66=$(extract_percentile 66)
    P75=$(extract_percentile 75)
    P80=$(extract_percentile 80)
    P90=$(extract_percentile 90)
    P95=$(extract_percentile 95)
    P98=$(extract_percentile 98)
    P99=$(extract_percentile 99)
    P100=$(extract_percentile 100)
else
    # Benchmark failed — zero out all metrics
    RPS=0; TPR_MEAN=0; TPR_MEAN_ALL=0; TRANSFER_RATE=0
    P50=0; P66=0; P75=0; P80=0; P90=0; P95=0; P98=0; P99=0; P100=0
fi

# --------------------------------------------------------------------------- #
# Write CSV
# --------------------------------------------------------------------------- #
CSV_HEADER="timestamp,tag,url,total_requests,concurrency,status,\
requests_per_second,time_per_request_mean_ms,time_per_request_mean_all_ms,\
transfer_rate_kbps,\
p50_ms,p66_ms,p75_ms,p80_ms,p90_ms,p95_ms,p98_ms,p99_ms,p100_ms"

CSV_ROW="\"${TIMESTAMP}\",\"${TAG}\",\"${URL}\",\
${TOTAL_REQUESTS},${CONCURRENCY},\"${AB_STATUS}\",\
${RPS},${TPR_MEAN},${TPR_MEAN_ALL},\
${TRANSFER_RATE},\
${P50},${P66},${P75},${P80},${P90},${P95},${P98},${P99},${P100}"

# Create file with header if it does not exist yet
if [[ ! -f "$OUTPUT_FILE" ]]; then
    echo "$CSV_HEADER" > "$OUTPUT_FILE"
fi

echo "$CSV_ROW" >> "$OUTPUT_FILE"

# --------------------------------------------------------------------------- #
# Summary
# --------------------------------------------------------------------------- #
echo "=========================================="
echo " Metrics extracted"
echo "=========================================="
printf "  %-42s %s\n"  "Benchmark status:"                                 "${AB_STATUS}"
printf "  %-42s %s\n"  "Requests per second:"                              "${RPS} req/s"
printf "  %-42s %s\n"  "Time per request (mean):"                          "${TPR_MEAN} ms"
printf "  %-42s %s\n"  "Time per request (mean, across all concurrent):"   "${TPR_MEAN_ALL} ms"
printf "  %-42s %s\n"  "Transfer rate:"                                    "${TRANSFER_RATE} Kbytes/s"
echo ""
printf "  %-42s %s\n"  "50th percentile:"   "${P50} ms"
printf "  %-42s %s\n"  "66th percentile:"   "${P66} ms"
printf "  %-42s %s\n"  "75th percentile:"   "${P75} ms"
printf "  %-42s %s\n"  "80th percentile:"   "${P80} ms"
printf "  %-42s %s\n"  "90th percentile:"   "${P90} ms"
printf "  %-42s %s\n"  "95th percentile:"   "${P95} ms"
printf "  %-42s %s\n"  "98th percentile:"   "${P98} ms"
printf "  %-42s %s\n"  "99th percentile:"   "${P99} ms"
printf "  %-42s %s\n"  "100th percentile:"  "${P100} ms"
echo "=========================================="
echo " Results appended to: $OUTPUT_FILE"
echo "=========================================="

# Exit with failure code if the benchmark itself failed
[[ "$AB_STATUS" == "failed" ]] && exit 1

