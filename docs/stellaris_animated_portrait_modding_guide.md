# Stellaris Animated Portrait Modding Guide

This guide covers the process of creating animated 2D portraits for Stellaris, from texture preparation in Photoshop to rigging/animating in Autodesk Maya, and finally configuring the required files for the game.

## 1. Asset Preparation in Photoshop
Rather than utilizing traditional 3D models, Stellaris animated portraits are made of multiple 2D planes layered and animated in a 3D skeleton.
* **Software Required**: Photoshop (or equivalent) & Autodesk Maya (ideally 2015 or 2016 for compatibility with the official exporter).
* **Format**: `.dds` (DXT5 settings, interpolating alpha).

### Workflow:
1. **Layer Separation**: Paint your alien/character. Place moving parts (arms, tentacles, ears, jaw) onto individual layers. Fill/paint in the background behind those moving parts to avoid transparency gaps when they move.
2. **Spacing**: Arrange all separated layers on a single texture map, leaving at least a few pixels of empty space between them.
3. **Padding (Alpha Bleeding Prevention)**:
   * Duplicate all separated layers and merge the duplicates together.
   * Place the merged layer at the bottom of your layer stack and name it `padding`.
   * Ctrl-click the padding layer, expand the selection by ~5 pixels (`Select > Modify > Expand`), and fill the expanded region with the colors of the body parts (using standard brush strokes or a plugin like Flaming Pear's "Solidify A"). This prevents black lines or artifact borders (alpha bleeding) around your assets when scaled in-game.
4. **Alpha Channel**:
   * Ctrl+Shift-click all your active layer thumbnails to select their opacity.
   * Go to the **Channels** tab and click **Save selection as channel** to create the alpha channel.
5. **DDS Export**:
   * Save your working `.psd` file.
   * Resize the texture to `512x512` pixels (do NOT overwrite your primary `.psd` at this size).
   * Save as a `.dds` using **DXT5** compression settings.
   * Undo (`Ctrl+Z`) in Photoshop to restore your working file's original size.

---

## 2. Maya Setup and Rigging
1. **Joint Hierarchy**: Create a standard joint system. The base skeleton hierarchy should be:
   `root` -> `root_center` -> `spine_1` -> `spine_2` ...
2. **Material Setup**:
   * Check your `.dds` file's dimensions in centimeters (e.g., 512x512 px is approx. 18.06 x 18.06 cm).
   * In Maya, create a polygon plane matching those centimeter dimensions (`Create > Polygon Primitives > Plane > Options`, set Axis to Z).
   * Open the **Hypershade** window and create a **Phong** material:
     * Assign your portrait texture to the **Diffuse** channel.
     * Assign empty black textures to the **Normal** and **Specular** channels.
     * Assign this material to your plane.
3. **Mesh Layout**:
   * Duplicate the plane for each individual part of your portrait.
   * Add divisions/vertices and shape each plane to perfectly enclose its corresponding body part. Ensure the **Preserve UVs** setting in the Move Tool is checked so modifying the geometry does not distort the texture.
   * Position the planes slightly offset from each other on the Z-axis to order them correctly (front-to-back layer depth).
   * **Rendering Sort Order**: Sort the meshes in the Maya Outliner. The back-most layer must be at the very top of the list, and the front-most layer must be at the very bottom.
4. **Skinning**:
   * Select your joints and meshes.
   * Go to `Skin > Bind Skin > Options`. Set Bind to: **Joint hierarchy** and Max Influences: **4** (uncheck "Remove unused influences").
   * Paint skin weights to define how the planes deform when joints are animated.

---

## 3. Animation
* **Setting up the Animation Attribute**:
  1. Select the `root` joint.
  2. Open the Maya Exporter window, set the project to **Stellaris**, and click **Add animation attr.**.
  3. This adds an animation attribute (visible in the Channel Box) which allows you to define states like `"none"`, `"idle"`, etc.
  4. Ensure you have at least `"none"` and one active state (e.g. `"idle"`). If `"none"` is missing, go to `Modify > Edit Attribute`, select the animation attribute, and add `"none"` under **Enum Names**.
* **Keyframing**:
  * Set a key on the animation attribute of the `root` joint to your animation name (e.g. `"idle"`) at the start frame, and set it to `"none"` at the end frame. 
  * Each step on the attribute graph represents a specific animation state. The 0 value represents `"none"`.

---

## 4. Shaders and Export
* **Shaders**:
  * Use the `"PdxMeshPortrait"` shader for the main body.
  * In the Hypershade, select the shader. In the exporter window, click **Add shader attr.** to append the Paradox shader type to the Maya material.
  * For clothes, use `"PdxMeshPortraitClothes"`. For hair, use `"PdxMeshPortraitHair"`.
* **Exporting**:
  * Open the Clausewitz / Jorodox exporter window.
  * **IMPORTANT**: Check **Skip Merge** under general options. If unchecked, the exporter will merge all planes into a single mesh, destroying your layered rendering hierarchy.
  * Select your export path and file name, then export.

---

## 5. In-Game Implementation

### A. Asset Configuration (`.asset` file)
Define the animation mappings and the entity itself in a `.asset` file located under `gfx/models/portraits/[species_group]/`.

```txt
# Define individual animations
animation = {
    name = "my_portrait_idle_animation"
    file = "my_portrait_idle.anim"
}
animation = {
    name = "my_portrait_happy_animation"
    file = "my_portrait_happy.anim"
}

# Define the entity
entity = {
    name = "portrait_my_alien_entity"
    pdxmesh = "portrait_my_alien_mesh"
    default_state = "idle"
    
    # State mapping (allows random selection/blend times)
    state = {
        name = "idle"
        animation = "idle"
        animation_blend_time = 0
        chance = 2.0
        looping = no
        next_state = idle
    }
    state = {
        name = "idle"
        animation = "idle2"
        animation_blend_time = 0
        chance = 1.0
        looping = no
        next_state = idle
    }
    
    scale = 1.0
}
```

### B. Mesh Definition (`.gfx` or `.asset` file)
Link the exported mesh file and associate it with the defined animation names.

```txt
objectTypes = {
    pdxmesh = {
        name = "portrait_my_alien_mesh"
        file = "gfx/models/portraits/[species]/my_portrait.mesh"
        animation = { id = "idle" type = "my_portrait_idle_animation" }
        animation = { id = "idle2" type = "my_portrait_happy_animation" }
        scale = 1.0
    }
}
```

### C. Portrait Group Configuration (`gfx/portraits/portraits/` directory)
Define the portrait entries and associate them with textures and selectors:

```txt
portraits = {
    my_alien_portrait = {
        entity = "portrait_my_alien_entity"
        clothes_selector = "no_texture" # Or path to a clothes asset selector txt
        hair_selector = "no_texture"
        greeting_sound = "molluscoid_01_greetings"
        character_textures = {
            "gfx/models/portraits/[species]/my_alien_texture.dds"
        }
    }
}
```

### D. Species Classes Configuration (`common/species_classes/00_species_classes.txt`)
Add your new portrait key to the relevant phenotype so it shows up in the empire designer:

```txt
FUN = {
    portraits = {
        "fun1"
        "fun2"
        "my_alien_portrait"
    }
    graphical_culture = fungoid_01
    move_pop_sound_effect = "fungoid_pops_move"
}
```

---

## 6. Non-Intrusive Replacer Method (Alternative)
If you want to replace an existing vanilla portrait group (e.g., Humanoid 2) without changing the checksum (allowing achievements) and letting the modded and vanilla portraits coexist:

1. **Skip the common folder**: Do not edit `common/species_classes/00_species_classes.txt`.
2. **Comment out Vanilla Portrait**: Copy the vanilla portrait file containing the group you want to replace (e.g., `gfx/portraits/portraits/09_portraits_humanoid.txt`) to your mod folder, and comment out the target portrait group block (e.g. `# humanoid_02 = { ... }`).
3. **Add Triggers**: In your custom portrait file, define your portrait groups using species or country scope triggers:

```txt
portrait_groups = {
    my_replacement_group = {
        default = batarian_male
        
        # Shows in game setup
        game_setup = {
            add = {
                trigger = { ruler = { gender = male } }
                portraits = { batarian_male my_alien_portrait }
            }
        }
        
        # Triggers in-game based on Species name
        species = {
            add = {
                trigger = { is_species = "Batarian" }
                portraits = { batarian_male batarian_female }
            }
            add = {
                trigger = { NOT = { is_species = "Batarian" } }
                portraits = { my_alien_portrait }
            }
        }
        
        # Triggers in-game based on Country Flags
        leader = {
            add = {
                trigger = {
                    owner = { has_country_flag = "my_custom_flag" }
                    NOT = { is_species = "Husk" }
                }
                portraits = { my_alien_portrait }
            }
        }
    }
}
```
