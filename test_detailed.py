#!/usr/bin/env python3
import requests
import json

def test_detailed():
    base_url = "http://localhost:8081/api"

    print("Testing Schema endpoint...")
    response = requests.get(f"{base_url}/tables/Agent/schema")
    print(f"Status: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2)}")

    print("\n\nTesting Records endpoint...")
    response = requests.get(f"{base_url}/tables/Properties/records")
    print(f"Status: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2)}")

if __name__ == "__main__":
    test_detailed()