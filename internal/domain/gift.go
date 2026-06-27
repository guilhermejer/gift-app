package domain

type Gift struct {
	GiftID      string
	FriendID    string
	Title       string
	Description string
	PriceRange  string
	Tags        []string
}
