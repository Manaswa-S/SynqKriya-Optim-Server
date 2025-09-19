package server

import (
	"midoptim/cmd/db"
	optim "midoptim/internal/orchestrator"
)

func InitOptimServer(datastore *db.DataStore) (*optim.Optim, error) {

	optim := optim.NewOptim(datastore.Queries, datastore.Redis, datastore.PgPool)
	err := optim.InitOptim()
	if err != nil {
		return nil, err
	}

	return optim, nil
}
