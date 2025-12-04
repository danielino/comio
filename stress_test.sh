#!/bin/bash

# Stress test for ComIO S3-compatible storage
# Tests with various file sizes to stress the slab allocator

BASE_URL="http://localhost:8080"
BUCKET="test"
NUM_SMALL_FILES=100    # 1-100KB files
NUM_MEDIUM_FILES=50    # 100KB-1MB files
NUM_LARGE_FILES=20     # 1-5MB files

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "======================================"
echo "ComIO Stress Test"
echo "======================================"
echo "Small files (1-100KB): $NUM_SMALL_FILES"
echo "Medium files (100KB-1MB): $NUM_MEDIUM_FILES"
echo "Large files (1-5MB): $NUM_LARGE_FILES"
echo "======================================"
echo ""

success_count=0
error_count=0
start_time=$(date +%s)

# Function to upload a file
upload_file() {
    local key=$1
    local size=$2

    # Create random data
    dd if=/dev/urandom bs=$size count=1 2>/dev/null | \
        curl -s -X PUT -H "Content-Type: application/octet-stream" \
        --data-binary @- "$BASE_URL/$BUCKET/$key" -o /dev/null -w "%{http_code}"
}

# Test small files (1-100KB)
echo "Testing small files (1-100KB)..."
for i in $(seq 1 $NUM_SMALL_FILES); do
    size=$((RANDOM % 100 + 1))  # 1-100KB
    size_bytes=$((size * 1024))
    key="small_${i}_${size}k"

    http_code=$(upload_file "$key" "$size_bytes")

    if [ "$http_code" = "200" ]; then
        ((success_count++))
        echo -ne "${GREEN}.${NC}"
    else
        ((error_count++))
        echo -ne "${RED}F${NC}"
        echo "" >> stress_errors.log
        echo "Failed: $key (size: ${size}KB, http: $http_code)" >> stress_errors.log
    fi

    # Progress every 10 files
    if [ $((i % 10)) -eq 0 ]; then
        echo -ne " [$i/$NUM_SMALL_FILES]\n"
    fi
done
echo ""

# Test medium files (100KB-1MB)
echo "Testing medium files (100KB-1MB)..."
for i in $(seq 1 $NUM_MEDIUM_FILES); do
    size=$((RANDOM % 900 + 100))  # 100-1000KB
    size_bytes=$((size * 1024))
    key="medium_${i}_${size}k"

    http_code=$(upload_file "$key" "$size_bytes")

    if [ "$http_code" = "200" ]; then
        ((success_count++))
        echo -ne "${GREEN}.${NC}"
    else
        ((error_count++))
        echo -ne "${RED}F${NC}"
        echo "" >> stress_errors.log
        echo "Failed: $key (size: ${size}KB, http: $http_code)" >> stress_errors.log
    fi

    # Progress every 5 files
    if [ $((i % 5)) -eq 0 ]; then
        echo -ne " [$i/$NUM_MEDIUM_FILES]\n"
    fi
done
echo ""

# Test large files (1-5MB)
echo "Testing large files (1-5MB)..."
for i in $(seq 1 $NUM_LARGE_FILES); do
    size=$((RANDOM % 4 + 1))  # 1-5MB
    size_bytes=$((size * 1024 * 1024))
    key="large_${i}_${size}mb"

    http_code=$(upload_file "$key" "$size_bytes")

    if [ "$http_code" = "200" ]; then
        ((success_count++))
        echo -ne "${GREEN}.${NC}"
    else
        ((error_count++))
        echo -ne "${RED}F${NC}"
        echo "" >> stress_errors.log
        echo "Failed: $key (size: ${size}MB, http: $http_code)" >> stress_errors.log
    fi

    # Progress every 2 files
    if [ $((i % 2)) -eq 0 ]; then
        echo -ne " [$i/$NUM_LARGE_FILES]\n"
    fi
done
echo ""

end_time=$(date +%s)
duration=$((end_time - start_time))

echo ""
echo "======================================"
echo "Stress Test Results"
echo "======================================"
echo -e "Successful uploads: ${GREEN}$success_count${NC}"
echo -e "Failed uploads: ${RED}$error_count${NC}"
echo "Total duration: ${duration}s"
echo "======================================"

if [ $error_count -gt 0 ]; then
    echo ""
    echo -e "${YELLOW}Errors logged to stress_errors.log${NC}"
fi

# Show server metrics
echo ""
echo "======================================"
echo "Server Metrics"
echo "======================================"
./bin/comio admin metrics

exit $error_count
