package storage

import (
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // Импортируем драйвер pgx
	"github.com/stretchr/testify/assert"
)

/*func TestNewStorage_Postgres_TemporaryTable(t *testing.T) {
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
}*/

func TestNewStorage_Postgres_ErrorHandling(t *testing.T) {
	dsn := "invalid_dsn"

	storage := &DBStorage{
		DBDSN: dsn,
	}

	err := storage.NewStorage()
	assert.Error(t, err)
}

/*func BenchmarkSave(b *testing.B) {

	dsn := "host=localhost port=5432 user=postgres password=12345 dbname=testdb sslmode=disable"

	storage := &DBStorage{
		DBDSN: dsn,
	}
	err := storage.NewStorage()
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}

	gaugeValue := service.GaugeMetricValue(42.0)
	metric := service.Metrics{
		ID:    "test_gauge",
		MType: service.GaugeMetric,
		Value: &gaugeValue,
	}

	ctx := context.TODO()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := storage.Save(ctx, metric)
		if err != nil {
			b.Fatalf("Failed to save metric: %v", err)
		}
	}

}*/

/*func BenchmarkSaveAll(b *testing.B) {

	dsn := "host=localhost port=5432 user=postgres password=12345 dbname=testdb sslmode=disable"

	storage := &DBStorage{
		DBDSN: dsn,
	}
	err := storage.NewStorage()
	if err != nil {
		b.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.db.Close()

	metrics := make([]service.Metrics, 0, 1000)
	for i := 0; i < 1000; i++ {
		gaugeValue := service.GaugeMetricValue(float64(i))
		metric := service.Metrics{
			ID:    fmt.Sprintf("test_metric_%d", i),
			MType: service.GaugeMetric,
			Value: &gaugeValue,
		}
		metrics = append(metrics, metric)
	}

	ctx := context.TODO()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := storage.SaveAll(ctx, &metrics)
		if err != nil {
			b.Fatalf("Failed to save metrics: %v", err)
		}
	}
}
*/
