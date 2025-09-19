package server

import (
	"optim/cmd/db"
	optim "optim/internal/orchestrator"
)

func InitOptimServer(datastore *db.DataStore) (*optim.Optim, error) {

	optim := optim.NewOptim(datastore.Queries, datastore.Redis, datastore.PgPool)
	err := optim.InitOptim()
	if err != nil {
		return nil, err
	}

	return optim, nil
}
