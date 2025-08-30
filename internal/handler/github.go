// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"log"
	"net/http"
)

// GitHub processes incoming push webhooks from GitHub.
func GitHub(w http.ResponseWriter, r *http.Request) {
	log.Printf("marco")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	deliveryId := r.Header.Get("X-GitHub-Delivery")

	if event == "ping" {
		log.Printf("Received ping event from GitHub with delivery ID %s", deliveryId)
		w.WriteHeader(http.StatusOK)
		return
	}

	if event != "push" {
		http.Error(w, "Invalid event. This server only processes 'push' events.", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
