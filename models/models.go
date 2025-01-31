package models

type UtilACcountCreds struct {
	Username string `json:"username"`
	Password string `json:"password"`
	DID      string `json:"did"`
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
	DID            string     `json:"did"`
	DisplayName    string     `json:"displayName"`
	Handle         string     `json:"handle"`
	Description    string     `json:"description,omitempty"`
	Avatar         string     `json:"avatar,omitempty"`
	Banner         string     `json:"banner,omitempty"`
	FollowersCount int        `json:"followersCount,omitempty"`
	FollowsCount   int        `json:"followsCount,omitempty"`
	PostsCount     int        `json:"postsCount,omitempty"`
	PinnedPost     PinnedPost `json:"pinnedPost,omitempty"`
	IndexedAt      string     `json:"indexedAt,omitempty"`
	CreatedAt      string     `json:"createdAt,omitempty"`
	Viewer         Viewer     `json:"viewer,omitempty"`
	Labels         []Labels   `json:"labels,omitempty"`
}

type Viewer struct {
	Muted       bool    `json:"muted"`
	MutedByList []Muter `json:"mutedByList,omitempty"`
}

type Muter struct {
	URI           string   `json:"uri,omitempty"`
	CID           string   `json:"cid,omitempty"`
	Name          string   `json:"name,omitempty"`
	Avatar        string   `json:"avatar,omitempty"`
	ListItemCount int      `json:"listItemCount,omitempty"`
	Labels        []Labels `json:"labels,omitempty"`
	Viewer        Viewer   `json:"viewer,omitempty"`
	IndexedAt     string   `json:"indexedAt,omitempty"`
}

type PinnedPost struct {
	URI string `json:"uri,omitempty"`
	CID string `json:"cid,omitempty"`
}

type Associated struct {
	Lists        int  `json:"lists,omitempty"`
	Feedgens     int  `json:"feedgens,omitempty"`
	StarterPacks int  `json:"starterPacks,omitempty"`
	Labeler      bool `json:"labeler,omitempty"`
	Chat
}

type Chat struct {
	AllowIncoming string `json:"allowIncoming,omitempty"`
}

type JoinedViaStarterPack struct {
	URI string `json:"uri,omitempty"`
	CID string `json:"cid,omitempty"`
}

type Creator struct {
	DID         string     `json:"did,omitempty"`
	Handle      string     `json:"handle,omitempty"`
	DisplayName string     `json:"displayName,omitempty"`
	Avatar      string     `json:"avatar,omitempty"`
	Associated  Associated `json:"associated,omitempty"`
	Labels      []Labels   `json:"labels,omitempty"`
}

type Labels struct {
	Version   int    `json:"ver,omitempty"`
	Source    string `json:"src,omitempty"`
	URI       string `json:"uri,omitempty"`
	CID       string `json:"cid,omitempty"`
	Val       string `json:"val,omitempty"`
	Neg       bool   `json:"neg,omitempty"`
	CTS       string `json:"cts,omitempty"`
	Exp       string `json:"exp,omitempty"`
	Signature byte   `json:"sig,omitempty"`
}
