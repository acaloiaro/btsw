package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

var (
	version = "dev"
	commit  = "none"
	date    = time.Now().Format(time.RFC3339)
)

const (
	bluezBusName       = "org.bluez"
	objectManagerIface = "org.freedesktop.DBus.ObjectManager"
	deviceIface        = "org.bluez.Device1"
	adapterIface       = "org.bluez.Adapter1"
)

type deviceInfo struct {
	path      dbus.ObjectPath
	addr      string
	name      string
	connected bool
	paired    bool
}

type adapterInfo struct {
	name string // e.g. "hci0"
	path dbus.ObjectPath
}

func main() {
	executableTable := path.Base(os.Args[0])
	usage := func() {

		fmt.Printf(`Usage:
  %s [--adapter=hciX] <command> [args]

Commands:
  adapters                List all available Bluetooth adapters.
  devices                 List paired devices on the selected adapter (default: hci0).
  connect <id>            Connect to a paired device by its numeric ID (from 'devices').
  disconnect <id>         Disconnect from a paired device by its numeric ID (from 'devices').

Options:
  --adapter=hciX          Specify which Bluetooth adapter to use (default: hci0).
  --version               Print the version. 

Examples:
  %s adapters               # List available bluetooth adapaters 
  %s devices                # List available devices on the chosen bluetooth adapter 
  %s --adapter=hci1 devices # List devics on hci1
  %s connect 0              # Connect to device 0 (from 'devices' list)
  %s disconnect 1           # Disconnect from device 1 (from 'devics' list)
`, executableTable, executableTable, executableTable, executableTable, executableTable, executableTable)
	}
	flags := flag.NewFlagSet("global", flag.ContinueOnError)

	flags.Usage = usage
	adapter := flags.String("adapter", "hci0", "Bluetooth adapter (e.g., hci0, hci1)")
	versionFlag := flags.Bool("version", false, "Print the version")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	if versionFlag != nil && *versionFlag {
		fmt.Fprintf(os.Stdout, "btsw version: %s built at: %s from commit: %s\n", version, date, commit)
		os.Exit(0)
	}

	args := flags.Args()
	if len(args) < 1 {
		usage()
		os.Exit(0)
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Failed to connect to system bus: %v", err)
	}

	cmd := os.Args[1]
	switch cmd {
	case "adapters":
		adapters, err := getAdapters(conn)
		if err != nil {
			log.Fatalf("Error fetching bluetooth adapters: %v", err)
		}
		if len(adapters) == 0 {
			fmt.Println("No bluetooth adapters found.")
			return
		}
		for i, a := range adapters {
			fmt.Printf("[%d] %s\n", i, a.name)
		}

	case "devices":
		paired, err := getPairedDevices(conn, *adapter)
		if err != nil {
			log.Fatalf("Error fetching paired devices: %v", err)
		}
		if len(paired) == 0 {
			fmt.Println("No paired devices found.")
			return
		}
		for i, d := range paired {
			fmt.Printf("[%d] %s (%s) Connected=%t\n", i, displayName(d), d.addr, d.connected)
		}

	case "connect", "disconnect":
		if len(args) < 2 {
			usage()
			os.Exit(1)
		}
		id, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("Invalid id: %v", err)
		}

		paired, err := getPairedDevices(conn, *adapter)
		if err != nil {
			log.Fatalf("Error fetching paired devices: %v", err)
		}
		if id < 0 || id >= len(paired) {
			log.Fatalf("Invalid device id %d", id)
		}
		d := paired[id]

		devObj := conn.Object(bluezBusName, d.path)
		if cmd == "connect" {
			fmt.Printf("Connecting to %s on %s...\n", displayName(d), *adapter)
			if call := devObj.Call(deviceIface+".Connect", 0); call.Err != nil {
				log.Fatalf("Failed to connect: %v", call.Err)
			}
			fmt.Println("Connected.")
		} else {
			fmt.Printf("Disconnecting from %s on %s...\n", displayName(d), *adapter)
			if call := devObj.Call(deviceIface+".Disconnect", 0); call.Err != nil {
				log.Fatalf("Failed to disconnect: %v", call.Err)
			}
			fmt.Println("Disconnected.")
		}

	default:
		usage()
		os.Exit(0)
	}
}

func getBool(props map[string]dbus.Variant, key string) bool {
	if v, ok := props[key]; ok {
		if b, ok := v.Value().(bool); ok {
			return b
		}
	}
	return false
}

func getString(props map[string]dbus.Variant, key string) string {
	if v, ok := props[key]; ok {
		if s, ok := v.Value().(string); ok {
			return s
		}
	}
	return ""
}

func displayName(d deviceInfo) string {
	if d.name != "" {
		return d.name
	}
	if d.addr != "" {
		return d.addr
	}
	return string(d.path)
}

// getPairedDevices queries BlueZ for paired devices on a specific adapter
func getPairedDevices(conn *dbus.Conn, adapter string) ([]deviceInfo, error) {
	obj := conn.Object(bluezBusName, dbus.ObjectPath("/"))

	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := obj.Call(objectManagerIface+".GetManagedObjects", 0).Store(&managedObjects); err != nil {
		return nil, fmt.Errorf("GetManagedObjects failed: %w", err)
	}

	adapterPath := "/org/bluez/" + adapter
	var paired []deviceInfo
	for objPath, ifaces := range managedObjects {
		if props, ok := ifaces[deviceIface]; ok {
			// Only include devices under the chosen adapter
			if !strings.HasPrefix(string(objPath), adapterPath) {
				continue
			}
			if getBool(props, "Paired") {
				paired = append(paired, deviceInfo{
					path:      objPath,
					addr:      getString(props, "Address"),
					name:      getString(props, "Name"),
					connected: getBool(props, "Connected"),
					paired:    true,
				})
			}
		}
	}

	// Ensure stable order by sorting on MAC address
	sort.Slice(paired, func(i, j int) bool {
		return paired[i].addr < paired[j].addr
	})

	return paired, nil
}

// getAdapters queries BlueZ for available adapters
func getAdapters(conn *dbus.Conn) ([]adapterInfo, error) {
	obj := conn.Object(bluezBusName, dbus.ObjectPath("/"))

	var managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := obj.Call(objectManagerIface+".GetManagedObjects", 0).Store(&managedObjects); err != nil {
		return nil, fmt.Errorf("GetManagedObjects failed: %w", err)
	}

	var adapters []adapterInfo
	for objPath, ifaces := range managedObjects {
		if _, ok := ifaces[adapterIface]; ok {
			parts := strings.Split(string(objPath), "/")
			if len(parts) > 0 {
				name := parts[len(parts)-1] // e.g. "hci0"
				adapters = append(adapters, adapterInfo{name: name, path: objPath})
			}
		}
	}

	sort.Slice(adapters, func(i, j int) bool {
		return adapters[i].name < adapters[j].name
	})

	return adapters, nil
}
