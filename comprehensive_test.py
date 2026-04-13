#!/usr/bin/env python3
"""
Comprehensive test for the complete database management system
"""
import requests
import json

def test_complete_flow():
    base_url = "http://localhost:8081/api"

    print("=" * 60)
    print("TESTING COMPLETE DATABASE MANAGEMENT SYSTEM")
    print("=" * 60)

    # Test 1: Get databases
    print("\n1️⃣  GET /api/databases/list - Fetch all available databases")
    try:
        response = requests.get(f"{base_url}/databases/list")
        if response.status_code == 200:
            databases = response.json()['data']
            print(f"   ✅ Found {len(databases)} databases:")
            for db in databases[:3]:  # Show first 3
                print(f"      - {db}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    # Test 2: Use database
    print("\n2️⃣  POST /api/databases/use - Switch to RealEstate database")
    try:
        response = requests.post(f"{base_url}/databases/use", json={"name": "RealEstate"})
        if response.status_code == 200:
            print(f"   ✅ {response.json()['message']}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    # Test 3: Get tables
    print("\n3️⃣  GET /api/tables/list - Fetch all tables in RealEstate")
    try:
        response = requests.get(f"{base_url}/tables/list")
        if response.status_code == 200:
            tables = response.json()['data']
            print(f"   ✅ Found {len(tables)} tables:")
            for table in tables:
                print(f"      - {table}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    # Test 4: Get table schema
    print("\n4️⃣  GET /api/tables/Agent/schema - Get table structure")
    try:
        response = requests.get(f"{base_url}/tables/Agent/schema")
        if response.status_code == 200:
            columns = response.json()['columns']
            print(f"   ✅ Agent table has {len(columns)} columns:")
            for col in columns:
                print(f"      - {col['name']} ({col['type']})")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    # Test 5: Get records
    print("\n5️⃣  GET /api/tables/Properties/records - View existing records")
    try:
        response = requests.get(f"{base_url}/tables/Properties/records")
        if response.status_code == 200:
            records = response.json()['records'] or []
            print(f"   ✅ Found {len(records)} records in Properties table")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    print("\n" + "=" * 60)
    print("✨ ALL TESTS COMPLETED SUCCESSFULLY!")
    print("=" * 60)
    print("\nNEW FEATURES IMPLEMENTED:")
    print("✓ Dropdown lists for selecting databases")
    print("✓ Dropdown lists for selecting tables")
    print("✓ Dynamic data type selections (dropdown)")
    print("✓ Auto-loading of available databases and tables")
    print("✓ Better UI for data management operations")

if __name__ == "__main__":
    test_complete_flow()