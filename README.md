# Stellaris Personal Modding Tools

A collection of automated tools designed to streamline modding Stellaris, specifically focusing on injecting custom portraits using the clean, native "Portrait Sets" architecture.

## Features

*   **Portrait Mod Generator (`create_portrait_mod`)**: Automatically resizes and converts standard PNG images into DXT5 DDS textures and injects them into a new Mod using safe overwrite triggers.
*   **Portrait Gallery Generator (`portrait_gallery_generator`)**: Scans your local Stellaris installation and builds a visual HTML gallery of every single vanilla portrait in the game.
*   **Namelist Mod Generator (`create_namelist_mod`)**: Compiles TSV spreadsheets of names into valid Stellaris namelists.
*   **Empire Mod Generator (`create_empire_mod`)**: Provides a beautiful, interactive Web UI to forge custom empires, seamlessly integrating custom portraits and namelists into prescripted empires.
*   **Mod Installer (`install_mods`)**: Safely deploys your generated mods into your local Stellaris `mod` directory so they appear in the Paradox Launcher.

## Prerequisites

*   [Go](https://golang.org/) (for compiling the tools)
*   [ImageMagick](https://imagemagick.org/) (specifically the `convert` or `magick` command) for automatic image conversion.

## Quickstart

### 1. Build the Tools
Run the included build script to compile the tools into standalone binaries in the root directory:
```bash
./build.sh
```

### 2. Create a Custom Portrait Mod
Grab a `.png` image you want to use for your new species and run the portrait generator. You must provide the image, the vanilla archetype you want to inherit mechanics from (as a subcommand, e.g., `machine`), and the exact species name you will use in-game to trigger the portrait.

```bash
./create_portrait_mod machine --image my_robot.png --name "Awesomebots"
```
*This automatically creates a new mod called `awesomebots_portrait` inside the `mod/` directory.*

### 3. Forge a Prescripted Empire
If you want to spawn your custom race in-game, you can generate a prescripted empire.
The empire generator uses a configuration file, `empire.yaml`. The easiest way to configure it is to use the interactive web UI:

```bash
./create_empire_mod edit
```

This will launch a local dashboard at `http://localhost:44793`. From the UI you can visually pick room backgrounds, design your empire's flag, assign ethics and civics from the live game data, and seamlessly upload custom portrait and namelist files.

When you're ready, generate the final empire mod:
```bash
./create_empire_mod generate empire.yaml
```

### 4. Install Your Mods
Run the installer script to deploy all mods located in the `mod/` folder to your actual Stellaris user directory:
```bash
./install_mods
```
