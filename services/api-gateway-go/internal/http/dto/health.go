package dto

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type UpstreamHealth struct {
	Name       string `json:"name"`
	OK         bool   `json:"ok"`
	StatusCode int    `json:"statusCode,omitempty"`
	Error      string `json:"error,omitempty"`
}

type UpstreamsHealthResponse struct {
	Status   string           `json:"status"`
	Service  string           `json:"service"`
	Upstream []UpstreamHealth `json:"upstream"`
}
