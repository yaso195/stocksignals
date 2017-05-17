package store

import (
	"database/sql"
	"fmt"

	"github.com/heroku/stocksignals/model"
)

//  GetHoldingsBySignalID reads the holdings from the database based on the given signal id and orders them based on the given field
func GetHoldingsBySignalID(signalID int, field string, descend bool) ([]model.Holding, error) {
	if db == nil {
		return nil, fmt.Errorf("no connection is created to the database")
	}

	if field == "" {
		field = DEFAULT_HOLDING_FIELD
	}

	order := "DESC"
	if !descend {
		order = "ASC"
	}

	var holdings []model.Holding
	err := db.Select(&holdings, fmt.Sprintf("SELECT * FROM holdings WHERE signal_id = %d ORDER BY %s %s", signalID, field, order))
	if err != nil {
		return nil, fmt.Errorf("error reading holdings: %q", err)
	}

	return holdings, nil
}

// getHolding reads the holding from the database based on the given signal id and stock code
func getHolding(signalID int, code string) (model.Holding, error) {
	if db == nil {
		return model.Holding{}, fmt.Errorf("no connection is created to the database")
	}

	var result model.Holding
	query := fmt.Sprintf("SELECT * FROM holdings WHERE signal_id = %d and code = '%s'", signalID, code)
	err := db.Get(&result, query)
	if err == sql.ErrNoRows {
		return model.Holding{}, nil
	}

	if err != nil {
		return model.Holding{}, fmt.Errorf("error reading holding: %q", err)
	}

	return result, nil
}
