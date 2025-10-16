"""
Proto file parser for extracting structure from .proto files.
Supports proto2 and proto3 syntax.
"""

import re
from dataclasses import dataclass, field
from typing import List, Optional, Dict
from pathlib import Path


@dataclass
class ProtoField:
    """Represents a field in a message or enum value."""
    name: str
    type: str
    number: int
    label: Optional[str] = None  # optional, repeated, required
    comment: Optional[str] = None
    options: Dict[str, str] = field(default_factory=dict)


@dataclass
class ProtoMessage:
    """Represents a message definition."""
    name: str
    full_name: str
    fields: List[ProtoField] = field(default_factory=list)
    nested_messages: List['ProtoMessage'] = field(default_factory=list)
    nested_enums: List['ProtoEnum'] = field(default_factory=list)
    comment: Optional[str] = None


@dataclass
class ProtoEnum:
    """Represents an enum definition."""
    name: str
    full_name: str
    values: List[ProtoField] = field(default_factory=list)
    comment: Optional[str] = None


@dataclass
class ProtoRPC:
    """Represents an RPC method in a service."""
    name: str
    request_type: str
    response_type: str
    request_streaming: bool = False
    response_streaming: bool = False
    comment: Optional[str] = None


@dataclass
class ProtoService:
    """Represents a service definition."""
    name: str
    full_name: str
    rpcs: List[ProtoRPC] = field(default_factory=list)
    comment: Optional[str] = None


@dataclass
class ProtoFile:
    """Represents a complete parsed proto file."""
    path: str
    package: str = ""
    syntax: str = "proto2"
    services: List[ProtoService] = field(default_factory=list)
    messages: List[ProtoMessage] = field(default_factory=list)
    enums: List[ProtoEnum] = field(default_factory=list)
    imports: List[str] = field(default_factory=list)


