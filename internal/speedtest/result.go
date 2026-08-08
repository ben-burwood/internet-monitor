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

	ServerID       int    `json:"server_id,omitempty"`
	ServerName     string `json:"server_name,omitempty"`
	ServerLocation string `json:"server_location,omitempty"`
	ISP            string `json:"isp,omitempty"`
	ExternalIP     string `json:"external_ip,omitempty"`
	ResultURL      string `json:"result_url,omitempty"`
}

// ooklaResult mirrors the JSON emitted by `speedtest -f json`.
type ooklaResult struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Ping      struct {
		Jitter  float64 `json:"jitter"`
		Latency float64 `json:"latency"`
	} `json:"ping"`
	Download struct {
		Bandwidth float64 `json:"bandwidth"` // bytes per second
	} `json:"download"`
	Upload struct {
		Bandwidth float64 `json:"bandwidth"` // bytes per second
	} `json:"upload"`
	PacketLoss float64 `json:"packetLoss"`
	ISP        string  `json:"isp"`
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

	location := o.Server.Location
	if location != "" && o.Server.Country != "" {
		location = location + ", " + o.Server.Country
	}

	return Result{
		ID:             id,
		Timestamp:      o.Timestamp,
		Success:        true,
		DownloadMbps:   &dl,
		UploadMbps:     &ul,
		PingMs:         &ping,
		JitterMs:       &jitter,
		PacketLossPct:  &loss,
		ServerID:       o.Server.ID,
		ServerName:     o.Server.Name,
		ServerLocation: location,
		ISP:            o.ISP,
		ExternalIP:     o.Interface.ExternalIP,
		ResultURL:      o.Result.URL,
	}
}
