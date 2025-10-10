# -*- mode: python ; coding: utf-8 -*-

block_cipher = None

a = Analysis(
    ['mcp_proto_server.py'],
    pathex=[],
    binaries=[],
    datas=[
        ('examples', 'examples'),  # Include example proto files
    ],
    hiddenimports=[
        'mcp',
        'mcp.server',
        'mcp.server.stdio',
        'mcp.types',
        'protobuf',
        'rapidfuzz',
        'watchdog',
        'proto_indexer',
        'proto_parser',
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=block_cipher,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=block_cipher)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.zipfiles,
    a.datas,
    [],
    name='mcp-proto-server',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    # macOS code signing - will use ad-hoc signature if no identity provided
    codesign_identity=None,  # Set to your Developer ID for distribution
    entitlements_file='entitlements.plist',  # macOS entitlements for permissions
)

