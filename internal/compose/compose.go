// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package compose

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// StackInfo holds information about a Docker Compose stack
type StackInfo struct {
	Path     string
	Name     string
	Services []container.Summary
	Exists   bool
}

// GetStackName extracts the stack name from the compose file path
func GetStackName(stackPath string) string {
	dir := filepath.Dir(stackPath)
	return filepath.Base(dir)
}

// GetStackServices retrieves running containers for a specific stack
func GetStackServices(ctx context.Context, dockerClient *client.Client, stackName string) ([]container.Summary, error) {
	containers, err := dockerClient.ContainerList(ctx, container.ListOptions{
		All: true,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var stackContainers []container.Summary
	for _, container := range containers {
		// Check if container belongs to the stack by looking at labels
		if project, exists := container.Labels["com.docker.compose.project"]; exists && project == stackName {
			stackContainers = append(stackContainers, container)
		}
	}

	return stackContainers, nil
}

// PrintStackInfo displays information about a stack
func PrintStackInfo(stack StackInfo) {
	if stack.Exists {
		log.Printf("Stack %s found with %d services running", stack.Name, len(stack.Services))
	} else {
		log.Printf("WARN stack %s not found in filesystem", stack.Name)
	}
}

func StopStack(ctx context.Context, stack StackInfo) error {
	dir := filepath.Dir(stack.Path)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", stack.Path, "down")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop compose stack: %w", err)
	}

	return nil
}

func StartStack(ctx context.Context, stack StackInfo) error {
	dir := filepath.Dir(stack.Path)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", stack.Path, "up", "-d")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start compose stack: %w", err)
	}

	return nil
}
