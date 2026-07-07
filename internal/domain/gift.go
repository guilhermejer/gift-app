package domain

const (
	GiftTypeGift   = "gift"
	GiftTypeOuting = "outing"
)

func IsValidGiftType(t string) bool {
	return t == GiftTypeGift || t == GiftTypeOuting
}

type Gift struct {
	GiftID          string   `json:"giftID"`
	FriendID        string   `json:"friendID"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	PriceRange      string   `json:"priceRange"`
	Tags            []string `json:"tags"`
	Type            string   `json:"type,omitempty"`
	OccasionDetails string   `json:"occasionDetails,omitempty"`
	ReminderID      string   `json:"reminderID,omitempty"`
}
