package domain

type Gift struct {
	GiftID          string   `json:"giftID"`
	FriendID        string   `json:"friendID"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	PriceRange      string   `json:"priceRange"`
	Tags            []string `json:"tags"`
	OccasionDetails string   `json:"occasionDetails,omitempty"`
	ReminderID      string   `json:"reminderID,omitempty"`
}
