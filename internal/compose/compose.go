// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package compose

import (
	"github.com/docker/docker/api/types/container"
)

// StackInfo holds information about a Docker Compose stack
type StackInfo struct {
	Path     string
	Name     string
	Services []container.Summary
	Exists   bool
}
