package models

type UtilACcountCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DID      string `json:"did"`
}

type ProfileRequest struct {
	DID string `json:"did"`
}

type SessionRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type SessionResponse struct {
	AccessJwt string `json:"accessJwt"`
	Did       string `json:"did"`
	Handle    string `json:"handle"`
}

type ProfileResponse struct {
	DisplayName string `json:"displayName"`
	Handle      string `json:"handle"`
	Avatar      string `json:"avatar,omitempty"`
	Description string `json:"description,omitempty"`
}
