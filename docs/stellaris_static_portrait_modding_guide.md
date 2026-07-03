# Stellaris Static Portrait Modding Guide

This guide covers the core concepts and techniques for creating **Static Portraits** in Stellaris, as well as the **Non-intrusive Replacer** method for achievement-compatible mods. This information is based on the Paradox Wikis.

## 1. What are Static Portraits?

Unlike animated portraits which require complex 3D meshes and Maya rigging, static portraits are simple 2D images. They do not animate or breathe, but they are significantly easier to create and implement.

*   **File Format:** The image must be a transparent `.dds` file. The character should be cut out cleanly with no background.
*   **File Location:** The `.dds` file should be saved somewhere within the `\gfx` folder of your mod (e.g., `gfx/interface/portraits/`).

## 2. Defining a Static Portrait

To tell the game about your new static portrait, you must define it in a text file.

*   **Definition File:** Portraits are defined in `\gfx\portraits\portraits\00_portraits.txt` (or a custom named `.txt` file in that directory, like `01_my_custom_portraits.txt`).

### Code Syntax

You can define a static portrait in two ways:

**Method A: Using `spriteType`** (Requires defining the sprite elsewhere in a `.gfx` file)
```txt
portraits = {
    my_custom_portrait = {
        spriteType = "GFX_portrait_my_custom_portrait"
    }
}
```

**Method B: Directly referencing the texture file** (Simpler, all in one file)
```txt
portraits = {
    my_custom_portrait = {
        texturefile = "gfx/interface/portraits/my_custom_portrait.dds"
    }
}
```

> [!NOTE]
> Once defined, the portrait identifier (e.g., `my_custom_portrait`) can be referenced in other files, such as species definitions or event scripts.

---

## 3. The Non-intrusive Replacer Method

Normally, modifying files in the `common/` folder changes the game's checksum, which disables Achievements. The **Non-intrusive Replacer** method allows you to replace vanilla portraits *without* altering the checksum. 

This method keeps both the unmodded and modded portraits in the game. The game will display the modded portrait only when specific conditions (like a species name or an empire flag) are met.

> [!IMPORTANT]
> **Rule of Thumb:** Do NOT modify anything in the `common` folder for this method to work.

### Step-by-Step Implementation

1.  **Identify the Target:** Pick a vanilla portrait group you want to replace (e.g., `humanoid_02`).
2.  **Copy the Vanilla File:** Locate the corresponding vanilla portrait file (e.g., `\gfx\portraits\portraits\09_portraits_humanoid.txt`) and copy it into your mod's `\gfx\portraits\portraits\` folder. Keep the exact same filename.
3.  **Comment Out the Vanilla Block:** Open your copied file and comment out the *entire block* for the group you want to replace by adding a `#` at the start of every line within that group's definition.
4.  **Define Your Triggers:** In your own custom portrait definition file, recreate the block for that group (e.g., `humanoid_02`) but add `trigger` blocks to conditionally apply your custom portrait.

### Syntax Example: Triggering by Species Name

In this example, the custom portrait `mol14` replaces the vanilla `batarian_male` ONLY if the species name is exactly "Batarian".

```txt
portrait_groups = {
    mol14 = {
        default = batarian_male
        
        game_setup = {
            # Runs in the empire designer
            add = {
                trigger = {
                    ruler = { gender = male }
                }
                portraits = { batarian_male mol14 }
            }
        }
        
        species = {
            # Generic portrait for a species in the species screen
            add = {
                trigger = { is_species = "Batarian" }
                portraits = { batarian_male batarian_female }
            }
            add = {
                trigger = { NOT = { is_species = "Batarian" } }
                portraits = { mol14 }
            }
        }
        
        pop = {
            # For specific pops (similar triggers to species)
        }
        
        leader = {
            # Scientists, generals, admirals, governors
            add = {
                trigger = {
                    gender = female
                    is_species = "Batarian"
                }
                portraits = { batarian_female }
            }
            add = {
                trigger = { NOT = { is_species = "Batarian" } }
                portraits = { mol14 }
            }
        }
        
        ruler = {
            # Ruler scope (similar triggers to leader)
        }
    }
}
```

### Syntax Example: Triggering by Empire Flag

If you want the portrait to apply to a specific empire regardless of what the player names the species, you can check for a country flag instead.

```txt
trigger = {
    owner = {
        has_country_flag = "my_custom_empire_flag"
    }
    NOT = {
        is_species = "Husk"
    }
}
```

> [!TIP]
> This requires the prescripted empire definition to include the flag:
> ```txt
> flags = {
>     "my_custom_empire_flag"
> }
> ```
> Any empire created and saved from this prescripted design will inherit the flag, and the portrait replacement will automatically trigger.
