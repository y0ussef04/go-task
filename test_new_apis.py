#!/usr/bin/env python3
import requests
import json

def test_new_apis():
    base_url = "http://localhost:8081/api"

    print("Testing new APIs...\n")

    # Test 1: Get list of databases
    try:
        response = requests.get(f"{base_url}/databases/list")
        print(f"GET /api/databases/list: {response.status_code}")
        print(f"Response: {json.dumps(response.json(), indent=2)}\n")
    except Exception as e:
        print(f"Error: {e}\n")

    # Test 2: Use a database
    try:
        response = requests.post(f"{base_url}/databases/use", json={"name": "RealEstate"})
        print(f"POST /api/databases/use (RealEstate): {response.status_code}")
        print(f"Response: {json.dumps(response.json(), indent=2)}\n")
    except Exception as e:
        print(f"Error: {e}\n")

    # Test 3: Get list of tables
    try:
        response = requests.get(f"{base_url}/tables/list")
        print(f"GET /api/tables/list: {response.status_code}")
        print(f"Response: {json.dumps(response.json(), indent=2)}\n")
    except Exception as e:
        print(f"Error: {e}\n")

    print("API testing completed!")

if __name__ == "__main__":
    test_new_apis()