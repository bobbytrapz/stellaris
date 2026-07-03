#!/bin/bash
cd "$(dirname "$0")"

echo "[*] Building tools..."

echo "[*] Compiling create_portrait_mod..."
go build -o create_portrait_mod generators/create_portrait_mod.go

echo "[*] Compiling portrait_gallery_generator..."
go build -o portrait_gallery_generator generators/portrait_gallery_generator.go

echo "[*] Compiling install_mods..."
go build -o install_mods generators/install_mods.go

echo "[+] Build complete!"
