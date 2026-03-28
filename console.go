package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type consoleClient struct {
	base string
	cmd  *exec.Cmd
}

// startConsole blocks and reads commands from stdin, dispatching them to the
// control API. It should be called in the launcher (not the server subprocess).
func startConsole(controlPort int, controlAddress string, cmd *exec.Cmd) {
	// Give the server subprocess a moment to start up before we print the banner
	time.Sleep(2 * time.Second)

	c := &consoleClient{
		base: fmt.Sprintf("http://%s:%d", controlAddress, controlPort),
		cmd:  cmd,
	}

	fmt.Println()
	fmt.Println("=========================================")
	fmt.Println("  Polyserver console ready — type 'help'")
	fmt.Println("=========================================")
	fmt.Println()
	printPrompt()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			c.dispatch(line)
		}
		printPrompt()
	}
}

func printPrompt() {
	fmt.Print("> ")
}

func (c *consoleClient) dispatch(line string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return
	}
	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	switch cmd {
	case "help", "?":
		printHelp()
	case "status":
		c.cmdStatus()
	case "players", "list", "who":
		c.cmdPlayers()
	case "kick":
		c.cmdKick(args)
	case "tracks":
		c.cmdTracks()
	case "track":
		c.cmdTrack(args)
	case "session":
		c.cmdSession(args)
	case "invite":
		c.cmdInvite(args)
	case "stats":
		c.cmdStats()
	case "autorotate", "ar":
		c.cmdAutoRotate(args)
	case "reload":
		c.cmdReload()
	case "stop":
		c.cmdStop()
	case "exit", "quit":
		c.cmdExit()
	default:
		fmt.Printf("Unknown command %q — type 'help' for available commands\n", cmd)
	}
}

// ---------- HTTP helpers ----------

func (c *consoleClient) get(path string) (map[string]interface{}, error) {
	resp, err := http.Get(c.base + path)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse error: %w (body: %s)", err, body)
	}
	return result, nil
}

func (c *consoleClient) post(path string, payload interface{}) (int, map[string]interface{}, error) {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader([]byte("{}"))
	}
	req, err := http.NewRequest("POST", c.base+path, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		return resp.StatusCode, nil, nil
	}
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return resp.StatusCode, result, nil
}

// ---------- Individual commands ----------

func printHelp() {
	fmt.Print(`
Commands:
  status                                  Server status, invite code, current track
  stats                                   Runtime stats (memory, bandwidth, tick time)
  players / list                          List connected players
  kick <id>                               Kick a player by their numeric ID
  tracks                                  List all available tracks (* = current)
  track <dir> <name>                      Switch track (preserves gamemode/maxplayers)
  session end                             End the current race session
  session start                           Start / resume the current session
  session set <gamemode> <dir> <name> <maxPlayers>
                                          Full session reconfigure
  invite                                  Show current invite code
  invite new                              Regenerate invite (keep same key)
  invite key <key>                        Create invite with a specific key
  autorotate start <folder> <secs>        Start playlist auto-rotation
  autorotate stop                         Stop auto-rotation
  autorotate skip                         Skip to next track immediately
  reload                                  Reload track files from disk
  stop                                    Kill the game server subprocess
  exit / quit                             Kill the game server and exit the launcher
  help / ?                                Show this message

Gamemodes: 0 = Casual  1 = Competitive

`)
}

func (c *consoleClient) cmdStatus() {
	data, err := c.get("/status")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println()
	fmt.Printf("  Invite Code  : %v\n", data["invite"])
	if key := data["inviteKey"]; key != nil {
		fmt.Printf("  Invite Key   : %v\n", key)
	}
	fmt.Printf("  Expires In   : %v\n", data["timeoutIn"])
	fmt.Printf("  Current Track: %v / %v\n", data["currentDir"], data["current"])

	if sess, ok := data["session"].(string); ok {
		var s map[string]interface{}
		if json.Unmarshal([]byte(sess), &s) == nil {
			gm := "Casual"
			if gmVal, ok := s["GameMode"].(float64); ok && int(gmVal) == 1 {
				gm = "Competitive"
			}
			fmt.Printf("  Game Mode    : %s\n", gm)
			fmt.Printf("  Max Players  : %v\n", s["MaxPlayers"])
			fmt.Printf("  Switching    : %v\n", s["SwitchingSession"])
		}
	}

	if ar, ok := data["autorotate"].(map[string]interface{}); ok {
		if ar["enabled"] == true {
			fmt.Printf("  Auto-Rotate  : enabled — folder: %v, interval: %vs, state: %v, next in: %vs\n",
				ar["folder"], ar["interval"], ar["state"], ar["timeLeft"])
		} else {
			fmt.Printf("  Auto-Rotate  : disabled\n")
		}
	}
	fmt.Println()
}

