# Configuration

> Learn how `bonjour` reads, manages, and resolves **user settings/configurations**.

## Configuration Paths

By default, `bonjour` reads and persists configuration settings at OS-specific paths, thanks to the functionality provided by Go's `os` library:

| Operating System | Default Config File Path |
| :--- | :--- |
| **Linux** | `~/.config/bonjour/config.json` |
| **macOS** | `~/Library/Application Support/bonjour/config.json` |
| **Windows** | `%APPDATA%\bonjour\config.json` |


## Configuration Precedence

When multiple configuration sources exist, `bonjour` resolves values in the following order (highest priority first):

1. **CLI Flags** (e.g., `bonjour` will give more importance to the `--noweather` CLI flag than to the `show_weather` value in the `config.json` file)
2. **Config File** (`config.json`)

## `config.json` Schema Reference

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| **`user`** | `object` | `{}` | User identity and geolocation configuration object. |
| `user.name` | `string` | `""` | Full name or display name of the user. |
| `user.latitude` | `float` | `0.0` | Geographic latitude coordinates for location-based services. |
| `user.longitude` | `float` | `0.0` | Geographic longitude coordinates for location-based services. |
| `user.city` | `string` | `""` | Primary city name for weather lookups. |
| **`dashboard`** | `object` | `{}` | Dashboard user interface preference settings. |
| `dashboard.theme` | `string` | `"charm"` | Visual theme applied to the UI (e.g., `charm`). |
| `dashboard.show_weather` | `boolean` | `false` | Toggles whether weather information is rendered on the dashboard. |
| `dashboard.show_quotes` | `boolean` | `true` | Toggles whether daily quotes are displayed. |
| `dashboard.show_events` | `boolean` | `true` | Toggles whether upcoming calendar events are displayed. |
| **`events`** | `array` | `[]` | List of scheduled calendar event objects. |
| `events[].date` | `string` | `""` | Event target date formatted as `YYYY-MM-DD`. |
| `events[].title` | `string` | `""` | Short name or title describing the event. |
| `events[].warningstart` | `string` | `""` | Optional advance notice date for event warnings. |

## Creation and editing

Upon first running the application, the `config.json` file does not exist. To create it, the application launches a wizard that helps users accross configuration choices to set up their preferences. These are transformed into JSON format and inserted into a newly created `config.json`.

If the user runs the application with the `config` command, then a different wizard will allow them to individually change their name, their location information (used for weather services), and their dashboard configuration (which includes theme and module choices). All changes will be reflected into `config.json`.

## Examples

Examples of a complete and minimal configuration file can be found in the `/examples` folder of this repository.
