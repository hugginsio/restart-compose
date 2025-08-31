// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package tailnet

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"tailscale.com/tsnet"
)

// Funnel creates a Tailscale Funnel with the provided hostname.
func Funnel(hostname string) (net.Listener, string) {
	server := &tsnet.Server{
		Dir:      filepath.Join(os.TempDir(), hostname),
		Hostname: hostname,
		UserLogf: log.Printf,
	}

	if _, err := os.Stat(server.Dir); os.IsNotExist(err) {
		log.Println("WARN new DNS records may take 10min to propogate to the public internet")
	}

	srv, err := server.Up(context.Background())
	if err != nil {
		log.Fatalf("Tailscale failed to start: %v", err)
	}

	if len(srv.CertDomains) == 0 {
		log.Fatalln("No certificate domains found. Do you have HTTPS enabled?")
	}

	ln, err := server.ListenFunnel("tcp", ":443")

	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", ln.Addr().String(), err)
	}

	return ln, fmt.Sprintf("https://%s", srv.CertDomains[0])
}
