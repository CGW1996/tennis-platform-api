#!/bin/bash

echo "Testing coach search with frontend-style parameters..."

# Test the coach search endpoint with different parameter styles
echo "Testing with price_min and price_max..."
curl -s "http://localhost:8080/api/v1/coaches?price_min=500&price_max=3000&sort_by=rating&sort_order=desc" | jq '.'

echo ""
echo "Testing with minHourlyRate and maxHourlyRate (legacy)..."
curl -s "http://localhost:8080/api/v1/coaches?minHourlyRate=500&maxHourlyRate=3000&sortBy=rating&sortOrder=desc" | jq '.'

echo ""
echo "Testing with mixed parameters..."
curl -s "http://localhost:8080/api/v1/coaches?price_min=500&max_experience=10&sort_by=experience" | jq '.'