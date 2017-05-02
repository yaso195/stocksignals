package model

const (
	// BUY represents stock buy order.
	BUY = "buy"

	// SELL represents stock sell order.
	SELL = "sell"

	// DEPOSIT represents the deposit order type
	DEPOSIT = "deposit"

	// WITHDRAW represents the withdraw order type
	WITHDRAW = "withdraw"
)

type Signal struct {
	ID             int     `json:"id" db:"id"`
	Name           string  `json:"name" binding:"required" db:"name"`
	Description    string  `json:"description,omitempty" db:"description"`
	NumSubscribers int     `json:"num_subscribers" db:"num_subscribers"`
	NumTrades      int     `json:"num_trades" db:"num_trades"`
	Price          float64 `json:"price" binding:"required" db:"price"`
	Growth         float64 `json:"growth" db:"growth"`
}

type Holding struct {
	ID        int     `json:"id" db:"id"`
	SignalID  int     `json:"signal_id" db:"signal_id"`
	Code      string  `json:"code" binding:"required" db:"code"`
	Name      string  `json:"name,omitempty" db:"name"`
	NumShares int     `json:"num_shares" db:"num_shares"`
	Price     float64 `json:"price" binding:"required" db:"price"`
}

type Order struct {
	ID        int     `json:"id" db:"id"`
	SignalID  int     `json:"signal_id" binding:"required" db:"signal_id"`
	Time      int64   `json:"order_time" db:"order_time"`
	Type      string  `json:"type,omitempty" binding:"required" db:"type"`
	Code      string  `json:"code,omitempty" db:"code"`
	Name      string  `json:"name,omitempty" db:"name"`
	NumShares int     `json:"num_shares" db:"num_shares"`
	Price     float64 `json:"price" db:"price"`
	Profit    float64 `json:"profit" db:"profit"`
}

type Stats struct {
	ID          int     `json:"id" db:"id"`
	Deposits    float64 `json:"deposits" db:"deposits"`
	Withdrawals float64 `json:"withdrawals" db:"withdrawals"`
	Funds       float64 `json:"funds" db:"funds"`
	Profit      float64 `json:"profit" db:"profit"`
	NumTrades   int     `json:"num_trades" db:"num_trades"`
}