class ProtoParser:
    """Parser for .proto files."""
    
    def __init__(self):
        self.current_package = ""
        
    def parse_file(self, file_path: str) -> ProtoFile:
        """Parse a .proto file and extract all definitions."""
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        proto_file = ProtoFile(path=file_path)
        
        # Remove C++ style comments but keep them for parsing
        lines_with_comments = self._preprocess_content(content)
        
        # Parse top-level constructs
        proto_file.syntax = self._extract_syntax(content)
        proto_file.package = self._extract_package(content)
        self.current_package = proto_file.package
        proto_file.imports = self._extract_imports(content)
        
        # Parse services, messages, and enums
        proto_file.services = self._extract_services(lines_with_comments)
        proto_file.messages = self._extract_messages(lines_with_comments)
        proto_file.enums = self._extract_enums(lines_with_comments)
        
        return proto_file
    
    def _preprocess_content(self, content: str) -> List[tuple]:
        """Preprocess content and extract lines with their comments."""
        lines = []
        current_comment = []
        
        for line in content.split('\n'):
            # Extract trailing comment
            comment_match = re.search(r'//(.*)$', line)
            comment = comment_match.group(1).strip() if comment_match else None
            
            # Remove comment from line
            line_without_comment = re.sub(r'//.*$', '', line).strip()
            
            # Check if this is a standalone comment line
            if not line_without_comment and comment:
                current_comment.append(comment)
            else:
                # Attach accumulated comments to this line
                full_comment = ' '.join(current_comment) if current_comment else None
                if comment and full_comment:
                    full_comment = f"{full_comment} {comment}"
                elif comment:
                    full_comment = comment
                    
                lines.append((line_without_comment, full_comment))
                current_comment = []
        
        return lines
    
    def _extract_syntax(self, content: str) -> str:
        """Extract syntax version (proto2 or proto3)."""
        match = re.search(r'syntax\s*=\s*["\'](\w+)["\']', content)
        return match.group(1) if match else "proto2"
    
    def _extract_package(self, content: str) -> str:
        """Extract package name."""
        match = re.search(r'package\s+([\w.]+)\s*;', content)
        return match.group(1) if match else ""
    
    def _extract_imports(self, content: str) -> List[str]:
        """Extract import statements."""
        imports = []
        for match in re.finditer(r'import\s+(?:public\s+|weak\s+)?["\']([^"\']+)["\']', content):
            imports.append(match.group(1))
        return imports
    
    def _extract_services(self, lines: List[tuple]) -> List[ProtoService]:
        """Extract service definitions."""
        services = []
        content = '\n'.join([line for line, _ in lines])
        
        # Find all service blocks
        for service_match in re.finditer(
            r'service\s+(\w+)\s*\{([^}]*)\}',
            content,
            re.DOTALL
        ):
            service_name = service_match.group(1)
            service_body = service_match.group(2)
            
            # Get comment for service
            service_comment = self._find_comment_for_construct(lines, service_name, 'service')
            
            full_name = f"{self.current_package}.{service_name}" if self.current_package else service_name
            service = ProtoService(
                name=service_name,
                full_name=full_name,
                comment=service_comment
            )
            
            # Extract RPCs
            for rpc_match in re.finditer(
                r'rpc\s+(\w+)\s*\(\s*(stream\s+)?(\w+)\s*\)\s*returns\s*\(\s*(stream\s+)?(\w+)\s*\)',
                service_body
            ):
                rpc_name = rpc_match.group(1)
                request_streaming = bool(rpc_match.group(2))
                request_type = rpc_match.group(3)
                response_streaming = bool(rpc_match.group(4))
                response_type = rpc_match.group(5)
                
                # Get comment for RPC
                rpc_comment = self._find_comment_in_body(service_body, rpc_name)
                
                rpc = ProtoRPC(
                    name=rpc_name,
                    request_type=request_type,
                    response_type=response_type,
                    request_streaming=request_streaming,
                    response_streaming=response_streaming,
                    comment=rpc_comment
                )
                service.rpcs.append(rpc)
            
            services.append(service)
        
        return services
    
    def _extract_messages(self, lines: List[tuple], prefix: str = "") -> List[ProtoMessage]:
        """Extract message definitions."""
        messages = []
        content = '\n'.join([line for line, _ in lines])
        
        # Find all message blocks
        for message_match in re.finditer(
            r'message\s+(\w+)\s*\{([^}]*(?:\{[^}]*\}[^}]*)*)\}',
            content,
            re.DOTALL
        ):
            message_name = message_match.group(1)
            message_body = message_match.group(2)
            
            # Get comment for message
            message_comment = self._find_comment_for_construct(lines, message_name, 'message')
            
            full_name = f"{self.current_package}.{prefix}{message_name}" if self.current_package else f"{prefix}{message_name}"
            message = ProtoMessage(
                name=message_name,
                full_name=full_name,
                comment=message_comment
            )
            
            # Extract fields
            for field_match in re.finditer(
                r'(optional|required|repeated)?\s*(\w+)\s+(\w+)\s*=\s*(\d+)',
                message_body
            ):
                label = field_match.group(1)
                field_type = field_match.group(2)
                field_name = field_match.group(3)
                field_number = int(field_match.group(4))
                
                # Skip nested message/enum definitions
                if field_type in ['message', 'enum', 'service']:
                    continue
                
                # Get comment for field
                field_comment = self._find_comment_in_body(message_body, field_name)
                
                proto_field = ProtoField(
                    name=field_name,
                    type=field_type,
                    number=field_number,
                    label=label,
                    comment=field_comment
                )
                message.fields.append(proto_field)
            
            messages.append(message)
        
        return messages
    
    def _extract_enums(self, lines: List[tuple], prefix: str = "") -> List[ProtoEnum]:
        """Extract enum definitions."""
        enums = []
        content = '\n'.join([line for line, _ in lines])
        
        # Find all enum blocks
        for enum_match in re.finditer(
            r'enum\s+(\w+)\s*\{([^}]*)\}',
            content,
            re.DOTALL
        ):
            enum_name = enum_match.group(1)
            enum_body = enum_match.group(2)
            
            # Get comment for enum
            enum_comment = self._find_comment_for_construct(lines, enum_name, 'enum')
            
            full_name = f"{self.current_package}.{prefix}{enum_name}" if self.current_package else f"{prefix}{enum_name}"
            proto_enum = ProtoEnum(
                name=enum_name,
                full_name=full_name,
                comment=enum_comment
            )
            
            # Extract enum values
            for value_match in re.finditer(
                r'(\w+)\s*=\s*(\d+)',
                enum_body
            ):
                value_name = value_match.group(1)
                value_number = int(value_match.group(2))
                
                # Get comment for enum value
                value_comment = self._find_comment_in_body(enum_body, value_name)
                
                enum_value = ProtoField(
                    name=value_name,
                    type="enum_value",
                    number=value_number,
                    comment=value_comment
                )
                proto_enum.values.append(enum_value)
            
            enums.append(proto_enum)
        
        return enums
    
    def _find_comment_for_construct(self, lines: List[tuple], name: str, keyword: str) -> Optional[str]:
        """Find comment for a top-level construct (service, message, enum)."""
        content = '\n'.join([line for line, _ in lines])
        pattern = rf'{keyword}\s+{name}'
        match = re.search(pattern, content)
        
        if not match:
            return None
        
        # Find the line with this construct
        for i, (line, comment) in enumerate(lines):
            if keyword in line and name in line:
                # Check previous lines for comments
                comments = []
                j = i - 1
                while j >= 0 and (not lines[j][0] or lines[j][1]):
                    if lines[j][1]:
                        comments.insert(0, lines[j][1])
                    j -= 1
                
                # Also include inline comment
                if comment:
                    comments.append(comment)
                
                return ' '.join(comments) if comments else None
        
        return None
    
    def _find_comment_in_body(self, body: str, name: str) -> Optional[str]:
        """Find comment for a field/rpc/value within a body."""
        lines = body.split('\n')
        for i, line in enumerate(lines):
            if name in line:
                # Check for inline comment
                comment_match = re.search(r'//(.*)$', line)
                if comment_match:
                    return comment_match.group(1).strip()
                
                # Check previous line
                if i > 0:
                    prev_comment = re.search(r'//(.*)$', lines[i-1])
                    if prev_comment:
                        return prev_comment.group(1).strip()
        
        return None


def parse_proto_file(file_path: str) -> ProtoFile:
    """Convenience function to parse a proto file."""
    parser = ProtoParser()
    return parser.parse_file(file_path)

