package speedtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// binary is the Ookla CLI executable name, expected on PATH.
const binary = "speedtest"

// Run executes a single Ookla speedtest and returns the parsed result.
//
// serverID pins a specific Ookla server; when it is <= 0 the CLI selects the
// closest server automatically. The license/GDPR acceptance flags are always
// passed so the CLI never blocks on a prompt.
func Run(ctx context.Context, serverID int) (Result, error) {
	cmd := exec.CommandContext(ctx, binary, buildArgs(serverID)...)

	out, err := cmd.Output()
	if err != nil {
		return Result{}, wrapExecError(err)
	}

	var o ooklaResult
	if err := json.Unmarshal(out, &o); err != nil {
		return Result{}, fmt.Errorf("parse speedtest output: %w", err)
	}
	if o.Type != "result" {
		return Result{}, fmt.Errorf("unexpected speedtest output type %q", o.Type)
	}

	res := o.toResult(uuid.NewString())
	if res.Timestamp.IsZero() {
		res.Timestamp = time.Now().UTC()
	}
	return res, nil
}

// baseArgs are the flags every Ookla invocation needs: JSON output plus the
// license/GDPR acceptance so the CLI never blocks on a prompt.
func baseArgs() []string {
	return []string{
		"--format=json",
		"--accept-license",
		"--accept-gdpr",
	}
}

// buildArgs assembles the Ookla CLI arguments, pinning a server only when
// serverID is set (> 0).
func buildArgs(serverID int) []string {
	args := baseArgs()
	if serverID > 0 {
		args = append(args, "--server-id="+strconv.Itoa(serverID))
	}
	return args
}

// wrapExecError turns an *exec.ExitError into a message that includes the
// CLI's stderr, which is where Ookla reports failures (no servers, etc.).
func wrapExecError(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		stderr := strings.TrimSpace(string(exitErr.Stderr))
		if stderr != "" {
			return fmt.Errorf("speedtest failed: %s", stderr)
		}
	}
	return fmt.Errorf("run speedtest: %w", err)
}
