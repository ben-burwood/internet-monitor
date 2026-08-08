package speedtest

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Server is a nearby Ookla speedtest server the user can pin.
type Server struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Country  string `json:"country"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// ooklaServerList mirrors `speedtest --servers -f json`:
// {"type":"serverList","timestamp":"…","servers":[{id,host,port,name,location,country}]}
type ooklaServerList struct {
	Type    string   `json:"type"`
	Servers []Server `json:"servers"`
}

// ListServers returns the nearest Ookla servers, as offered by the CLI.
func ListServers(ctx context.Context) ([]Server, error) {
	cmd := exec.CommandContext(ctx, binary, append([]string{"--servers"}, baseArgs()...)...)

	out, err := cmd.Output()
	if err != nil {
		return nil, wrapExecError(err)
	}
	return parseServers(out)
}

func parseServers(out []byte) ([]Server, error) {
	var list ooklaServerList
	if err := json.Unmarshal(out, &list); err != nil {
		return nil, fmt.Errorf("parse server list: %w", err)
	}
	return list.Servers, nil
}
