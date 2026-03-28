# Polyserver Console Reference

The interactive console is available whenever the server is running in launcher mode. Type commands directly into the terminal. Type `help` or `?` at any time to print a summary.

---

## `status`

Display a snapshot of the server's current state.

```
status
```

**Output includes:** invite code, invite key (if set), time until invite expires, current track directory and name, game mode, max players, whether a session switch is in progress, and auto-rotate state.

---

## `stats`

Display runtime performance metrics.

```
stats
```

**Output includes:** goroutine count, heap memory allocation (MB), total bytes sent and received (MB), and the last tick processing time (µs).

---

## `players` / `list`

List all currently connected players.

```
players
list
```

**Output columns:**

| Column | Description |
|---|---|
| ID | Numeric player ID used for `kick` |
| Name | Player's in-game nickname |
| Best Time | Fastest finish time this session, or `-` if not finished |
| Ping | Last measured round-trip latency in milliseconds |

---

## `kick`

Disconnect a player from the server.

```
kick <id>
```

**Arguments:**

| Argument | Type | Description |
|---|---|---|
| `id` | integer | The numeric player ID shown in `players` output |

**Example:**
```
kick 3
```

---

## `tracks`

List all tracks loaded from the tracks directory, grouped by folder. The currently active track is marked with `*`.

```
tracks
```

Use the displayed directory and name values with the `track` command.

---

## `track`

Switch the current track. The game mode and max players are preserved from the current session.

```
track <dir> <name>
```

**Arguments:**

| Argument | Type | Description |
|---|---|---|
| `dir` | string | The folder/category name shown in `tracks` output |
| `name` | string | The track name shown under that folder in `tracks` output |

**Example:**
```
track official Coastal Sprint
```

> **Note:** If the track name contains spaces, write it as-is — everything after `<dir>` is treated as the track name.

---

## `session`

Control the current race session lifecycle and configuration.

```
session end
session start
session set <gamemode> <dir> <name> <maxPlayers>
```

### `session end`

End the current session and put all clients into a "waiting" state. Players remain connected but the race stops.

### `session start`

Resume or start the session after it has been ended. Players will receive new session data and can begin racing.

### `session set`

Reconfigure the session completely — changing track, game mode, or player cap — and switch into it immediately.

```
session set <gamemode> <dir> <name> <maxPlayers>
```

**Arguments:**

| Argument | Type | Options / Range | Description |
|---|---|---|---|
| `gamemode` | integer | `0` = Casual, `1` = Competitive | Race scoring mode |
| `dir` | string | Any folder name from `tracks` | Track directory/category |
| `name` | string | Any track name from `tracks` | Track name within the directory |
| `maxPlayers` | integer | `1` – `200` | Maximum simultaneous connections |

**Example:**
```
session set 1 official Alpine Loop 32
```

---

## `invite`

View or regenerate the server's invite code.

```
invite
invite new
invite key <key>
```

### `invite`

Print the current invite code and its remaining validity time.

### `invite new`

Regenerate the invite code, preserving the current invite key (if one was set). Useful when the old code has expired.

### `invite key`

Create a new invite with a specific human-readable key. The key is appended to the invite code so players can join by typing the key instead of the full code.

```
invite key <key>
```

| Argument | Type | Description |
|---|---|---|
| `key` | string | Short alphanumeric label to attach to the invite (e.g. `race2`, `friends`) |

**Example:**
```
invite key friday
```

---

## `autorotate` / `ar`

Control automatic playlist rotation through tracks in a folder.

```
autorotate start <folder> <secs>
autorotate stop
autorotate skip
```

Both `autorotate` and `ar` are accepted.

### `autorotate start`

Begin cycling through every track in the given folder, switching at a fixed interval.

```
autorotate start <folder> <secs>
```

| Argument | Type | Description |
|---|---|---|
| `folder` | string | Track directory/category name (must match a folder shown in `tracks`) |
| `secs` | integer | Number of seconds each track plays before rotating to the next |

The rotation starts from the first track in the folder and advances in order, wrapping around at the end.

**Example:**
```
autorotate start official 300
```

### `autorotate stop`

Stop the playlist. The current track remains active; no further automatic rotation occurs.

### `autorotate skip`

Immediately advance to the next track in the rotation without waiting for the timer.

---

## `reload`

Reload all track files from the tracks directory on disk. Useful after adding or modifying tracks without restarting the server.

```
reload
```

---

## `stop`

Kill the game server subprocess. The launcher and dashboard remain running, so you can restart the server later via the dashboard's start button.

```
stop
```

---

## `exit` / `quit`

Kill the game server subprocess and then terminate the launcher process entirely. This shuts everything down.

```
exit
quit
```

---

## `help` / `?`

Print a compact command summary in the terminal.

```
help
?
```

---

## Quick Reference

```
status
stats
players
kick <id>
tracks
track <dir> <name>
session end
session start
session set <gamemode> <dir> <name> <maxPlayers>
invite
invite new
invite key <key>
autorotate start <folder> <secs>
autorotate stop
autorotate skip
reload
stop
exit / quit
help
```

---

## Gamemode Values

| Value | Name | Description |
|---|---|---|
| `0` | Casual | No competitive scoring; relaxed session |
| `1` | Competitive | Ranked session with finish-time tracking |
