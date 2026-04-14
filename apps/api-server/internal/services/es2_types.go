package services

// ES2 request/response types
type DownloadOrderRequest struct {
	EID          string            `json:"eid"`
	ICCID        string            `json:"iccid"`
	ProfileType  string            `json:"profileType"`
	Confirmation bool              `json:"confirmation"`
	Metadata     map[string]string `json:"metadata"`
}

type DownloadOrderResponse struct {
	ICCID              string `json:"iccid"`
	ProfileID          string `json:"profileId"`
	ActivationCode     string `json:"activationCode"`
	ConfirmationAddress string `json:"confirmationAddress"`
}

type ActivationRequest struct {
	EID       string `json:"eid"`
	ProfileID string `json:"profileId"`
}

type DeactivationRequest struct {
	EID       string `json:"eid"`
	ProfileID string `json:"profileId"`
}

type DeletionRequest struct {
	EID       string `json:"eid"`
	ProfileID string `json:"profileId"`
}

type ProfileStatusResponse struct {
	ICCID       string `json:"iccid"`
	ProfileID   string `json:"profileId"`
	ProfileName string `json:"profileName"`
	State       string `json:"state"`
	Operator    string `json:"operator"`
}

type ListProfilesResponse struct {
	Profiles []ProfileStatusResponse `json:"profiles"`
}

// ProfileInfo represents eSIM profile information
type ProfileInfo struct {
	ICCID       string            `json:"iccid"`
	ProfileID   string            `json:"profileId"`
	ProfileName string            `json:"profileName"`
	State       string            `json:"state"`
	Operator    string            `json:"operator"`
	Activation  ProfileActivation `json:"activation"`
}

// ProfileActivation represents profile activation details
type ProfileActivation struct {
	ActivationCode string `json:"activationCode"`
	ConfAddress    string `json:"confirmationAddress"`
}
