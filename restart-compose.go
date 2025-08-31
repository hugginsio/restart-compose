// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
	"github.com/hugginsio/restart-compose/internal/compose"
	"github.com/hugginsio/restart-compose/internal/config"
	"github.com/hugginsio/restart-compose/internal/handler"
	"github.com/hugginsio/restart-compose/internal/tailnet"
)

// validateStackExists checks if the compose file exists relative to the config directory
func validateStackExists(configDir, stackPath string) (string, bool) {
	fullPath := filepath.Join(configDir, stackPath)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", false
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return absPath, false
	}

	return absPath, true
}

func main() {
	var baseDir string
	flag.StringVar(&baseDir, "d", "", "Directory to scan for .restart-compose.yaml config file (defaults to current directory)")
	flag.Parse()

	log.Println("restart-compose is starting")

	configPath, err := config.Find(baseDir)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	config, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if len(config.Stacks) == 0 {
		log.Fatalln("No stacks configured in .restart-compose.yaml")
	}

	// Get the directory containing the config file
	baseDir = filepath.Dir(configPath)

	// Initialize Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	defer func() {
		if closeErr := dockerClient.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				log.Fatalf("Error closing Docker client: %v", closeErr)
			}
		}
	}()

	// Validate and gather information about each stack
	var stacks []compose.StackInfo
	ctx := context.Background()

	for _, stackPath := range config.Stacks {
		stackName := compose.GetStackName(stackPath)
		fullPath, exists := validateStackExists(baseDir, stackPath)

		stackInfo := compose.StackInfo{
			Path:   fullPath,
			Name:   stackName,
			Exists: exists,
		}

		if exists {
			// Get running services for this stack
			services, err := compose.GetStackServices(ctx, dockerClient, stackName)
			if err != nil {
				log.Printf("Warning: failed to get services for stack %s: %v", stackName, err)
			} else {
				stackInfo.Services = services
			}
		}

		stacks = append(stacks, stackInfo)
	}

	for _, stack := range stacks {
		compose.PrintStackInfo(stack)
	}

	hostname := "restart-compose"
	if os.Getenv("DEBUG") == "true" {
		hostname += "-debug"
	}

	// TODO: health endpoint to monitor tailnet connection
	funneln, baseUrl := tailnet.Funnel(hostname)

	http.Handle("/github", handler.NewGitHub(&handler.GitHubConfig{Secret: os.Getenv("GH_SECRET"), Stacks: stacks, Path: baseDir}))
	http.Handle("/ping", http.HandlerFunc(handler.Ping))

	log.Printf("Funnel available: %s\n", baseUrl+"/github")
	if err := http.Serve(funneln, nil); err != nil {
		log.Fatalf("Error serving HTTP server: %v", err)
	}
}
