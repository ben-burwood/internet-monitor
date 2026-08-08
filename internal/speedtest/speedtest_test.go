package speedtest

import "testing"

func TestBuildArgs(t *testing.T) {
	base := []string{"--format=json", "--accept-license", "--accept-gdpr"}

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
