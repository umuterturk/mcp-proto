"""
In-memory indexer for proto files with fuzzy search capabilities.
"""

import os
from pathlib import Path
from typing import List, Dict, Optional, Set
from rapidfuzz import fuzz, process
import logging

from proto_parser import (
    ProtoFile, ProtoService, ProtoMessage, ProtoEnum,
    ProtoRPC, ProtoField, parse_proto_file
)

logger = logging.getLogger(__name__)


class ProtoIndex:
    """In-memory index of proto files with search capabilities."""
    
    def __init__(self):
        self.files: Dict[str, ProtoFile] = {}
        self.services: Dict[str, ProtoService] = {}
        self.messages: Dict[str, ProtoMessage] = {}
        self.enums: Dict[str, ProtoEnum] = {}
        
        # For fuzzy search
        self._search_entries: List[tuple] = []  # (full_name, type, object)
        
    def index_directory(self, root_path: str) -> int:
        """
        Recursively scan directory for .proto files and index them.
        Returns the number of files indexed.
        """
        root = Path(root_path)
        if not root.exists():
            raise ValueError(f"Directory does not exist: {root_path}")
        
        count = 0
        for proto_file in root.rglob("*.proto"):
            try:
                self.index_file(str(proto_file))
                count += 1
            except Exception as e:
                logger.error(f"Failed to index {proto_file}: {e}")
        
        logger.info(f"Indexed {count} proto files")
        return count
    
    def index_file(self, file_path: str):
        """Parse and index a single proto file."""
        try:
            proto_file = parse_proto_file(file_path)
            self.files[file_path] = proto_file
            
            # Index services
            for service in proto_file.services:
                self.services[service.full_name] = service
                self._search_entries.append((
                    service.full_name,
                    'service',
                    service,
                    file_path
                ))
            
            # Index messages
            for message in proto_file.messages:
                self.messages[message.full_name] = message
                self._search_entries.append((
                    message.full_name,
                    'message',
                    message,
                    file_path
                ))
            
            # Index enums
            for enum in proto_file.enums:
                self.enums[enum.full_name] = enum
                self._search_entries.append((
                    enum.full_name,
                    'enum',
                    enum,
                    file_path
                ))
                
            logger.debug(f"Indexed {file_path}: "
                        f"{len(proto_file.services)} services, "
                        f"{len(proto_file.messages)} messages, "
                        f"{len(proto_file.enums)} enums")
                        
        except Exception as e:
            logger.error(f"Error indexing {file_path}: {e}")
            raise
    
    def remove_file(self, file_path: str):
        """Remove a file from the index."""
        if file_path not in self.files:
            return
        
        proto_file = self.files[file_path]
        
        # Remove services
        for service in proto_file.services:
            self.services.pop(service.full_name, None)
        
        # Remove messages
        for message in proto_file.messages:
            self.messages.pop(message.full_name, None)
        
        # Remove enums
        for enum in proto_file.enums:
            self.enums.pop(enum.full_name, None)
        
        # Remove from search entries
        self._search_entries = [
            entry for entry in self._search_entries
            if entry[3] != file_path
        ]
        
        del self.files[file_path]
        logger.debug(f"Removed {file_path} from index")
    
    def search(self, query: str, limit: int = 20, min_score: int = 60) -> List[Dict]:
        """
        Fuzzy search across all proto definitions.
        
        Args:
            query: Search query string
            limit: Maximum number of results to return
            min_score: Minimum fuzzy match score (0-100)
        
        Returns:
            List of search results with metadata
        """
        if not query:
            return []
        
        results = []
        
        # Search in names
        name_matches = process.extract(
            query,
            [entry[0] for entry in self._search_entries],
            scorer=fuzz.WRatio,
            limit=limit * 2
        )
        
        # Collect high-scoring matches
        seen = set()
        for match_text, score, idx in name_matches:
            if score < min_score:
                continue
                
            entry = self._search_entries[idx]
            full_name, entry_type, obj, file_path = entry
            
            if full_name in seen:
                continue
            seen.add(full_name)
            
            result = {
                'name': full_name,
                'type': entry_type,
                'file': file_path,
                'score': score,
                'match_type': 'name'
            }
            
            # Add type-specific metadata
            if entry_type == 'service':
                result['rpcs'] = [rpc.name for rpc in obj.rpcs]
                result['rpc_count'] = len(obj.rpcs)
            elif entry_type == 'message':
                result['fields'] = [f.name for f in obj.fields]
                result['field_count'] = len(obj.fields)
            elif entry_type == 'enum':
                result['values'] = [v.name for v in obj.values]
                result['value_count'] = len(obj.values)
            
            if obj.comment:
                result['comment'] = obj.comment
            
            results.append(result)
        
        # Also search in comments and field names
        for full_name, entry_type, obj, file_path in self._search_entries:
            if full_name in seen:
                continue
            
            # Search in comments
            if obj.comment and fuzz.partial_ratio(query.lower(), obj.comment.lower()) >= min_score:
                result = {
                    'name': full_name,
                    'type': entry_type,
                    'file': file_path,
                    'score': fuzz.partial_ratio(query.lower(), obj.comment.lower()),
                    'match_type': 'comment',
                    'comment': obj.comment
                }
                results.append(result)
                seen.add(full_name)
                continue
            
            # Search in field names for messages
            if entry_type == 'message':
                for field in obj.fields:
                    field_score = fuzz.ratio(query.lower(), field.name.lower())
                    if field_score >= min_score:
                        result = {
                            'name': full_name,
                            'type': entry_type,
                            'file': file_path,
                            'score': field_score,
                            'match_type': 'field',
                            'matched_field': field.name,
                            'fields': [f.name for f in obj.fields],
                            'field_count': len(obj.fields)
                        }
                        results.append(result)
                        seen.add(full_name)
                        break
            
            # Search in RPC names for services
            if entry_type == 'service':
                for rpc in obj.rpcs:
                    rpc_score = fuzz.ratio(query.lower(), rpc.name.lower())
                    if rpc_score >= min_score:
                        result = {
                            'name': full_name,
                            'type': entry_type,
                            'file': file_path,
                            'score': rpc_score,
                            'match_type': 'rpc',
                            'matched_rpc': rpc.name,
                            'rpcs': [r.name for r in obj.rpcs],
                            'rpc_count': len(obj.rpcs)
                        }
                        results.append(result)
                        seen.add(full_name)
                        break
        
        # Sort by score and limit
        results.sort(key=lambda x: x['score'], reverse=True)
        return results[:limit]
    
    def get_service(self, name: str, resolve_types: bool = True, max_depth: int = 10) -> Optional[Dict]:
        """
        Get full service definition by name with optional recursive type resolution.
        
        Args:
            name: Service name (simple or fully qualified)
            resolve_types: If True, recursively resolve all request/response types
            max_depth: Maximum recursion depth to prevent infinite loops
        
        Supports both simple name and fully qualified name.
        """
        # Try exact match first
        service = self.services.get(name)
        
        # Try fuzzy match if exact match fails
        if not service:
            for full_name, svc in self.services.items():
                if full_name.endswith(f".{name}") or svc.name == name:
                    service = svc
                    break
        
        if not service:
            return None
        
        result = {
            'name': service.name,
            'full_name': service.full_name,
            'comment': service.comment,
            'rpcs': [
                {
                    'name': rpc.name,
                    'request_type': rpc.request_type,
                    'response_type': rpc.response_type,
                    'request_streaming': rpc.request_streaming,
                    'response_streaming': rpc.response_streaming,
                    'comment': rpc.comment
                }
                for rpc in service.rpcs
            ],
            'file': self._find_file_for_definition(service.full_name, 'service')
        }
        
        # Recursively resolve request/response types
        if resolve_types and max_depth > 0:
            resolved_types = {}
            visited = set()
            
            for rpc in service.rpcs:
                # Resolve request type
                req_msg = self._find_message_by_type(rpc.request_type, service.full_name)
                if req_msg and rpc.request_type not in visited:
                    visited.add(rpc.request_type)
                    resolved_types[rpc.request_type] = {
                        'kind': 'message',
                        'name': req_msg.name,
                        'full_name': req_msg.full_name,
                        'comment': req_msg.comment,
                        'fields': [
                            {
                                'name': f.name,
                                'type': f.type,
                                'number': f.number,
                                'label': f.label,
                                'comment': f.comment
                            }
                            for f in req_msg.fields
                        ],
                        'file': self._find_file_for_definition(req_msg.full_name, 'message')
                    }
                    # Recursively resolve nested types
                    nested = self._resolve_message_types(req_msg, max_depth - 1, visited)
                    resolved_types.update(nested)
                
                # Resolve response type
                resp_msg = self._find_message_by_type(rpc.response_type, service.full_name)
                if resp_msg and rpc.response_type not in visited:
                    visited.add(rpc.response_type)
                    resolved_types[rpc.response_type] = {
                        'kind': 'message',
                        'name': resp_msg.name,
                        'full_name': resp_msg.full_name,
                        'comment': resp_msg.comment,
                        'fields': [
                            {
                                'name': f.name,
                                'type': f.type,
                                'number': f.number,
                                'label': f.label,
                                'comment': f.comment
                            }
                            for f in resp_msg.fields
                        ],
                        'file': self._find_file_for_definition(resp_msg.full_name, 'message')
                    }
                    # Recursively resolve nested types
                    nested = self._resolve_message_types(resp_msg, max_depth - 1, visited)
                    resolved_types.update(nested)
            
            if resolved_types:
                result['resolved_types'] = resolved_types
        
        return result
    
    def get_message(self, name: str, resolve_types: bool = True, max_depth: int = 10) -> Optional[Dict]:
        """
        Get full message definition by name with optional recursive type resolution.
        
        Args:
            name: Message name (simple or fully qualified)
            resolve_types: If True, recursively resolve all field types
            max_depth: Maximum recursion depth to prevent infinite loops
        
        Supports both simple name and fully qualified name.
        """
        # Try exact match first
        message = self.messages.get(name)
        
        # Try fuzzy match if exact match fails
        if not message:
            for full_name, msg in self.messages.items():
                if full_name.endswith(f".{name}") or msg.name == name:
                    message = msg
                    break
        
        if not message:
            return None
        
        result = {
            'name': message.name,
            'full_name': message.full_name,
            'comment': message.comment,
            'fields': [
                {
                    'name': field.name,
                    'type': field.type,
                    'number': field.number,
                    'label': field.label,
                    'comment': field.comment
                }
                for field in message.fields
            ],
            'file': self._find_file_for_definition(message.full_name, 'message')
        }
        
        # Recursively resolve field types
        if resolve_types and max_depth > 0:
            resolved_types = self._resolve_message_types(message, max_depth)
            if resolved_types:
                result['resolved_types'] = resolved_types
        
        return result
    
    def get_enum(self, name: str) -> Optional[Dict]:
        """
        Get full enum definition by name.
        Supports both simple name and fully qualified name.
        """
        # Try exact match first
        proto_enum = self.enums.get(name)
        
        # Try fuzzy match if exact match fails
        if not proto_enum:
            for full_name, enm in self.enums.items():
                if full_name.endswith(f".{name}") or enm.name == name:
                    proto_enum = enm
                    break
        
        if not proto_enum:
            return None
        
        return {
            'name': proto_enum.name,
            'full_name': proto_enum.full_name,
            'comment': proto_enum.comment,
            'values': [
                {
                    'name': value.name,
                    'number': value.number,
                    'comment': value.comment
                }
                for value in proto_enum.values
            ],
            'file': self._find_file_for_definition(proto_enum.full_name, 'enum')
        }
    
    def _resolve_message_types(self, message: ProtoMessage, max_depth: int, visited: Optional[Set[str]] = None) -> Dict:
        """
        Recursively resolve all field types in a message.
        
        Args:
            message: The message to resolve
            max_depth: Maximum recursion depth
            visited: Set of already visited type names to prevent cycles
        
        Returns:
            Dictionary mapping type names to their definitions
        """
        if visited is None:
            visited = set()
        
        if max_depth <= 0:
            return {}
        
        resolved = {}
        
        for field in message.fields:
            field_type = field.type
            
            # Skip primitive types
            if field_type in ['string', 'int32', 'int64', 'uint32', 'uint64', 
                             'sint32', 'sint64', 'fixed32', 'fixed64', 'sfixed32', 
                             'sfixed64', 'bool', 'bytes', 'float', 'double']:
                continue
            
            # Skip if already visited (circular reference)
            if field_type in visited:
                continue
            
            visited.add(field_type)
            
            # Try to resolve as message
            msg = self._find_message_by_type(field_type, message.full_name)
            if msg:
                resolved[field_type] = {
                    'kind': 'message',
                    'name': msg.name,
                    'full_name': msg.full_name,
                    'comment': msg.comment,
                    'fields': [
                        {
                            'name': f.name,
                            'type': f.type,
                            'number': f.number,
                            'label': f.label,
                            'comment': f.comment
                        }
                        for f in msg.fields
                    ],
                    'file': self._find_file_for_definition(msg.full_name, 'message')
                }
                
                # Recursively resolve nested types
                nested_types = self._resolve_message_types(msg, max_depth - 1, visited)
                resolved.update(nested_types)
                continue
            
            # Try to resolve as enum
            enum = self._find_enum_by_type(field_type, message.full_name)
            if enum:
                resolved[field_type] = {
                    'kind': 'enum',
                    'name': enum.name,
                    'full_name': enum.full_name,
                    'comment': enum.comment,
                    'values': [
                        {
                            'name': v.name,
                            'number': v.number,
                            'comment': v.comment
                        }
                        for v in enum.values
                    ],
                    'file': self._find_file_for_definition(enum.full_name, 'enum')
                }
        
        return resolved
    
    def _find_message_by_type(self, type_name: str, context_package: str) -> Optional[ProtoMessage]:
        """
        Find a message by type name, considering package context.
        
        Args:
            type_name: The type name (may be simple or qualified)
            context_package: The package of the referring message
        
        Returns:
            ProtoMessage if found, None otherwise
        """
        # Try exact match
        if type_name in self.messages:
            return self.messages[type_name]
        
        # Try with context package
        package_prefix = context_package.rsplit('.', 1)[0] if '.' in context_package else ''
        if package_prefix:
            qualified_name = f"{package_prefix}.{type_name}"
            if qualified_name in self.messages:
                return self.messages[qualified_name]
        
        # Try matching by simple name
        for full_name, msg in self.messages.items():
            if msg.name == type_name or full_name.endswith(f".{type_name}"):
                return msg
        
        return None
    
    def _find_enum_by_type(self, type_name: str, context_package: str) -> Optional[ProtoEnum]:
        """
        Find an enum by type name, considering package context.
        
        Args:
            type_name: The type name (may be simple or qualified)
            context_package: The package of the referring message
        
        Returns:
            ProtoEnum if found, None otherwise
        """
        # Try exact match
        if type_name in self.enums:
            return self.enums[type_name]
        
        # Try with context package
        package_prefix = context_package.rsplit('.', 1)[0] if '.' in context_package else ''
        if package_prefix:
            qualified_name = f"{package_prefix}.{type_name}"
            if qualified_name in self.enums:
                return self.enums[qualified_name]
        
        # Try matching by simple name
        for full_name, enum in self.enums.items():
            if enum.name == type_name or full_name.endswith(f".{type_name}"):
                return enum
        
        return None
    
    def _find_file_for_definition(self, full_name: str, def_type: str) -> Optional[str]:
        """Find the file path that contains a definition."""
        for file_path, proto_file in self.files.items():
            if def_type == 'service':
                if any(s.full_name == full_name for s in proto_file.services):
                    return file_path
            elif def_type == 'message':
                if any(m.full_name == full_name for m in proto_file.messages):
                    return file_path
            elif def_type == 'enum':
                if any(e.full_name == full_name for e in proto_file.enums):
                    return file_path
        return None
    
    def get_stats(self) -> Dict:
        """Get statistics about the indexed proto files."""
        return {
            'total_files': len(self.files),
            'total_services': len(self.services),
            'total_messages': len(self.messages),
            'total_enums': len(self.enums),
            'total_searchable_entries': len(self._search_entries)
        }

