#!/bin/bash
cd "$(dirname "$0")"

echo "[*] Building tools..."

echo "[*] Compiling create_portrait_mod..."
go build -o create_portrait_mod generators/create_portrait_mod.go

echo "[*] Compiling portrait_gallery_generator..."
go build -o portrait_gallery_generator generators/portrait_gallery_generator.go

echo "[*] Compiling install_mods..."
go build -o install_mods generators/install_mods.go

echo "[*] Compiling create_namelist_mod..."
go build -o create_namelist_mod generators/create_namelist_mod.go

echo "[*] Compiling create_empire_mod..."
go build -o create_empire_mod generators/create_empire_mod.go

echo "[*] Compiling loc_editor..."
go build -tags sqlite_fts5 -o loc_editor generators/cmd/loc_editor/main.go

echo "[+] Build complete!"
