#!/bin/bash

# Test script for authentication API endpoints
BASE_URL="http://localhost:8080/api/v1"

echo "Testing Authentication API..."

# Test user registration
echo "1. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "firstName": "Test",
    "lastName": "User"
  }')

echo "Register Response: $REGISTER_RESPONSE"

# Extract access token from registration response
ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $REGISTER_RESPONSE | grep -o '"refreshToken":"[^"]*' | cut -d'"' -f4)

echo "Access Token: $ACCESS_TOKEN"
echo "Refresh Token: $REFRESH_TOKEN"

# Test user login
echo -e "\n2. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

echo "Login Response: $LOGIN_RESPONSE"

# Test protected endpoint - get user profile
echo -e "\n3. Testing protected endpoint - get user profile..."
if [ ! -z "$ACCESS_TOKEN" ]; then
  PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/users/profile" \
    -H "Authorization: Bearer $ACCESS_TOKEN")
  echo "Profile Response: $PROFILE_RESPONSE"
else
  echo "No access token available for profile test"
fi

# Test refresh token
echo -e "\n4. Testing refresh token..."
if [ ! -z "$REFRESH_TOKEN" ]; then
  REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
    -H "Content-Type: application/json" \
    -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}")
  echo "Refresh Response: $REFRESH_RESPONSE"
else
  echo "No refresh token available for refresh test"
fi

# Test logout
echo -e "\n5. Testing logout..."
if [ ! -z "$REFRESH_TOKEN" ]; then
  LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
    -H "Content-Type: application/json" \
    -d "{\"refreshToken\": \"$REFRESH_TOKEN\"}")
  echo "Logout Response: $LOGOUT_RESPONSE"
else
  echo "No refresh token available for logout test"
fi

echo -e "\nAPI testing completed!"