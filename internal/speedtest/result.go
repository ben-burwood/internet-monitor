package speedtest

import "time"

// Result is a single speedtest measurement, stored and served to the frontend.
//
// The numeric metrics are pointers so a failed run (Success == false) serialises
// them as JSON null, which the charts render as gaps rather than zeros.
type Result struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`

	DownloadMbps  *float64 `json:"download_mbps"`
	UploadMbps    *float64 `json:"upload_mbps"`
	PingMs        *float64 `json:"ping_ms"`
	JitterMs      *float64 `json:"jitter_ms"`
	PacketLossPct *float64 `json:"packet_loss_pct"`

	DurationMs *float64 `json:"duration_ms"`

	// Data transferred during the download and upload phases.
	DownloadBytes *int64 `json:"download_bytes"`
	UploadBytes   *int64 `json:"upload_bytes"`

	ServerID       int    `json:"server_id,omitempty"`
	ServerName     string `json:"server_name,omitempty"`
	ServerLocation string `json:"server_location,omitempty"`
	ISP            string `json:"isp,omitempty"`
	ExternalIP     string `json:"external_ip,omitempty"`
	ResultURL      string `json:"result_url,omitempty"`
}

// ooklaResult mirrors a line emitted by `speedtest -f jsonl`. The final line
// (type "result") is the full result; earlier lines (testStart, ping, download,
// upload) are progress updates that carry a subset of the same fields. "log"
// lines carry a message and level, which is how the CLI reports some errors.
type ooklaResult struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`   // log lines only
	Message   string    `json:"message"` // log lines only
	Ping      struct {
		Jitter   float64 `json:"jitter"`
		Latency  float64 `json:"latency"`
		Progress float64 `json:"progress"` // 0–1, progress lines only
	} `json:"ping"`
	Download   ooklaTransfer `json:"download"`
	Upload     ooklaTransfer `json:"upload"`
	PacketLoss float64       `json:"packetLoss"`
	ISP        string        `json:"isp"`
	Interface  struct {
		ExternalIP string `json:"externalIp"`
	} `json:"interface"`
	Server struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Location string `json:"location"`
		Country  string `json:"country"`
	} `json:"server"`
	Result struct {
		URL string `json:"url"`
	} `json:"result"`
}

// ooklaTransfer is the download/upload block of an Ookla payload.
type ooklaTransfer struct {
	Bandwidth float64 `json:"bandwidth"` // bytes per second
	Bytes     int64   `json:"bytes"`     // total transferred
	Progress  float64 `json:"progress"`  // 0–1, progress lines only
}

// Progress is a live snapshot of a running speedtest. A metric stays nil until
// its phase starts, then holds the latest interim value.
type Progress struct {
	// Phase is one of "start", "ping", "download" or "upload".
	Phase string `json:"phase"`
	// PhaseProgress is how far through the current phase the test is (0–1).
	PhaseProgress float64 `json:"phase_progress"`

	DownloadMbps *float64 `json:"download_mbps"`
	UploadMbps   *float64 `json:"upload_mbps"`
	PingMs       *float64 `json:"ping_ms"`
	JitterMs     *float64 `json:"jitter_ms"`

	ServerName     string `json:"server_name,omitempty"`
	ServerLocation string `json:"server_location,omitempty"`
	ISP            string `json:"isp,omitempty"`
}

// apply folds a progress line into the snapshot. It returns false for line
// types that are not progress updates (e.g. the final result).
func (p *Progress) apply(o ooklaResult) bool {
	switch o.Type {
	case "testStart":
		p.Phase, p.PhaseProgress = "start", 0
		p.ServerName = o.Server.Name
		p.ServerLocation = o.serverLocation()
		p.ISP = o.ISP
	case "ping":
		p.Phase, p.PhaseProgress = "ping", o.Ping.Progress
		ping, jitter := o.Ping.Latency, o.Ping.Jitter
		p.PingMs, p.JitterMs = &ping, &jitter
	case "download":
		p.Phase, p.PhaseProgress = "download", o.Download.Progress
		dl := bandwidthToMbps(o.Download.Bandwidth)
		p.DownloadMbps = &dl
	case "upload":
		p.Phase, p.PhaseProgress = "upload", o.Upload.Progress
		ul := bandwidthToMbps(o.Upload.Bandwidth)
		p.UploadMbps = &ul
	default:
		return false
	}
	return true
}

// serverLocation joins the server's location and country, e.g. "Yate, United Kingdom".
func (o ooklaResult) serverLocation() string {
	location := o.Server.Location
	if location != "" && o.Server.Country != "" {
		location = location + ", " + o.Server.Country
	}
	return location
}

// bandwidthToMbps converts Ookla's bytes-per-second bandwidth to megabits/sec.
func bandwidthToMbps(bytesPerSec float64) float64 {
	return bytesPerSec * 8 / 1e6
}

// toResult maps a parsed Ookla payload into our stored Result.
func (o ooklaResult) toResult(id string) Result {
	dl := bandwidthToMbps(o.Download.Bandwidth)
	ul := bandwidthToMbps(o.Upload.Bandwidth)
	ping := o.Ping.Latency
	jitter := o.Ping.Jitter
	loss := o.PacketLoss
	dlBytes := o.Download.Bytes
	ulBytes := o.Upload.Bytes

	return Result{
		ID:             id,
		Timestamp:      o.Timestamp,
		Success:        true,
		DownloadMbps:   &dl,
		UploadMbps:     &ul,
		PingMs:         &ping,
		JitterMs:       &jitter,
		PacketLossPct:  &loss,
		DownloadBytes:  &dlBytes,
		UploadBytes:    &ulBytes,
		ServerID:       o.Server.ID,
		ServerName:     o.Server.Name,
		ServerLocation: o.serverLocation(),
		ISP:            o.ISP,
		ExternalIP:     o.Interface.ExternalIP,
		ResultURL:      o.Result.URL,
	}
}
