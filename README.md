# Stellaris Personal Modding Tools

A collection of automated tools designed to streamline modding Stellaris, specifically focusing on injecting custom portraits using the achievement-compatible "Non-Intrusive Replacer" method.

## Features

*   **Portrait Mod Generator (`create_portrait_mod`)**: Automatically resizes and converts standard PNG images into DXT5 DDS textures and injects them into a new Mod using safe overwrite triggers.
*   **Portrait Gallery Generator (`portrait_gallery_generator`)**: Scans your local Stellaris installation and builds a visual HTML gallery of every single vanilla portrait in the game to help you easily identify which portrait archetypes you want to overwrite.
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

### 2. Generate the Portrait Gallery (Optional)
If you want to view all vanilla portraits to see what you can replace:
```bash
./portrait_gallery_generator
```
*Open `portrait_gallery/index.html` in your browser to view the generated gallery.*

### 3. Create a Custom Portrait Mod
Grab a `.png` image you want to use for your new species and run the portrait generator. You must provide the image, the vanilla archetype you want to inherit mechanics from (as a subcommand, e.g., `machine`), and the exact species name you will use in-game to trigger the portrait.

```bash
./create_portrait_mod machine --image my_robot.png --name "Awesomebots"
```
*This automatically creates a new mod called `awesomebots_portrait` inside the `mod/` directory.*

**Valid Archetypes:** `machine`, `humanoid`, `mammalian`, `avian`, `reptilian`, `fungoid`, `plantoid`, `lithoid`, `aquatic`, `toxoid`.

### 4. Install Your Mods
Run the installer script to deploy all mods located in the `mod/` folder to your actual Stellaris user directory:
```bash
./install_mods
```

## How It Works

The `create_portrait_mod` tool uses the **Non-Intrusive Replacer** method. It safely overwrites a vanilla portrait group without touching the game's core `00_species_classes.txt` file (meaning it doesn't break checksums/achievements). 

It injects a custom trigger so that if an empire's species name perfectly matches the name you provided (e.g. "Awesomebots"), it uses your custom image. Otherwise, it falls back to the default vanilla portrait for all AI empires.
