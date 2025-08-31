// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/hugginsio/restart-compose/internal/compose"
	"github.com/hugginsio/restart-compose/internal/git"
)

type GitHubConfig struct {
	Secret string
	Stacks []compose.StackInfo
	Path   string
}

type GitHubEvent struct {
	Ref        string `json:"ref"`
	Repository struct {
		DefaultBranch string `json:"default_branch"`
	} `json:"repository"`
	Commits []struct {
		ID       string   `json:"id"`
		Added    []any    `json:"added"`
		Removed  []any    `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"commits"`
}

func NewGitHub(config *GitHubConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		eventType := r.Header.Get("X-GitHub-Event")
		deliveryId := r.Header.Get("X-GitHub-Delivery")
		signature := r.Header.Get("X-Hub-Signature-256")

		if deliveryId == "" {
			http.Error(w, "Missing delivery ID", http.StatusBadRequest)
			return
		}

		if signature == "" {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		if !verifySignature(config.Secret, body, signature) {
			http.Error(w, "Invalid signature", http.StatusForbidden)
			return
		}

		if eventType != "push" && eventType != "ping" {
			http.Error(w, "Invalid event. This server only processes 'push' and 'ping' events.", http.StatusBadRequest)
			return
		}

		log.Printf("GitHub %s: received %s", deliveryId, eventType)

		if eventType == "ping" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var event GitHubEvent
		if err := json.Unmarshal(body, &event); err != nil {
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// Only process pushes to the default branch
		if !strings.HasSuffix(event.Ref, event.Repository.DefaultBranch) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Only process pushes that contain commits
		if len(event.Commits) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		pathsToRestart := []string{}
		for _, v := range event.Commits {
			pathsToRestart = append(pathsToRestart, v.Modified...)
		}

		if len(pathsToRestart) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		stackNamesToRestart := []string{}
		stacksToRestart := []compose.StackInfo{}
		for _, path := range pathsToRestart {
			if path == "" {
				continue
			}

			for _, stack := range config.Stacks {
				if strings.HasSuffix(stack.Path, path) {
					stackNamesToRestart = append(stackNamesToRestart, stack.Name)
					stacksToRestart = append(stacksToRestart, stack)
					break
				}
			}
		}

		log.Printf("GitHub %s: restarting stacks %v", deliveryId, stackNamesToRestart)

		for _, stack := range stacksToRestart {
			if err := compose.StopStack(ctx, stack); err != nil {
				log.Printf("GitHub %s: failed to stop stack %s: %v", deliveryId, stack.Name, err)
			} else {
				log.Printf("GitHub %s: successfully stopped stack %s", deliveryId, stack.Name)
			}
		}

		log.Printf("GitHub %s: updating git repository", deliveryId)
		if err := git.Update(ctx, config.Path); err != nil {
			log.Printf("GitHub %s: %v", deliveryId, err)
		}

		for _, stack := range stacksToRestart {
			if err := compose.StartStack(ctx, stack); err != nil {
				log.Printf("GitHub %s: failed to start stack %s: %v", deliveryId, stack.Name, err)
			} else {
				log.Printf("GitHub %s: successfully restarted stack %s", deliveryId, stack.Name)
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

// verifySignature verifies the HMAC-SHA256 signature of the request body
func verifySignature(secret string, body []byte, signature string) bool {
	signature = strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)

	// Convert to hex string
	expectedSignature := hex.EncodeToString(expectedMAC)

	// Compare signatures using constant time comparison
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