func (c *consoleClient) cmdStats() {
	data, err := c.get("/status")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	stats, ok := data["stats"].(map[string]interface{})
	if !ok {
		fmt.Println("No stats available.")
		return
	}
	fmt.Println()
	fmt.Printf("  Goroutines  : %v\n", stats["goroutines"])
	if mem, ok := stats["memoryAlloc"].(float64); ok {
		fmt.Printf("  Memory      : %.2f MB\n", mem/1024/1024)
	}
	if sent, ok := stats["bytesSent"].(float64); ok {
		fmt.Printf("  Bytes Sent  : %.2f MB\n", sent/1024/1024)
	}
	if recv, ok := stats["bytesReceived"].(float64); ok {
		fmt.Printf("  Bytes Recv  : %.2f MB\n", recv/1024/1024)
	}
	if tick, ok := stats["tickTime"].(float64); ok {
		fmt.Printf("  Tick Time   : %dµs\n", int64(tick)/1000)
	}
	fmt.Println()
}

func (c *consoleClient) cmdPlayers() {
	data, err := c.get("/players")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	players, ok := data["players"].([]interface{})
	if !ok || len(players) == 0 {
		fmt.Println("No players connected.")
		return
	}
	fmt.Println()
	fmt.Printf("  %-6s  %-24s  %-12s  %s\n", "ID", "Name", "Best Time", "Ping")
	fmt.Println("  " + strings.Repeat("-", 54))
	for _, p := range players {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		id := fmt.Sprintf("%.0f", pm["id"].(float64))
		name := fmt.Sprintf("%v", pm["name"])
		t := fmt.Sprintf("%v", pm["time"])
		ping := fmt.Sprintf("%.0fms", pm["ping"].(float64))
		fmt.Printf("  %-6s  %-24s  %-12s  %s\n", id, name, t, ping)
	}
	fmt.Println()
}

func (c *consoleClient) cmdKick(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: kick <player_id>")
		fmt.Println("  Run 'players' to see IDs.")
		return
	}
	id, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		fmt.Println("Invalid player ID:", args[0])
		return
	}
	status, _, err := c.post("/kick", map[string]interface{}{"id": uint32(id)})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if status == 204 {
		fmt.Printf("Player %d kicked.\n", id)
	} else {
		fmt.Printf("Kick returned status %d.\n", status)
	}
}

func (c *consoleClient) cmdTracks() {
	data, err := c.get("/status")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	currentDir, _ := data["currentDir"].(string)
	currentName, _ := data["current"].(string)

	tracks, ok := data["tracks"].(map[string]interface{})
	if !ok {
		fmt.Println("No track data available.")
		return
	}
	fmt.Println()
	for dir, names := range tracks {
		fmt.Printf("  [%s]\n", dir)
		nameList, ok := names.([]interface{})
		if !ok {
			continue
		}
		for _, n := range nameList {
			name := fmt.Sprintf("%v", n)
			marker := "  "
			if dir == currentDir && name == currentName {
				marker = "* "
			}
			fmt.Printf("    %s%s\n", marker, name)
		}
	}
	fmt.Println()
	fmt.Printf("  (* = current)  Switch with: track <dir> <name>\n")
	fmt.Println()
}

