#!/bin/bash

# Circle Client Test Runner
# This script runs the Circle client test suite

echo "🚀 Starting Circle Client Test Suite"
echo "===================================="

# Check if required environment variables are set
if [ -z "$CIRCLE_API_KEY" ]; then
    echo "❌ Error: CIRCLE_API_KEY environment variable is required"
    echo "Please set your Circle API key:"
    echo "export CIRCLE_API_KEY='your_api_key_here'"
    exit 1
fi

# Set default environment variables if not provided
export CIRCLE_ENVIRONMENT=${CIRCLE_ENVIRONMENT:-"sandbox"}
export CIRCLE_BASE_URL=${CIRCLE_BASE_URL:-""}
export CIRCLE_DEFAULT_WALLET_SET_NAME=${CIRCLE_DEFAULT_WALLET_SET_NAME:-"STACK-Test-WalletSet"}

echo "Configuration:"
echo "  - Environment: $CIRCLE_ENVIRONMENT"
echo "  - API Key: ${CIRCLE_API_KEY:0:10}..."
echo "  - Base URL: ${CIRCLE_BASE_URL:-"default"}"
echo ""

# Run the test
echo "Running Circle client tests..."
go run test_circle_client.go

echo ""
echo "Test execution completed!"
