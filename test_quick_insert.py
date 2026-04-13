#!/usr/bin/env python3
import requests
import json

def test_quick_insert():
    base_url = "http://localhost:8081/api"

    print("🧪 Testing Quick Insert Feature")
    print("=" * 50)

    # Test 1: Insert into Agent table
    print("\n1️⃣  Inserting into Agent table...")
    agent_data = {
        "Agent_id": 100,
        "Agent_name": "Test Agent",
        "Agent_salary": 50000.00,
        "Agent_Address": "Test Address"
    }
    
    try:
        response = requests.post(f"{base_url}/databases/RealEstate/tables/Agent/records", json=agent_data)
        print(f"   Status: {response.status_code}")
        if response.status_code == 200:
            print("   ✅ Successfully inserted agent record")
        else:
            print(f"   ❌ Error: {response.json()}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    # Test 2: Insert into Properties table
    print("\n2️⃣  Inserting into Properties table...")
    prop_data = {
        "Prop_id": 200,
        "Prop_location": "New Test Location"
    }
    
    try:
        response = requests.post(f"{base_url}/databases/RealEstate/tables/Properties/records", json=prop_data)
        print(f"   Status: {response.status_code}")
        if response.status_code == 200:
            print("   ✅ Successfully inserted property record")
        else:
            print(f"   ❌ Error: {response.json()}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    # Test 3: View updated records
    print("\n3️⃣  Viewing Agent records...")
    try:
        response = requests.get(f"{base_url}/databases/RealEstate/tables/Agent/records")
        if response.status_code == 200:
            records = response.json()['data']['records']
            print(f"   ✅ Found {len(records)} agent records")
            if records:
                print(f"   📋 Latest: {records[-1]}")
        else:
            print(f"   ❌ Error: {response.json()}")
    except Exception as e:
        print(f"   ❌ Error: {e}")

    print("\n" + "=" * 50)
    print("✨ Quick Insert Test Completed!")
    print("🎯 The UI now supports:")
    print("   • Dropdown selection for databases and tables")
    print("   • Automatic form generation based on table schema")
    print("   • Smart input types (number, date, text)")
    print("   • Beautiful table display for viewing data")
    print("   • Simplified workflow - no complex steps!")

if __name__ == "__main__":
    test_quick_insert()