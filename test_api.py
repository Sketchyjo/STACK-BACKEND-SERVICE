#!/usr/bin/env python3
"""
Simple test script to validate the Stack Service API endpoints
"""
import requests
import json
import sys

BASE_URL = "http://localhost:8080/api/v1"

def test_health():
    """Test health endpoint"""
    print("🔍 Testing health endpoint...")
    try:
        response = requests.get(f"{BASE_URL.replace('/api/v1', '')}/health")
        print(f"   Status: {response.status_code}")
        print(f"   Response: {response.json()}")
        return response.status_code == 200
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_register():
    """Test user registration"""
    print("🔍 Testing user registration...")
    try:
        payload = {
            "email": "test@example.com",
            "password": "password123"
        }
        response = requests.post(f"{BASE_URL}/auth/register", json=payload)
        print(f"   Status: {response.status_code}")
        
        if response.status_code == 201:
            data = response.json()
            print(f"   ✅ Success: User registered with ID {data['user']['id']}")
            print(f"   Access Token: {data['accessToken'][:50]}...")
            return data['accessToken'], data['user']['id']
        elif response.status_code == 409:
            print("   ⚠️  User already exists, trying login...")
            return None, None
        else:
            print(f"   ❌ Failed: {response.text}")
            return None, None
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return None, None

def test_login():
    """Test user login"""
    print("🔍 Testing user login...")
    try:
        payload = {
            "email": "test@example.com",
            "password": "password123"
        }
        response = requests.post(f"{BASE_URL}/auth/login", json=payload)
        print(f"   Status: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"   ✅ Success: User logged in with ID {data['user']['id']}")
            print(f"   Access Token: {data['accessToken'][:50]}...")
            return data['accessToken'], data['user']['id']
        else:
            print(f"   ❌ Failed: {response.text}")
            return None, None
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return None, None

def test_onboarding_start(token):
    """Test onboarding start"""
    print("🔍 Testing onboarding start...")
    try:
        payload = {
            "email": "test@example.com"
        }
        # Use mock auth middleware by passing user_id as query param
        params = {"user_id": "550e8400-e29b-41d4-a716-446655440000"}
        response = requests.post(f"{BASE_URL}/onboarding/start", json=payload, params=params)
        print(f"   Status: {response.status_code}")
        
        if response.status_code in [200, 201]:
            data = response.json()
            print(f"   ✅ Success: Onboarding started")
            print(f"   User ID: {data['userId']}")
            print(f"   Status: {data['onboardingStatus']}")
            print(f"   Next Step: {data.get('nextStep', 'N/A')}")
            return True
        else:
            print(f"   ❌ Failed: {response.text}")
            return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def test_onboarding_status():
    """Test onboarding status"""
    print("🔍 Testing onboarding status...")
    try:
        # Use mock auth middleware
        params = {"user_id": "550e8400-e29b-41d4-a716-446655440000"}
        response = requests.get(f"{BASE_URL}/onboarding/status", params=params)
        print(f"   Status: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print(f"   ✅ Success: Retrieved onboarding status")
            print(f"   User ID: {data['userId']}")
            print(f"   Status: {data['onboardingStatus']}")
            print(f"   KYC Status: {data['kycStatus']}")
            return True
        else:
            print(f"   ❌ Failed: {response.text}")
            return False
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return False

def main():
    """Run all tests"""
    print("🚀 Starting API tests for Stack Service...")
    print("=" * 50)
    
    # Test health first
    if not test_health():
        print("❌ Health check failed. Is the server running?")
        sys.exit(1)
    
    # Test registration
    token, user_id = test_register()
    
    # If registration failed due to existing user, try login
    if not token:
        token, user_id = test_login()
    
    if not token:
        print("❌ Authentication failed. Cannot proceed with protected endpoints.")
        sys.exit(1)
    
    # Test onboarding endpoints
    success = True
    success &= test_onboarding_start(token)
    success &= test_onboarding_status()
    
    print("=" * 50)
    if success:
        print("✅ All tests passed! 🎉")
        print("The authentication and onboarding endpoints are working correctly.")
    else:
        print("❌ Some tests failed.")
        sys.exit(1)

if __name__ == "__main__":
    main()