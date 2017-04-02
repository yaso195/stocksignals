package model

type Signal struct {
	ID             int     `json:"id" db:"id"`
	Name           string  `json:"name" db:"name"`
	Description    string  `json:"description,omitempty" db:"description"`
	NumSubscribers int     `json:"num_subscribers" db:"num_subscribers"` // How many instance managers
	NumTrades      int     `json:"num_trades" db:"num_trades"`           // Number of central instance
	Price          float64 `json:"price" db:"price"`                     // How many storage servers
	Growth         float64 `json:"growth" db:"growth"`
}
