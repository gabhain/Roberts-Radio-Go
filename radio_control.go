// Package main provides a command-line interface for controlling Roberts radios
// via the UNDOK FSAPI protocol.
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	radioPin = "1234"
	radioIP  string
)

// FSAPIResponse represents the standard XML response structure from the radio's FSAPI.
type FSAPIResponse struct {
	XMLName xml.Name `xml:"fsapiResponse"`
	Status  string   `xml:"status"`
	Value   struct {
		InnerXML string `xml:",innerxml"`
	} `xml:"value"`
}

// GetValue extracts the inner text value from the FSAPIResponse's value tag, 
// stripping any nested XML tags.
func (f *FSAPIResponse) GetValue() string {
	raw := f.Value.InnerXML
	// Find the value between the first '>' and the following '<'
	// This handles <u8>1</u8>, <s32>123</s32>, <enum>0</enum>, etc.
	start := 0
	for i, c := range raw {
		if c == '>' {
			start = i + 1
			break
		}
	}
	end := len(raw)
	for i := start; i < len(raw); i++ {
		if raw[i] == '<' {
			end = i
			break
		}
	}
	if start > end {
		return ""
	}
	return raw[start:end]
}

func usage() {
	exe := filepath.Base(os.Args[0])
	fmt.Printf(`Roberts Radio Control (Go Version)

Usage: %s [flags] [command] [value]

Flags:
  -i, --ip <addr>   IP address of the radio (default: XX.XX.XX.XX)
  -h, --help        Show this help message

Commands:
  on                Turn the radio ON
  off               Turn the radio OFF (Standby)
  status            Check power status
  vol [0-32]        Set volume or get current volume
  volup             Increase volume by 1
  voldown           Decrease volume by 1
  mute              Mute the radio
  unmute            Unmute the radio
  togglemute        Toggle mute state
  mode [id]         Set source mode or get current mode
  next              Next track or station
  prev              Previous track or station
  play              Start playback
  pause             Pause playback
  info              Show "Now Playing" information
  pair              Initiate Bluetooth pairing
  device            Show device information (Model, Version, IP)

Common Mode IDs:
  0: Internet Radio   1: Tidal           2: Deezer
  3: Amazon Music     4: Spotify         5: Local Music
  6: Music Player     7: DAB             8: FM Radio
  9: Bluetooth        10: AUX

Note: Run "mode" without an ID to see the current source's ID.

`, exe)
	os.Exit(1)
}

