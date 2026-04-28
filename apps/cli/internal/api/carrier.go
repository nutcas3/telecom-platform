package api

// Carrier Connector Integration
type CarrierInfo struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	NetworkType string `json:"network_type"`
	Status      string `json:"status"`
}

type ConnectivityStatus struct {
	Connected bool   `json:"connected"`
	Latency   string `json:"latency"`
	Message   string `json:"message"`
}

type ProfileInfo struct {
	ProfileID      string `json:"profile_id"`
	IMSI           string `json:"imsi"`
	Status         string `json:"status"`
	ActivationCode string `json:"activation_code"`
}

type OrderProfileRequest struct {
	IMSI      string `json:"imsi"`
	ProfileID string `json:"profile_id"`
}

// GetCarrierInfo retrieves carrier information
func (c *Client) GetCarrierInfo() (*CarrierInfo, error) {
	var info CarrierInfo
	if err := c.doGetJSON("/api/v1/carrier/info", &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// CheckConnectivity checks carrier connectivity
func (c *Client) CheckConnectivity() (*ConnectivityStatus, error) {
	var status ConnectivityStatus
	if err := c.doGetJSON("/api/v1/carrier/connectivity", &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// ListProfiles lists eSIM profiles
func (c *Client) ListProfiles() ([]ProfileInfo, error) {
	var resp struct {
		Data []ProfileInfo `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/esim/profiles", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// OrderProfile orders a new eSIM profile
func (c *Client) OrderProfile(req *OrderProfileRequest) (*ProfileInfo, error) {
	var profile ProfileInfo
	if err := c.doPostJSON("/api/v1/esim/profiles", req, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// DeleteProfile deletes an eSIM profile
func (c *Client) DeleteProfile(profileID string) error {
	return c.doDelete("/api/v1/esim/profiles/" + profileID)
}
