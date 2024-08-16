package gosporestack

import (
	"encoding/json"
)

// ServerService manages server API actions
type ServerService struct {
	client *Client
}

// Machine represents the machine model with various fields.
type Machine struct {
	MachineID       string  `json:"machine_id" validate:"len=32, max=64"`                                  // Machine ID with length constraints
	CreatedAt       int     `json:"created_at" validate:"gte=0"`                                           // Creation timestamp
	Expiration      int     `json:"expiration" validate:"gte=0"`                                           // Expiration timestamp
	Token           string  `json:"token" validate:"len=32, max=64"`                                       // Token with length constraints
	Region          string  `json:"region"`                                                                // Region name
	IPv4            string  `json:"ipv4"`                                                                  // IPv4 address
	IPv6            string  `json:"ipv6"`                                                                  // IPv6 address
	DeletedAt       int     `json:"deleted_at" validate:"gte=0"`                                           // Deletion timestamp
	DeletedBy       *string `json:"deleted_by"`                                                            // User who deleted (nullable)
	ForgottenAt     *string `json:"forgotten_at"`                                                          // Forgotten timestamp (nullable)
	SuspendedAt     *string `json:"suspended_at"`                                                          // Suspension timestamp (nullable)
	Provider        string  `json:"provider"`                                                              // Provider name
	Running         bool    `json:"running"`                                                               // Machine status
	DenySMTP        bool    `json:"deny_smtp"`                                                             // SMTP access denial
	FlavorSlug      string  `json:"flavor_slug" validate:"len=3, max=64"`                                  // Flavor slug with length constraints
	OperatingSystem string  `json:"operating_system"`                                                      // Operating system name
	Hostname        string  `json:"hostname" validate:"omitempty, max=128, pattern=^$|^[a-zA-Z0-9-_. ]+$"` // Hostname with constraints
	Autorenew       bool    `json:"autorenew"`                                                             // Auto-renew flag
	Flavor          *Flavor `json:"flavor"`                                                                // Flavor object (read-only)
}

// Flavor represents the flavor model with various fields.
type Flavor struct {
	Slug              string  `json:"slug"`
	Cores             int     `json:"cores"`
	Memory            int     `json:"memory"`
	Disk              int     `json:"disk"`
	Price             int     `json:"price"`
	IPv4              string  `json:"ipv4"`
	IPv6              string  `json:"ipv6"`
	Bandwidth         int     `json:"bandwidth"` // Bandwidth in gigabytes per month
	BandwidthPerMonth float64 `json:"bandwidth_per_month"`
	ProviderSlug      string  `json:"provider_slug"` // Provider slug for the flavor
	Provider          string  `json:"provider"`
}

// ServerLaunchRequest represents the request structure for server configuration.
type ServerLaunchRequest struct {
	Flavor          string `json:"flavor"`           // Flavor of the server
	SSHKey          string `json:"ssh_key"`          // SSH key for authentication
	OperatingSystem string `json:"operating_system"` // Operating system name
	Provider        string `json:"provider"`         // Cloud provider
	Autorenew       bool   `json:"autorenew"`        // Flag to indicate if auto-renew is enabled
	Days            int    `json:"days"`             // Number of days the server is requested for
	Region          string `json:"region"`           // Region where the server will be deployed
	Hostname        string `json:"hostname"`         // Hostname for the server
	UserData        string `json:"user_data"`        // Additional user data or configuration
}

type ServerLaunchResponse struct {
	MachineID string `json:"machine_id" validate:"len=32, max=64"`
}

type QuoteResponse struct {
	Cents int    `json:"cents"`
	Usd   string `json:"usd" validate:"min=5"`
}

type TopUpRequest struct {
	Days  int     `json:"days"`  // Number of days to top up, must be between 1 and 90
	Token *string `json:"token"` // Optional token associated with the server (or null)
}

type UpdateRequest struct {
	HostName  string `json:"hostname" validate:"omitempty,min=1,max=255"`
	AutoRenew bool   `json:"autorenew"`
}

