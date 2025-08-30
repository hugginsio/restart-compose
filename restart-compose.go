// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure for restart-compose
type Config struct {
	Stacks []string `yaml:"stacks"`
}

// StackInfo holds information about a Docker Compose stack
type StackInfo struct {
	Path      string
	Name      string
	Services  []types.Container
	Exists    bool
	Directory string
}

// loadConfig reads and parses the .restart-compose.yaml configuration file
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// findConfigFile searches for .restart-compose.yaml in the specified directory
func findConfigFile(dir string) (string, error) {
	var searchDir string
	var err error

	if dir == "" {
		// If no directory specified, use current working directory
		searchDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
	} else {
		// Use the specified directory
		searchDir, err = filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve directory path: %w", err)
		}

		// Check if directory exists
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return "", fmt.Errorf("specified directory does not exist: %s", dir)
		}
	}

	configPath := filepath.Join(searchDir, ".restart-compose.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("configuration file .restart-compose.yaml not found in directory: %s", searchDir)
	}

	return configPath, nil
}

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

// getStackName extracts the stack name from the compose file path
func getStackName(stackPath string) string {
	dir := filepath.Dir(stackPath)
	return filepath.Base(dir)
}

// getStackServices retrieves running containers for a specific stack
func getStackServices(ctx context.Context, dockerClient *client.Client, stackName string) ([]types.Container, error) {
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var stackContainers []types.Container
	for _, container := range containers {
		// Check if container belongs to the stack by looking at labels
		if project, exists := container.Labels["com.docker.compose.project"]; exists && project == stackName {
			stackContainers = append(stackContainers, container)
		}
	}

	return stackContainers, nil
}

// printStackInfo displays information about a stack
func printStackInfo(stack StackInfo) {
	fmt.Printf("\n=== Stack: %s ===\n", stack.Name)
	fmt.Printf("Path: %s\n", stack.Path)
	fmt.Printf("Directory: %s\n", stack.Directory)
	fmt.Printf("Exists: %t\n", stack.Exists)

	if len(stack.Services) > 0 {
		fmt.Printf("Services (%d):\n", len(stack.Services))
		for _, service := range stack.Services {
			status := "Unknown"
			if len(service.Status) > 0 {
				status = service.Status
			}

			serviceName := "Unknown"
			if name, exists := service.Labels["com.docker.compose.service"]; exists {
				serviceName = name
			}

			fmt.Printf("  - %s: %s (%s)\n", serviceName, status, service.State)
		}
	} else {
		fmt.Println("Services: None running")
	}
}

func main() {
	// Parse command line flags
	var configDir string
	flag.StringVar(&configDir, "d", "", "Directory to scan for .restart-compose.yaml config file (defaults to current directory)")
	flag.Parse()

	fmt.Println("Starting restart-compose webhook listener...")

	// Find configuration file
	configPath, err := findConfigFile(configDir)
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	fmt.Printf("Found configuration file: %s\n", configPath)

	// Load configuration
	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if len(config.Stacks) == 0 {
		log.Fatal("No stacks configured in .restart-compose.yaml")
	}

	fmt.Printf("Loaded %d stack(s) from configuration\n", len(config.Stacks))

	// Get the directory containing the config file
	configDir = filepath.Dir(configPath)

	// Initialize Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	fmt.Println("Connected to Docker daemon")

	// Validate and gather information about each stack
	var stacks []StackInfo
	ctx := context.Background()

	for _, stackPath := range config.Stacks {
		stackName := getStackName(stackPath)
		fullPath, exists := validateStackExists(configDir, stackPath)

		stackInfo := StackInfo{
			Path:      fullPath,
			Name:      stackName,
			Exists:    exists,
			Directory: filepath.Dir(fullPath),
		}

		if exists {
			// Get running services for this stack
			services, err := getStackServices(ctx, dockerClient, stackName)
			if err != nil {
				log.Printf("Warning: Failed to get services for stack %s: %v", stackName, err)
			} else {
				stackInfo.Services = services
			}
		} else {
			log.Printf("Warning: Stack compose file not found: %s", fullPath)
		}

		stacks = append(stacks, stackInfo)
	}

	// Print stack information
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("DOCKER COMPOSE STACK INFORMATION")
	fmt.Println(strings.Repeat("=", 50))

	for _, stack := range stacks {
		printStackInfo(stack)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("Summary: %d stack(s) configured, %d exist on filesystem\n",
		len(stacks),
		countExistingStacks(stacks))

	// TODO: Add webhook server implementation
	fmt.Println("\nWebhook server functionality will be implemented next...")
}

// countExistingStacks counts how many stacks actually exist on the filesystem
func countExistingStacks(stacks []StackInfo) int {
	count := 0
	for _, stack := range stacks {
		if stack.Exists {
			count++
		}
	}
	return count
}
