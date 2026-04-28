package api

// Service describes a platform service
type Service struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Uptime  string  `json:"uptime"`
	CPU     float64 `json:"cpu"`
	Memory  string  `json:"memory"`
}

// PostRestart requests the API to restart the named service.
func (c *Client) PostRestart(name string) error {
	return c.doPostJSON("/api/v1/services/"+name+"/restart", nil, nil)
}

// ListServices returns platform services
func (c *Client) ListServices() ([]Service, error) {
	var resp struct {
		Data []Service `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/services", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
