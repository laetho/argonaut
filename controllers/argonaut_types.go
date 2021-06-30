package controllers

// Struct for generating Secret payload used in a Pod to run a tunnel.
type ArgonautTunnelSecret struct {
	AccountTag   string `json:"AccountTag"`
	TunnelSecret string `json:"TunnelSecret"`
	TunnelID     string `json:"TunnelID"`
	TunnelName   string `json:"TunnelName"`
}

// Struct for generating ConfigMap payload to be used in a Pod to run a tunnel
type ArgonautTunnelConfig struct {
	Tunnel          string                        `json:"tunnel"`
	CredentialsFile string                        `json:"credentials-file"`
	Ingress         []ArgonautTunnelConfigIngress `json:"ingress"`
}

// Struct for holding ingress information
type ArgonautTunnelConfigIngress struct {
	Hostname string `json:"hostname,omitempty"`
	Service  string `json:"service,omitempty"`
}
