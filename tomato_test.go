package tomato

import (
	"github.com/freeznet/tomato/orm"
	"github.com/freeznet/tomato/storage"
	"github.com/freeznet/tomato/storage/postgres"
	"github.com/freeznet/tomato/test"
	"testing"
)

func TestRun(t *testing.T) {
	Run()
}

func initPostgresEnv() {
	orm.InitOrm(getPostgresAdapter())
}

func getPostgresAdapter() storage.Adapter {
	return postgres.NewPostgresAdapter("tomato", test.OpenPostgreSQForTest())
}
