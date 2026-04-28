package api

// Charging Engine Integration
type CreditCheckRequest struct {
	BytesRequested uint64 `json:"bytes_requested"`
}

type CreditCheckResponse struct {
	Allowed   bool   `json:"allowed"`
	Available uint64 `json:"available"`
	Requested uint64 `json:"requested"`
	Remaining uint64 `json:"remaining"`
}

type CreditBalanceResponse struct {
	Balance uint64 `json:"balance"`
}

type CreditAddRequest struct {
	BytesToAdd uint64 `json:"bytes_to_add"`
}

type CreditDeductRequest struct {
	BytesUsed uint64 `json:"bytes_used"`
}

// CheckCredit checks if a subscriber has enough credit
func (c *Client) CheckCredit(ip string, bytesRequested uint64) (*CreditCheckResponse, error) {
	var resp CreditCheckResponse
	req := CreditCheckRequest{BytesRequested: bytesRequested}
	if err := c.doPostJSON("/v1/charging/credit/"+ip+"/check", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetBalance gets the current credit balance for a subscriber
func (c *Client) GetBalance(ip string) (*CreditBalanceResponse, error) {
	var resp CreditBalanceResponse
	if err := c.doGetJSON("/v1/charging/credit/"+ip+"/balance", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddCredit adds credit to a subscriber
func (c *Client) AddCredit(ip string, bytesToAdd uint64) (*CreditBalanceResponse, error) {
	var resp CreditBalanceResponse
	req := CreditAddRequest{BytesToAdd: bytesToAdd}
	if err := c.doPostJSON("/v1/charging/credit/"+ip+"/add", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeductCredit deducts credit from a subscriber
func (c *Client) DeductCredit(ip string, bytesUsed uint64) (*CreditBalanceResponse, error) {
	var resp CreditBalanceResponse
	req := CreditDeductRequest{BytesUsed: bytesUsed}
	if err := c.doPostJSON("/v1/charging/credit/"+ip+"/deduct", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
