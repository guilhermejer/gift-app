package domain

type Profile struct {
	FriendID  string    `json:"friendID"`
	Likes     []string  `json:"likes"`
	Dislikes  []string  `json:"dislikes"`
	Embedding []float32 `json:"embedding,omitempty"`
}
