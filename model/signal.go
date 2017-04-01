package model

type Signal struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"name,omitempty"`
	Growth         float64 `json:"growth"`
	NumTrades      int     `json:"num_trades"`      // Number of central instance
	NumSubscribers int     `json:"num_subscribers"` // How many instance managers
	Price          float64 `json:"price"`           // How many storage servers
}
