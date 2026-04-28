package repository

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a profile lookup misses.
var ErrNotFound = errors.New("profile not found")

// Profile is the stored representation of an eSIM profile.
type Profile struct {
	ICCID          string    `json:"iccid"`
	EID            string    `json:"eid,omitempty"`
	IMSI           string    `json:"imsi"`
	MCC            string    `json:"mcc"`
	MNC            string    `json:"mnc"`
	ProfileType    string    `json:"profileType"`
	State          string    `json:"state"`
	TenantID       string    `json:"tenantId,omitempty"`
	ActivationCode string    `json:"activationCode,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// ListFilter narrows the List query.
type ListFilter struct {
	TenantID string
	State    string
	Limit    int
	Offset   int
}

// ProfileRepository is the storage contract for eSIM profiles.
type ProfileRepository interface {
	Create(ctx context.Context, p *Profile) error
	Get(ctx context.Context, iccid string) (*Profile, error)
	List(ctx context.Context, f ListFilter) ([]*Profile, int, error)
	UpdateState(ctx context.Context, iccid, state string) (*Profile, error)
	Delete(ctx context.Context, iccid string) error
	Ping() error
}
