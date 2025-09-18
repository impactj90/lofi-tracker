package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

const detachedPrefix = "detached-"

func GetCurrentBranchName() (string, error) {
	return resolveBranchLabel(context.Background())
}

func resolveBranchLabel(ctx context.Context) (string, error) {
	name, err := runGitCommand(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}

	if name == "HEAD" || name == "" {
		shortSHA, shaErr := runGitCommand(ctx, "rev-parse", "--short", "HEAD")
		if shaErr != nil {
			return "", shaErr
		}

		return fmt.Sprintf("%s%s", detachedPrefix, shortSHA), nil
	}

	return name, nil
}

func runGitCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
