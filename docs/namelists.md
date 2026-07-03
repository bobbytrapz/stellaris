# Stellaris Namelist Modding Guide

This document provides a comprehensive overview of the keys and structure used in Stellaris namelist files (`common/name_lists/*.txt`).

## Top-Level Configuration Keys
These keys define the meta-properties of the namelist itself:
- `category` *(string)*: The UI category the namelist belongs to (e.g., "Humanoid", "Machine", "Aquatic").
- `alias` *(string)*: Alternative names by which the namelist can be referenced in scripts.
- `selectable` *(boolean)*: Determines if the namelist is available to the player in empire creation.
- `randomized` *(boolean)*: Determines if randomly generated AI empires can be assigned this namelist.
- `trigger` *(block)*: A script block containing conditions required for an empire to randomly pick this namelist.
- `customize_random_override` *(string)*: Redirects the random species/homeworld/home system generation to use the pools from a different namelist.
- `should_name_home_system_planets` *(boolean)*: Dictates whether non-home planets within the home system should pull from the `planet_names` pool during galaxy generation.

## 1. `ship_names = { ... }`
**Purpose**: Defines the pools of individual names assigned to ships when they are built or spawned. It is divided by ship size/type.
- `generic`: A fallback name pool used for any ship size.
- **Combat Ships**: `corvette`, `destroyer`, `cruiser`, `battleship`, `titan`, `colossus`, `juggernaut`
- **Utility & Civilian**: `science`, `constructor`, `colonizer`, `transport`, `sponsored_colonizer` (used for corporate/private prospector ships)
- **Defensive Stations**: `military_station_small`, `ion_cannon`

## 2. `ship_class_names = { ... }`
**Purpose**: Defines the names given to ship designs (classes) in the ship designer. If empty, the game falls back to `ship_names`.
- `generic`: General template names for any class.
- Can also accept specific ship sizes (e.g., `corvette`, `cruiser`) just like the `ship_names` block.

## 3. `fleet_names = { ... }`
**Purpose**: Defines naming formats for military armadas and fleets.
- `random_names`: A block containing explicitly predefined, unique fleet names (e.g., "Grand Armada", "First Fleet").
- `sequential_name`: A string template used for procedurally numbering fleets (e.g., "1st Strike Force", "2nd...").

## 4. `army_names = { ... }`
**Purpose**: Generates names for ground forces, mapped to specific army types. Within each type, `random_names` and `sequential_name` sub-keys determine the actual string.
- `generic`: Default naming format for any army type.
- **Standard**: `defense_army`, `assault_army`, `occupation_army`
- **Machine/Robotic**: `machine_defense`, `machine_assault_1`, `machine_assault_2`, `machine_assault_3`, `robotic_army`, `robotic_defense_army`, `individual_machine_occupation_army`, `robotic_occupation_army`
- **Specialized Biological**: `slave_army`, `clone_army`, `perfected_clone_army`, `undead_army`, `psionic_army`, `xenomorph_army`, `gene_warrior_army`
- **Primitive/Pre-FTL**: `primitive_army`, `industrial_army`, `postatomic_army`, `wilderness_pre_sapient_defence_army`, `wilderness_pre_sapient_assault_army`
- **Event/Crisis**: `warpling_army`, `ember_legion`, `lm_imbued_army`, `abomination_army`

## 5. `planet_names = { ... }`
**Purpose**: Used to assign names to colonized or discovered planets.
- `generic`: A fallback pool (inside a nested `names = { ... }` block) for any planet type.
- **Specific Planet Classes**: `pc_desert`, `pc_tropical`, `pc_arid`, `pc_continental`, `pc_ocean`, `pc_tundra`, `pc_arctic`, `pc_savannah`, `pc_alpine`. Each of these contains a nested `names = { ... }` block holding the actual strings.

## 6. `character_names = { ... }`
**Purpose**: Generates names for species leaders (Rulers, Governors, Scientists, Admirals, Generals). They are usually nested inside a culture block (e.g., `default = { ... }`) to allow for distinct naming groups within one species.
- **Standard Names**: `full_names`, `full_names_male`, `full_names_female` (Used when the name is a single, complete string).
- **Combined Names**: `first_names`, `first_names_male`, `first_names_female` combined with `second_names`, `second_names_male`, `second_names_female` to form "First Last" formats.
- **Regnal Names**: `regnal_full_names`, `regnal_first_names`, `regnal_second_names` (and their male/female variants). These replace regular names for Rulers in imperial/dictatorial governments.
- **Rules & Formatting**: 
  - `weight`: Determines the chance this culture block is chosen if multiple exist.
  - `use_full_regnal_name`, `use_full_regnal_name_male`, `use_full_regnal_name_female` (booleans): If set to false, some localization interfaces will only display the ruler's first name.
