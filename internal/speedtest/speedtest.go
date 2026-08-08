package speedtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

// binary is the Ookla CLI executable name, expected on PATH.
const binary = "speedtest"

// Run executes a single Ookla speedtest and returns the parsed result.
//
// The Ookla CLI selects the closest server automatically. The license/GDPR
// acceptance flags are always passed so the CLI never blocks on a prompt.
func Run(ctx context.Context) (Result, error) {
	cmd := exec.CommandContext(ctx, binary,
		"--format=json",
		"--accept-license",
		"--accept-gdpr",
	)

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
