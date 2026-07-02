# Stellaris Interface Modding Guide

This guide covers the core concepts, techniques, and elements used in Stellaris Interface (UI) Modding based on information from the Paradox Wikis.

## 1. The Core Files (`.gui` and `.gfx`)

Interface mods primarily live in the `Stellaris/interface/` folder and rely on two main file types:

*   **`.gui` files**: These dictate the structure, position, layout, and text of the UI windows. They use an object-oriented structure (nested elements).
*   **`.gfx` files**: These act as a library, pointing the game to your graphical assets (like `.dds` texture images or sprite sheets).

> [!NOTE]
> **Loading Order:** Stellaris loads interface files alphabetically, utilizing the LIOS (Last In, Only Survivor) method.
> **Variables:** Simple values can be defined (e.g., `@myvar = 200`) and used throughout `.gui` files to maintain layout consistency. However, variables cannot traverse or ascend nested elements.

## 2. Modifying Existing UI vs. Creating New UI

### Modifying Default UI
You can change coordinates, dimensions, fonts, and textures of existing elements. 

> [!WARNING]
> **You cannot delete default elements!** Doing so will crash the game's hardcoded UI engine. Instead, if you want to remove something, you have to "hide" it by setting its size to zero (`width = 0 height = 0`) or moving its coordinates far off-screen (e.g., `x = -1100 y = -1100`).

### Creating Brand New Screens (The "Custom GUI" Trick)
Stellaris doesn't natively support creating completely new UI windows from scratch via script. Modders bypass this by triggering a **diplomatic event**, declaring `custom_gui = "my_custom_screen"` in the event script, and then overriding the default diplomatic window. You hide all the standard diplomatic elements (like the alien portrait and ethics icons) and draw your custom buttons and graphics on the empty canvas.

## 3. GUI Element Types

UI layouts are built using specific element types. Some of the most common ones include:

*   **`containerWindowType`**: The foundational container for holding other elements, supporting scrollbars and specific orientations.
*   **`buttonType` & `effectButtonType`**: Clickable buttons. Standard buttons are tied to hardcoded game logic, but `effectButtonType` can trigger custom scripted effects (from `/common/button_effects/`) and display dynamic variables using bracket commands (e.g., `[MyVariable]`).
*   **`iconType`**: Used to display static graphics, backgrounds, and frames.
*   **`instantTextBoxType`**: Used for displaying text, which relies heavily on localized string files.
*   **Lists and Grids**: Containers like `gridBoxType` and `smoothListboxType` allow you to format multiple items dynamically.

## 4. Quality of Life Tools for Modding (Console Commands)

Stellaris provides some excellent console commands to help you iterate quickly without restarting the game:

*   `reload [filename].gui`: Reloads a specific UI layout file in real-time.
*   `reload texture all` (or `reload texture [name]`): Reloads sprite sheets and graphics without restarting the game.
*   `guibounds`: Displays visual boundaries (hitboxes/alignment markers) of UI elements on the screen to help you visualize margins and alignments.
*   `debugtooltip`: Shows extended debug info. Starting in 3.7, you can use `CTRL+ALT+Right Click` on an interface element to open/reload its corresponding GUI file.
