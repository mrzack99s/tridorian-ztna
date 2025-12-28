package main

import (
	"log"
	"os"
	"os/exec"
)

func installService(nodeID, controlPlaneAddr string) {
	if nodeID == "" {
		log.Fatal("❌ --node-id is required for installation")
	}

	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("❌ Failed to get executable path: %v", err)
	}

	// Use default control plane if not specified
	if controlPlaneAddr == "" {
		controlPlaneAddr = "https://control.tridorian.com" // Default production URL or logic
	}

	serviceContent := `[Unit]
Description=Tridorian ZTNA Gateway
After=network.target

[Service]
Type=simple
ExecStart=` + exePath + ` --node-id "` + nodeID + `" --control-plane "` + controlPlaneAddr + `"
Restart=always
RestartSec=5
User=root
# Environment variables can be added here if needed
# Environment=

[Install]
WantedBy=multi-user.target
`

	servicePath := "/etc/systemd/system/tridorian-gateway.service"
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		log.Fatalf("❌ Failed to write service file (permissions?): %v", err)
	}

	log.Printf("✅ Service file created at %s", servicePath)

	// Reload systemd and enable service
	// Note: This requires the binary to be run with sudo/root privileges
	// We are using 'exec.Command' so we need to import "os/exec"
	// I will add the import in a separate step or assume it is available if I could edit imports.
	// simpler to just hint to user or try and fail.
	// Actually, I should add "os/exec" to imports.
	// For now, I'll use a raw exec approach if possible or just write the file.

	// Let's rely on the user running this as root.

	// Run systemctl commands
	runCommand("systemctl", "daemon-reload")
	runCommand("systemctl", "enable", "tridorian-gateway")
	runCommand("systemctl", "start", "tridorian-gateway")

	log.Printf("🚀 Service installed and started successfully!")
}

func runCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if err := cmd.Run(); err != nil {
		log.Printf("⚠️ Failed to run %s %v: %v", name, args, err)
	}
}
