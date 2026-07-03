# Stellaris UI Modding Rules

These rules apply specifically to this workspace, which is dedicated to modifying the Stellaris UI. Always keep these constraints and best practices in mind when modifying or creating UI files.

## Core Modding Constraints

1.  **NEVER DELETE DEFAULT ELEMENTS**: When modifying an existing vanilla UI window, do **not** delete any default elements. Doing so will crash the game's hardcoded UI engine.
    *   **How to hide elements**: If you want to remove an element, you must hide it by either setting its size to zero (`width = 0 height = 0`) or moving its coordinates far off-screen (e.g., `x = -1100 y = -1100`).

2.  **Creating Custom Windows**: Stellaris does not support creating entirely new UI windows natively. To create a new screen, use the **diplomatic event trick**:
    *   Trigger a diplomatic event.
    *   Declare `custom_gui = "your_custom_window_name"` in the event script.
    *   Override the default diplomatic window, hiding all standard diplomatic elements (like alien portraits) and drawing your custom UI on top.

## File Structure & Loading

3.  **File Separation**: Maintain the separation of concerns between structure and graphics.
    *   `.gui` files control layout, nesting, text, and structure.
    *   `.gfx` files define sprites and textures.
4.  **Loading Order**: Interface files are loaded alphabetically and follow the **LIOS** (Last In, Only Survivor) method. Keep this in mind when overriding vanilla files.
5.  **Variables**: Use `@variables` for simple values (like standard widths or heights) to maintain layout consistency. Remember that variables cannot traverse or ascend nested elements.

## Common Element Types

6.  Use the correct element types for your UI:
    *   `containerWindowType`: The foundational container for holding other elements.
    *   `buttonType`: Standard buttons tied to hardcoded game logic.
    *   `effectButtonType`: Buttons that trigger custom scripted effects (from `/common/button_effects/`) and display dynamic variables using bracket commands.
    *   `iconType`: Static graphics, backgrounds, and frames.
    *   `instantTextBoxType`: Text display (use localized string keys).
    *   `gridBoxType` & `smoothListboxType`: For dynamically formatting lists of items.

## Development Workflow

7.  When working on UI, remind the user about the following in-game console commands to speed up iteration without restarting the game:
    *   `reload [filename].gui`
    *   `reload texture all`
    *   `guibounds`
    *   `debugtooltip` (and using `CTRL+ALT+Right Click` on elements to open the `.gui` file)

## Workspace Layout

8.  **Project Root vs. Stellaris Root**: 
    *   The top-level `game-mod` directory is the main workspace, but Stellaris-specific tools are housed in the `stellaris` subdirectory.
    *   Stellaris tools (like `create_portrait_mod` and `install_mods`) should **always** be executed from within the `stellaris` directory.
    *   The tools write their output to a local `mod/` directory located at `game-mod/stellaris/mod/` (i.e. `./mod` relative to the `stellaris` directory). Do **not** use `../mod/` for outputs assuming the tool is run from a subfolder or from the top-level root.

## References

*   For comprehensive documentation, see the full guide at `docs/stellaris_ui_modding_guide.md`.