// Launch launch a server
func (ss *ServerService) Launch(reqq *ServerLaunchRequest) (*ServerLaunchResponse, error) {
	body, err := json.Marshal(reqq)
	if err != nil {
		return nil, err
	}

	req, err := ss.client.NewRequest("POST", "/token/"+ss.client.token+"/servers", body)
	if err != nil {
		return nil, err
	}

	s := ServerLaunchResponse{}
	if err := ss.client.DoRequest(req, &s); err != nil {
		return nil, err
	}

	return &s, nil
}

// List Info on all servers launched by a given token.
func (ss *ServerService) List() ([]Machine, error) {
	req, err := ss.client.NewRequest("GET", "/token"+ss.client.token+"/servers", nil)
	if err != nil {
		return nil, err
	}

	var s []Machine
	if err := ss.client.DoRequest(req, &s); err != nil {
		return nil, err
	}

	return s, nil
}

// Quote Gives a quote for a new server.
func (ss *ServerService) Quote(days, flavor, provider string) (*QuoteResponse, error) {
	req, err := ss.client.NewRequest("GET",
		"/server/quote?"+"days="+days+"&flavor_slug="+flavor+"&provider="+provider,
		nil)
	if err != nil {
		return nil, err
	}

	var q = QuoteResponse{}
	if err := ss.client.DoRequest(req, &q); err != nil {
	}

	return &q, nil
}

// TopUp Renew an existing server. Also consider using autorenew.
func (ss *ServerService) TopUp(machineId string, days int) (string, error) {
	topUpRequest := TopUpRequest{
		Days:  days,
		Token: &ss.client.token,
	}

	body, err := json.Marshal(topUpRequest)
	if err != nil {
		return "", err
	}

	req, err := ss.client.NewRequest("POST",
		"/server/"+machineId+"/topup",
		body)
	if err != nil {
		return "", err
	}

	s := ""
	if err := ss.client.DoRequest(req, &s); err != nil {
		return "", err
	}
	return s, nil
}

// Update details about a server.
func (ss *ServerService) Update(machineId, hostname string, autoRenew bool) (string, error) {
	updateRequest := UpdateRequest{
		HostName:  hostname,
		AutoRenew: autoRenew,
	}

	body, err := json.Marshal(updateRequest)
	if err != nil {
		return "", err
	}

	req, err := ss.client.NewRequest("PATCH", "/server/"+machineId, body)
	if err != nil {
		return "", err
	}

	s := ""
	if err := ss.client.DoRequest(req, &s); err != nil {
		return "", err
	}
	return s, nil
}

// Delete the server and refunds an approximate remaining balance to the associated token.
func (ss *ServerService) Delete(machineId string) (string, error) {
	req, err := ss.client.NewRequest("DELETE", "/server/"+machineId, nil)
	if err != nil {
		return "", err
	}
	s := ""
	if err := ss.client.DoRequest(req, &s); err != nil {
		return "", err
	}
	return s, nil
}

// Forget about a deleted server.
func (ss *ServerService) Forget(machineId string) (string, error) {
	req, err := ss.client.NewRequest("POST", "/server/"+machineId+"/forget", nil)
	if err != nil {
		return "", err
	}
	s := ""
	if err := ss.client.DoRequest(req, &s); err != nil {
		return "", err
	}
	return s, nil
}

// Rebuild Rebuilds the server with the same operating system and SSH key provided initially.
// This will delete all data on the server! Will take a couple of minutes after the request is issued.
func (ss *ServerService) Rebuild(machineId string) (string, error) {
	req, err := ss.client.NewRequest("POST", "/server/"+machineId+"/rebuild", nil)
	if err != nil {
		return "", err
	}
	s := ""
	if err := ss.client.DoRequest(req, &s); err != nil {
		return "", err
	}
	return s, nil
}

// Stop Immediately powers off the server.
func (ss *ServerService) Stop(machineId string) (string, error) {
	req, err := ss.client.NewRequest("POST", "/server/"+machineId+"/stop", nil)
	if err != nil {
		return "", err
	}
	s := ""
	if err := ss.client.DoRequest(req, &s); err != nil {
		return "", err
	}
	return s, nil
}
