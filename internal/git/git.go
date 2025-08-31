// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package git

import (
	"context"
	"fmt"
	"os/exec"
)

func Update(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "pull", "--no-edit", "--ff-only")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		out, _ := cmd.CombinedOutput()
		return fmt.Errorf("failed to update git repository: %w:%s", err, out)
	}

	return nil
}
