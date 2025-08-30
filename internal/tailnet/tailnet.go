// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package tailnet

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"tailscale.com/tsnet"
)

// Funnel creates a Tailscale Funnel with the provided hostname.
func Funnel(hostname string) (net.Listener, string) {
	server := &tsnet.Server{
		Dir:      hostname,
		Hostname: hostname,
		UserLogf: log.Printf,
	}

	srv, err := server.Up(context.Background())
	if err != nil {
		log.Fatalf("Tailscale failed to start: %v", err)
	}

	ln, err := server.ListenFunnel("tcp", ":443", tsnet.FunnelOnly())

	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", ln.Addr().String(), err)
	}

	return ln, fmt.Sprintf("https://%s", strings.TrimSuffix(srv.Self.DNSName, "."))
}
