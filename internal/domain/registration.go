package domain

type AgentRegistrationRequest struct {
	AgentName    string `json:"agentName"`
	IPAddress    string `json:"ipAddress"`
	AgentCountry string `json:"agentCountry"`
}

type AgentRegistrationResponse struct {
	Token string `json:"token"`
}
