package api

// Packet Gateway Integration
type PacketStats struct {
	IP    uint64 `json:"ip"`
	Bytes uint64 `json:"bytes"`
}

type PacketCreditInfo struct {
	IP     uint64 `json:"ip"`
	Credit int64  `json:"credit"`
}

// GetPacketStats retrieves packet statistics
func (c *Client) GetPacketStats() ([]PacketStats, error) {
	var resp struct {
		Data []PacketStats `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/packet-gateway/stats", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetPacketCreditInfo retrieves credit information from packet gateway
func (c *Client) GetPacketCreditInfo() ([]PacketCreditInfo, error) {
	var resp struct {
		Data []PacketCreditInfo `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/packet-gateway/credits", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
