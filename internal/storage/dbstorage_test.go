package storage

import (
	"database/sql"
	"encoding/json"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // Импортируем драйвер pgx
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorage_Postgres_TemporaryTable(t *testing.T) {
	dsn := "host=localhost port=5432 user=postgres password=12345 dbname=testdb sslmode=disable"

	dbConn, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer dbConn.Close()

	createTempTable := `
    CREATE TEMP TABLE metrix (
        id VARCHAR(32) PRIMARY KEY,
        value JSONB
    );
    `
	_, err = dbConn.Exec(createTempTable)
	require.NoError(t, err)

	storage := &DBStorage{
		DBDSN: dsn,
		db:    dbConn,
	}

	err = storage.NewStorage()
	require.NoError(t, err)

	_, err = storage.saveGaugeStmt.Exec("test_id", "gauge", 42.0)
	assert.NoError(t, err)

	_, err = storage.saveCounterStmt.Exec("test_id", "counter", int64(100))
	assert.NoError(t, err)

	_, err = storage.saveGaugeStmt.Exec("test_gauge", "gauge", 42.0)
	require.NoError(t, err)

	var valueJSON string
	err = storage.getStmt.QueryRow("test_gauge").Scan(&valueJSON)
	assert.NoError(t, err)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(valueJSON), &data)
	assert.NoError(t, err)

	assert.Equal(t, "test_gauge", data["id"])
	assert.Equal(t, "gauge", data["type"])
	assert.Equal(t, 42.0, data["value"])
}

func TestNewStorage_Postgres_ErrorHandling(t *testing.T) {
	dsn := "invalid_dsn"

	storage := &DBStorage{
		DBDSN: dsn,
	}

	err := storage.NewStorage()
	assert.Error(t, err)
}
