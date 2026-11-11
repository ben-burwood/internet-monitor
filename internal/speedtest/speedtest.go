package speedtest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
//
// onProgress, when non-nil, is called synchronously with a fresh snapshot for
// each progress line the CLI emits while the test runs.
func Run(ctx context.Context, serverID int, onProgress func(Progress)) (Result, error) {
	cmd := exec.CommandContext(ctx, binary, buildArgs(serverID)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{}, fmt.Errorf("run speedtest: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return Result{}, fmt.Errorf("run speedtest: %w", err)
	}

	o, parseErr := readStream(stdout, onProgress)
	// Drain anything left so the CLI never blocks on a full pipe before Wait.
	io.Copy(io.Discard, stdout)

	if err := cmd.Wait(); err != nil {
		// Output() would fill ExitError.Stderr for us; with a pipe we do it by hand.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitErr.Stderr = stderr.Bytes()
		}
		// With nothing on stderr and no timeout, the stdout error (e.g. a
		// logged CLI error) explains the failure better than an exit status.
		if stderr.Len() == 0 && ctx.Err() == nil && parseErr != nil {
			return Result{}, parseErr
		}
		return Result{}, wrapExecError(err)
	}
	if parseErr != nil {
		return Result{}, parseErr
	}

	res := o.toResult(uuid.NewString())
	if res.Timestamp.IsZero() {
		res.Timestamp = time.Now().UTC()
	}
	return res, nil
}

// readStream consumes the CLI's JSONL output, reporting progress lines to
// onProgress, and returns the final "result" line. If the output ends without
// one, the last error the CLI logged (if any) becomes the returned error.
func readStream(r io.Reader, onProgress func(Progress)) (ooklaResult, error) {
	var progress Progress
	var loggedErr string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var o ooklaResult
		if err := json.Unmarshal(line, &o); err != nil {
			return ooklaResult{}, fmt.Errorf("parse speedtest output: %w", err)
		}
		switch {
		case o.Type == "result":
			return o, nil
		case o.Type == "log":
			if o.Level == "error" && o.Message != "" {
				loggedErr = o.Message
			}
		case progress.apply(o) && onProgress != nil:
			onProgress(progress)
		}
	}
	if err := scanner.Err(); err != nil {
		return ooklaResult{}, fmt.Errorf("read speedtest output: %w", err)
	}
	if loggedErr != "" {
		return ooklaResult{}, fmt.Errorf("speedtest failed: %s", loggedErr)
	}
	return ooklaResult{}, errors.New("speedtest output ended without a result")
}

// baseArgs are the flags every Ookla invocation needs: the license/GDPR
// acceptance so the CLI never blocks on a prompt.
func baseArgs() []string {
	return []string{
		"--accept-license",
		"--accept-gdpr",
	}
}

// buildArgs assembles the Ookla CLI arguments for a test run: line-delimited
// JSON with progress updates, pinning a server only when serverID is set (> 0).
func buildArgs(serverID int) []string {
	args := append([]string{"--format=jsonl", "--progress=yes"}, baseArgs()...)
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