// fsapiCall performs an HTTP GET request to the radio's FSAPI and returns the parsed response.
func fsapiCall(method, path, value string) (*FSAPIResponse, error) {
	url := fmt.Sprintf("http://%s/fsapi/%s/%s?pin=%s", radioIP, method, path, radioPin)
	if value != "" {
		url += "&value=" + value
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to reach radio at %s: %w", radioIP, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("radio returned unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var fsResp FSAPIResponse
	if err := xml.Unmarshal(body, &fsResp); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	if fsResp.Status != "FS_OK" {
		return nil, fmt.Errorf("FSAPI returned error status: %s", fsResp.Status)
	}

	return &fsResp, nil
}

func main() {
	// Setup and parse command-line flags.
	defaultIP := os.Getenv("RADIO_IP")
	if defaultIP == "" {
		defaultIP = "XX.XX.XX.XX"
	}

	flag.StringVar(&radioIP, "ip", defaultIP, "IP address of the radio")
	flag.StringVar(&radioIP, "i", defaultIP, "IP address of the radio (shorthand)")
	help := flag.Bool("help", false, "Show help")
	h := flag.Bool("h", false, "Show help (shorthand)")

	flag.Parse()

	if *help || *h {
		usage()
	}

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	cmd := args[0]

	if err := runCommand(cmd, args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runCommand dispatches the requested command to the appropriate logic.
func runCommand(cmd string, args []string) error {
	switch cmd {
	case "on":
		fmt.Println("Turning Radio ON...")
		_, err := fsapiCall("SET", "netRemote.sys.power", "1")
		return err

	case "off":
		fmt.Println("Turning Radio OFF (Standby)...")
		_, err := fsapiCall("SET", "netRemote.sys.power", "0")
		return err

	case "status":
		resp, err := fsapiCall("GET", "netRemote.sys.power", "")
		if err != nil {
			return err
		}
		if resp.GetValue() == "1" {
			fmt.Println("Radio is currently: ON")
		} else {
			fmt.Println("Radio is currently: OFF")
		}
		return nil

	case "vol":
		if len(args) > 0 {
			fmt.Printf("Setting Volume to %s...\n", args[0])
			_, err := fsapiCall("SET", "netRemote.sys.audio.volume", args[0])
			return err
		}
		resp, err := fsapiCall("GET", "netRemote.sys.audio.volume", "")
		if err != nil {
			return err
		}
		fmt.Printf("Current Volume: %s\n", resp.GetValue())
		return nil

	case "volup":
		resp, err := fsapiCall("GET", "netRemote.sys.audio.volume", "")
		if err != nil {
			return err
		}
		vol, err := strconv.Atoi(resp.GetValue())
		if err != nil {
			return fmt.Errorf("invalid volume value from radio: %w", err)
		}
		newVol := vol + 1
		if newVol > 32 {
			newVol = 32
		}
		fmt.Printf("Increasing volume to %d...\n", newVol)
		_, err = fsapiCall("SET", "netRemote.sys.audio.volume", strconv.Itoa(newVol))
		return err

	case "voldown":
		resp, err := fsapiCall("GET", "netRemote.sys.audio.volume", "")
		if err != nil {
			return err
		}
		vol, err := strconv.Atoi(resp.GetValue())
		if err != nil {
			return fmt.Errorf("invalid volume value from radio: %w", err)
		}
		newVol := vol - 1
		if newVol < 0 {
			newVol = 0
		}
		fmt.Printf("Decreasing volume to %d...\n", newVol)
		_, err = fsapiCall("SET", "netRemote.sys.audio.volume", strconv.Itoa(newVol))
		return err

	case "mute":
		fmt.Println("Muting...")
		_, err := fsapiCall("SET", "netRemote.sys.audio.mute", "1")
		return err

	case "unmute":
		fmt.Println("Unmuting...")
		_, err := fsapiCall("SET", "netRemote.sys.audio.mute", "0")
		return err

	case "togglemute":
		resp, err := fsapiCall("GET", "netRemote.sys.audio.mute", "")
		if err != nil {
			return err
		}
		if resp.GetValue() == "1" {
			fmt.Println("Unmuting...")
			_, err = fsapiCall("SET", "netRemote.sys.audio.mute", "0")
		} else {
			fmt.Println("Muting...")
			_, err = fsapiCall("SET", "netRemote.sys.audio.mute", "1")
		}
		return err

	case "mode":
		if len(args) > 0 {
			fmt.Printf("Changing Mode to %s...\n", args[0])
			_, err := fsapiCall("SET", "netRemote.sys.mode", args[0])
			return err
		}
		resp, err := fsapiCall("GET", "netRemote.sys.mode", "")
		if err != nil {
			return err
		}
		fmt.Printf("Current Mode: %s\n", resp.GetValue())
		return nil

	case "next":
		fmt.Println("Next track/station...")
		_, err := fsapiCall("SET", "netRemote.nav.action.navigate", "1")
		return err

	case "prev":
		fmt.Println("Previous track/station...")
		_, err := fsapiCall("SET", "netRemote.nav.action.navigate", "-1")
		return err

	case "play":
		fmt.Println("Playing...")
		_, err := fsapiCall("SET", "netRemote.nav.state", "1")
		return err

	case "pause":
		fmt.Println("Pausing...")
		_, err := fsapiCall("SET", "netRemote.nav.state", "2")
		return err

	case "info":
		name, err := fsapiCall("GET", "netRemote.play.info.name", "")
		if err != nil {
			return err
		}
		text, err := fsapiCall("GET", "netRemote.play.info.text", "")
		if err != nil {
			return err
		}
		fmt.Printf("Now Playing: %s\n", name.GetValue())
		fmt.Printf("Info: %s\n", text.GetValue())
		return nil

	case "pair":
		fmt.Println("Switching to Bluetooth mode (ID 9)...")
		if _, err := fsapiCall("SET", "netRemote.sys.mode", "9"); err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
		fmt.Println("Initiating Bluetooth pairing...")
		_, err := fsapiCall("SET", "netRemote.bluetooth.pairing", "1")
		return err

	case "device":
		friendly, _ := fsapiCall("GET", "netRemote.sys.info.friendlyName", "")
		model, _ := fsapiCall("GET", "netRemote.sys.info.modelName", "")
		version, _ := fsapiCall("GET", "netRemote.sys.info.version", "")
		ipRaw, _ := fsapiCall("GET", "netRemote.sys.net.ipConfig.address", "")

		ip := ipRaw.GetValue()
		if i, err := strconv.ParseUint(ip, 10, 32); err == nil {
			ip = fmt.Sprintf("%d.%d.%d.%d", byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
		}

		fmt.Printf("Device Name: %s\n", friendly.GetValue())
		fmt.Printf("Model:       %s\n", model.GetValue())
		fmt.Printf("Version:     %s\n", version.GetValue())
		fmt.Printf("Radio IP:    %s\n", ip)
		return nil

	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
