#!/usr/bin/env python3
import requests
import json

def test_api():
    base_url = "http://localhost:8081/api"

    print("Testing API endpoints...")

    # Test 1: Create database
    try:
        response = requests.post(f"{base_url}/databases", json={"name": "TestUI"})
        print(f"Create Database: {response.status_code}")
        print(f"Response: {response.json()}")
    except Exception as e:
        print(f"Create Database Error: {e}")

    # Test 2: Use database
    try:
        response = requests.post(f"{base_url}/databases/use", json={"name": "TestUI"})
        print(f"Use Database: {response.status_code}")
        print(f"Response: {response.json()}")
    except Exception as e:
        print(f"Use Database Error: {e}")

    # Test 3: Create table
    try:
        table_data = {
            "name": "test_table",
            "columns": {
                "id": "INT",
                "name": "VARCHAR(255)",
                "email": "VARCHAR(255)"
            }
        }
        response = requests.post(f"{base_url}/tables", json=table_data)
        print(f"Create Table: {response.status_code}")
        print(f"Response: {response.json()}")
    except Exception as e:
        print(f"Create Table Error: {e}")

    print("API testing completed!")

if __name__ == "__main__":
    test_api()