package api

// Subscriber represents a subscriber from the API
type Subscriber struct {
	ID        uint   `json:"id"`
	IMSI      string `json:"imsi"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Status    string `json:"status"`
}

// SubscriberAccount holds account info
type SubscriberAccount struct {
	IMSI    string  `json:"imsi"`
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Balance float64 `json:"balance"`
}

// ListSubscribers retrieves subscribers
func (c *Client) ListSubscribers() ([]SubscriberAccount, error) {
	var resp struct {
		Data []SubscriberAccount `json:"data"`
	}
	if err := c.doGetJSON("/api/v1/subscribers", &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetSubscriber retrieves a single subscriber
func (c *Client) GetSubscriber(imsi string) (*SubscriberAccount, error) {
	var sub SubscriberAccount
	if err := c.doGetJSON("/api/v1/subscribers/"+imsi, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// CreateSubscriber creates a subscriber
func (c *Client) CreateSubscriber(imsi, name string) (*SubscriberAccount, error) {
	var sub SubscriberAccount
	body := map[string]string{"imsi": imsi, "name": name}
	if err := c.doPostJSON("/api/v1/subscribers", body, &sub); err != nil {
		return nil, err
	}
	return &sub, nil
}

// DeleteSubscriber deletes a subscriber
func (c *Client) DeleteSubscriber(imsi string) error {
	return c.doDelete("/api/v1/subscribers/" + imsi)
}