func (c *consoleClient) cmdTrack(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: track <dir> <name>")
		fmt.Println("  Run 'tracks' to see available dirs and names.")
		return
	}
	trackDir := args[0]
	trackName := strings.Join(args[1:], " ")

	// Fetch current session to preserve gamemode and maxplayers
	gameMode := 1
	maxPlayers := 200
	if data, err := c.get("/status"); err == nil {
		if sess, ok := data["session"].(string); ok {
			var s map[string]interface{}
			if json.Unmarshal([]byte(sess), &s) == nil {
				if gm, ok := s["GameMode"].(float64); ok {
					gameMode = int(gm)
				}
				if mp, ok := s["MaxPlayers"].(float64); ok {
					maxPlayers = int(mp)
				}
			}
		}
	}

	status, _, err := c.post("/session/set", map[string]interface{}{
		"gamemode":   gameMode,
		"trackDir":   trackDir,
		"track":      trackName,
		"maxPlayers": maxPlayers,
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if status == 204 {
		fmt.Printf("Switching to %s / %s\n", trackDir, trackName)
	} else {
		fmt.Printf("Track switch failed (status %d) — check dir/name with 'tracks'.\n", status)
	}
}

func (c *consoleClient) cmdSession(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: session <end | start | set <gamemode> <trackDir> <trackName> <maxPlayers>>")
		return
	}
	switch strings.ToLower(args[0]) {

	case "end":
		status, _, err := c.post("/session/end", nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		switch status {
		case 204:
			fmt.Println("Session ended.")
		case 400:
			fmt.Println("Session is already ended.")
		default:
			fmt.Printf("Unexpected status: %d\n", status)
		}

	case "start":
		status, _, err := c.post("/session/start", nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		switch status {
		case 204:
			fmt.Println("Session started.")
		case 400:
			fmt.Println("Session is already running.")
		default:
			fmt.Printf("Unexpected status: %d\n", status)
		}

	case "set":
		// session set <gamemode> <trackDir> <trackName> <maxPlayers>
		if len(args) < 5 {
			fmt.Println("Usage: session set <gamemode> <trackDir> <trackName> <maxPlayers>")
			fmt.Println("  Gamemodes: 0=Casual  1=Competitive")
			return
		}
		gm, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println("Invalid gamemode:", args[1])
			return
		}
		trackDir := args[2]
		trackName := args[3]
		mp, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println("Invalid maxPlayers:", args[4])
			return
		}
		status, _, err := c.post("/session/set", map[string]interface{}{
			"gamemode":   gm,
			"trackDir":   trackDir,
			"track":      trackName,
			"maxPlayers": mp,
		})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if status == 204 {
			fmt.Printf("Session set: gamemode=%d, track=%s/%s, maxPlayers=%d\n", gm, trackDir, trackName, mp)
		} else {
			fmt.Printf("Failed (status %d) — track not found?\n", status)
		}

	default:
		fmt.Println("Unknown session subcommand:", args[0])
		fmt.Println("Usage: session <end | start | set>")
	}
}

func (c *consoleClient) cmdInvite(args []string) {
	if len(args) == 0 {
		data, err := c.get("/status")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("Invite: %v   (expires in %v)\n", data["invite"], data["timeoutIn"])
		return
	}
	switch strings.ToLower(args[0]) {

	case "new":
		status, data, err := c.post("/invite", map[string]interface{}{"regenerate": true})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if status == 200 && data != nil {
			fmt.Printf("New invite: %v   (key: %v, expires in %v)\n", data["invite"], data["key"], data["timeoutIn"])
		} else {
			fmt.Printf("Failed (status %d).\n", status)
		}

	case "key":
		if len(args) < 2 {
			fmt.Println("Usage: invite key <key>")
			return
		}
		key := args[1]
		status, data, err := c.post("/invite", map[string]interface{}{"key": key})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if status == 200 && data != nil {
			fmt.Printf("Invite with key %q: %v   (expires in %v)\n", key, data["invite"], data["timeoutIn"])
		} else {
			fmt.Printf("Failed (status %d).\n", status)
		}

	default:
		fmt.Println("Usage: invite [new | key <key>]")
	}
}

func (c *consoleClient) cmdAutoRotate(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: autorotate <start <folder> <secs> | stop | skip>")
		return
	}
	switch strings.ToLower(args[0]) {

	case "start":
		if len(args) < 3 {
			fmt.Println("Usage: autorotate start <folder> <interval_seconds>")
			return
		}
		folder := args[1]
		interval, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("Invalid interval:", args[2])
			return
		}
		status, _, err := c.post("/autorotate/start", map[string]interface{}{
			"folder":   folder,
			"interval": interval,
		})
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if status == 204 {
			fmt.Printf("Auto-rotate started: folder=%s, interval=%ds\n", folder, interval)
		} else {
			fmt.Printf("Failed (status %d).\n", status)
		}

	case "stop":
		status, _, err := c.post("/autorotate/stop", nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if status == 204 {
			fmt.Println("Auto-rotate stopped.")
		} else {
			fmt.Printf("Failed (status %d).\n", status)
		}

	case "skip":
		status, _, err := c.post("/autorotate/skip", nil)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if status == 204 {
			fmt.Println("Skipped to next track.")
		} else {
			fmt.Printf("Failed (status %d).\n", status)
		}

	default:
		fmt.Println("Unknown autorotate subcommand:", args[0])
		fmt.Println("Usage: autorotate <start | stop | skip>")
	}
}

func (c *consoleClient) cmdReload() {
	status, _, err := c.post("/reloadTracks", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if status == 204 {
		fmt.Println("Tracks reloaded from disk.")
	} else {
		fmt.Printf("Failed (status %d).\n", status)
	}
}

func (c *consoleClient) cmdStop() {
	if c.cmd.Process == nil {
		fmt.Println("Server process is not running.")
		return
	}
	if err := c.cmd.Process.Kill(); err != nil {
		fmt.Println("Error stopping server:", err)
		return
	}
	fmt.Println("Game server stopped.")
}

func (c *consoleClient) cmdExit() {
	fmt.Println("Shutting down...")
	if c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
	os.Exit(0)
}
