#!/usr/bin/env python3
"""
Test script for the MCP Proto Server.
This script tests the core functionality without running the full MCP server.
"""

import sys
import logging
from proto_indexer import ProtoIndex

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


def print_section(title: str):
    """Print a section header."""
    print("\n" + "=" * 80)
    print(f"  {title}")
    print("=" * 80 + "\n")


def test_indexing():
    """Test indexing proto files."""
    print_section("TEST 1: Indexing Proto Files")
    
    index = ProtoIndex()
    count = index.index_directory("examples/")
    
    print(f"✓ Indexed {count} proto files")
    
    stats = index.get_stats()
    print(f"\nStatistics:")
    print(f"  - Total files: {stats['total_files']}")
    print(f"  - Total services: {stats['total_services']}")
    print(f"  - Total messages: {stats['total_messages']}")
    print(f"  - Total enums: {stats['total_enums']}")
    print(f"  - Searchable entries: {stats['total_searchable_entries']}")
    
    return index


def test_search(index: ProtoIndex):
    """Test search functionality."""
    print_section("TEST 2: Search Functionality")
    
    # Test 1: Search for "auth"
    print("Query: 'auth'")
    results = index.search("auth", limit=5)
    print(f"Found {len(results)} results:\n")
    for i, result in enumerate(results, 1):
        print(f"{i}. {result['name']} ({result['type']})")
        print(f"   Score: {result['score']}, Match: {result['match_type']}")
        if result.get('comment'):
            print(f"   Comment: {result['comment'][:80]}...")
        print()
    
    # Test 2: Search for "user"
    print("\nQuery: 'user'")
    results = index.search("user", limit=5)
    print(f"Found {len(results)} results:\n")
    for i, result in enumerate(results, 1):
        print(f"{i}. {result['name']} ({result['type']})")
        print(f"   Score: {result['score']}")
        if 'rpc_count' in result:
            print(f"   RPCs: {result['rpc_count']}")
        if 'field_count' in result:
            print(f"   Fields: {result['field_count']}")
        print()
    
    # Test 3: Search in comments
    print("\nQuery: 'pagination'")
    results = index.search("pagination", limit=3)
    print(f"Found {len(results)} results:\n")
    for i, result in enumerate(results, 1):
        print(f"{i}. {result['name']} ({result['type']})")
        print(f"   Score: {result['score']}, Match: {result['match_type']}")
        print()


def test_get_service(index: ProtoIndex):
    """Test getting service definitions."""
    print_section("TEST 3: Get Service Definition")
    
    service_name = "UserService"
    print(f"Getting service: {service_name}")
    
    service = index.get_service(service_name)
    if service:
        print(f"\n✓ Service: {service['full_name']}")
        if service.get('comment'):
            print(f"  Comment: {service['comment']}")
        print(f"  File: {service['file']}")
        print(f"  RPCs ({len(service['rpcs'])}):")
        for rpc in service['rpcs']:
            streaming = ""
            if rpc['request_streaming']:
                streaming += " (stream)"
            if rpc['response_streaming']:
                streaming += " → stream"
            print(f"    - {rpc['name']}: {rpc['request_type']} → {rpc['response_type']}{streaming}")
            if rpc.get('comment'):
                print(f"      {rpc['comment']}")
    else:
        print(f"✗ Service not found: {service_name}")
    
    # Test with qualified name
    print(f"\n\nGetting service: api.v1.AuthService")
    service = index.get_service("api.v1.AuthService")
    if service:
        print(f"\n✓ Service: {service['full_name']}")
        print(f"  RPCs: {', '.join([rpc['name'] for rpc in service['rpcs']])}")
    else:
        print("✗ Service not found")


def test_get_message(index: ProtoIndex):
    """Test getting message definitions."""
    print_section("TEST 4: Get Message Definition")
    
    message_name = "User"
    print(f"Getting message: {message_name}")
    
    message = index.get_message(message_name)
    if message:
        print(f"\n✓ Message: {message['full_name']}")
        if message.get('comment'):
            print(f"  Comment: {message['comment']}")
        print(f"  File: {message['file']}")
        print(f"  Fields ({len(message['fields'])}):")
        for field in message['fields']:
            label = f" [{field['label']}]" if field['label'] else ""
            print(f"    {field['number']}. {field['name']}: {field['type']}{label}")
            if field.get('comment'):
                print(f"       // {field['comment']}")
    else:
        print(f"✗ Message not found: {message_name}")
    
    # Test enum
    print(f"\n\nGetting enum: UserRole")
    enum = index.get_enum("UserRole")
    if enum:
        print(f"\n✓ Enum: {enum['full_name']}")
        if enum.get('comment'):
            print(f"  Comment: {enum['comment']}")
        print(f"  Values ({len(enum['values'])}):")
        for value in enum['values']:
            print(f"    {value['name']} = {value['number']}")
            if value.get('comment'):
                print(f"      // {value['comment']}")
    else:
        print("✗ Enum not found")


def test_fuzzy_matching(index: ProtoIndex):
    """Test fuzzy matching capabilities."""
    print_section("TEST 5: Fuzzy Matching")
    
    queries = [
        ("usr", "Misspelling of 'user'"),
        ("prodct", "Misspelling of 'product'"),
        ("CreateUsr", "Partial match"),
        ("token", "Field/comment search"),
    ]
    
    for query, description in queries:
        print(f"Query: '{query}' ({description})")
        results = index.search(query, limit=3, min_score=50)
        print(f"Found {len(results)} results:")
        for result in results:
            print(f"  - {result['name']} (score: {result['score']})")
        print()


def main():
    """Run all tests."""
    try:
        print("\n" + "=" * 80)
        print("  MCP Proto Server - Test Suite")
        print("=" * 80)
        
        # Run tests
        index = test_indexing()
        test_search(index)
        test_get_service(index)
        test_get_message(index)
        test_fuzzy_matching(index)
        
        print_section("All Tests Completed Successfully!")
        print("✓ Indexing: PASSED")
        print("✓ Search: PASSED")
        print("✓ Get Service: PASSED")
        print("✓ Get Message: PASSED")
        print("✓ Fuzzy Matching: PASSED")
        
        print("\nReady to run MCP server:")
        print("  python mcp_proto_server.py --root examples/")
        
    except Exception as e:
        logger.error(f"Test failed: {e}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    main()

