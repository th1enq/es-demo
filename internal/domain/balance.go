package domain

type Balance struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}
