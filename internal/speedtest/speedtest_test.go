package speedtest

import (
	"strings"
	"testing"
)

func TestBuildArgs(t *testing.T) {
	base := []string{"--format=jsonl", "--progress=yes", "--accept-license", "--accept-gdpr"}

	t.Run("automatic when id <= 0", func(t *testing.T) {
		for _, id := range []int{0, -1} {
			got := buildArgs(id)
			if len(got) != len(base) {
				t.Fatalf("id=%d: got %v, want %v", id, got, base)
			}
			for _, a := range got {
				if a == "--server-id=0" || a == "--server-id=-1" {
					t.Errorf("id=%d: unexpected server flag %q", id, a)
				}
			}
		}
	})

	t.Run("pins server when id > 0", func(t *testing.T) {
		got := buildArgs(1234)
		last := got[len(got)-1]
		if last != "--server-id=1234" {
			t.Errorf("last arg = %q, want --server-id=1234", last)
		}
	})
}

func TestReadStream(t *testing.T) {
	// Trimmed real `speedtest -f jsonl -p yes` output.
	sample := `{"type":"testStart","timestamp":"2026-11-11T22:19:17Z","isp":"Virgin Media","interface":{"externalIp":"80.6.133.250"},"server":{"id":76429,"name":"Virgin Media O2","location":"Yate","country":"United Kingdom"}}
{"type":"ping","timestamp":"2026-11-11T22:19:17Z","ping":{"jitter":1.707,"latency":21.077,"progress":0.400}}
{"type":"download","timestamp":"2026-11-11T22:19:17Z","download":{"bandwidth":12500000,"bytes":12288,"elapsed":15,"progress":0.5}}
{"type":"upload","timestamp":"2026-11-11T22:19:30Z","upload":{"bandwidth":2500000,"bytes":201388,"elapsed":64,"progress":0.25}}
{"type":"result","timestamp":"2026-11-11T22:19:45Z","ping":{"jitter":4.967,"latency":20.664},"download":{"bandwidth":25000000},"upload":{"bandwidth":5000000},"packetLoss":0,"isp":"Virgin Media","server":{"id":76429,"name":"Virgin Media O2"}}
`

	var updates []Progress
	o, err := readStream(strings.NewReader(sample), func(p Progress) { updates = append(updates, p) })
	if err != nil {
		t.Fatalf("readStream: %v", err)
	}
	if o.Type != "result" || bandwidthToMbps(o.Download.Bandwidth) != 200 {
		t.Errorf("result = %+v", o)
	}

	if len(updates) != 4 {
		t.Fatalf("got %d progress updates, want 4", len(updates))
	}
	start := updates[0]
	if start.Phase != "start" || start.ServerName != "Virgin Media O2" || start.ServerLocation != "Yate, United Kingdom" {
		t.Errorf("start = %+v", start)
	}
	if start.PingMs != nil || start.DownloadMbps != nil {
		t.Errorf("start should have no metrics yet: %+v", start)
	}

	// Each snapshot keeps earlier phases' values.
	last := updates[3]
	if last.Phase != "upload" || last.PhaseProgress != 0.25 {
		t.Errorf("last phase = %q %v", last.Phase, last.PhaseProgress)
	}
	if last.PingMs == nil || *last.PingMs != 21.077 || last.DownloadMbps == nil || *last.DownloadMbps != 100 || last.UploadMbps == nil || *last.UploadMbps != 20 {
		t.Errorf("last = %+v", last)
	}

	t.Run("missing result", func(t *testing.T) {
		if _, err := readStream(strings.NewReader(`{"type":"testStart"}`), nil); err == nil {
			t.Error("expected error when output has no result line")
		}
	})

	t.Run("logged error", func(t *testing.T) {
		out := `{"type":"testStart"}
{"type":"log","timestamp":"2026-11-11T22:19:17Z","message":"Cannot open socket: Timeout occurred in connect.","level":"error"}
`
		_, err := readStream(strings.NewReader(out), nil)
		if err == nil || !strings.Contains(err.Error(), "Cannot open socket") {
			t.Errorf("err = %v, want the logged CLI message", err)
		}
	})
}

func TestParseServers(t *testing.T) {
	sample := `{"type":"serverList","timestamp":"2026-08-08T20:22:08Z","servers":[
		{"id":45909,"host":"speedtest.ogi.wales","port":8080,"name":"Ogi","location":"Cardiff","country":"United Kingdom"},
		{"id":28459,"host":"speedtest.oxford.oxide.ox.ac.uk","port":8080,"name":"University of Oxford","location":"Oxford","country":"United Kingdom"}
	]}`

	servers, err := parseServers([]byte(sample))
	if err != nil {
		t.Fatalf("parseServers: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("got %d servers, want 2", len(servers))
	}
	s := servers[0]
	if s.ID != 45909 || s.Name != "Ogi" || s.Location != "Cardiff" || s.Host != "speedtest.ogi.wales" || s.Port != 8080 {
		t.Errorf("first server = %+v", s)
	}
}
