package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/dvkhr/metrix.git/internal/service"
	gomock "github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestDBNewStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDB(ctrl)

	storage := &DBStorage{
		DBDSN: "test_dsn",
		db:    mockDB,
	}

	mockDB.EXPECT().Ping().Return(nil)

	createStmt := "create table if not exists metrix (id varchar(32) PRIMARY KEY, value jsonb not null)"
	mockDB.EXPECT().Exec(createStmt).Return(nil, nil)

	saveGaugeQuery := "insert into metrix values($1::varchar, jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'value', $3::double precision)) on conflict(id) do update set value = jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'value', $3::double precision) where metrix.id = $1::varchar;"
	mockDB.EXPECT().Prepare(saveGaugeQuery).Return(&sql.Stmt{}, nil)

	saveCounterQuery := "insert into metrix values($1::varchar, jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'delta', $3::bigint)) on conflict(id) do update set value = jsonb_set(metrix.value, '{delta}', ((metrix.value ->> 'delta')::bigint + $3::bigint)::text::jsonb, false) where metrix.id = $1::varchar;"
	mockDB.EXPECT().Prepare(saveCounterQuery).Return(&sql.Stmt{}, nil)

	getQuery := "select value from metrix where id = $1::varchar;"
	mockDB.EXPECT().Prepare(getQuery).Return(&sql.Stmt{}, nil)

	listQuery := "select jsonb_object_agg(k,v) from metrix, jsonb_each(jsonb_build_object(id, value)) as t(k,v);"
	mockDB.EXPECT().Prepare(listQuery).Return(&sql.Stmt{}, nil)

	err := storage.NewStorage()
	assert.NoError(t, err)
}

func TestDBSave(t *testing.T) {
	gaugeValue := service.GaugeMetricValue(42.0)
	counterDelta := service.CounterMetricValue(100)

	t.Run("Save Gauge Metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDB(ctrl)
		mockSaveGaugeStmt := NewMockStmt(ctrl)

		storage := &DBStorage{
			db:            mockDB,
			saveGaugeStmt: mockSaveGaugeStmt,
		}

		metric := service.Metrics{
			ID:    "gauge1",
			MType: service.GaugeMetric,
			Value: &gaugeValue,
		}

		mockDB.EXPECT().Ping().Return(nil).AnyTimes()

		mockSaveGaugeStmt.EXPECT().
			Exec(metric.ID, metric.MType, gomock.Any()).
			Return(nil, nil)

		err := storage.Save(context.Background(), metric)
		assert.NoError(t, err)
	})

	t.Run("Save Counter Metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDB(ctrl)
		mockSaveCounterStmt := NewMockStmt(ctrl)

		storage := &DBStorage{
			db:              mockDB,
			saveCounterStmt: mockSaveCounterStmt,
		}

		metric := service.Metrics{
			ID:    "counter1",
			MType: service.CounterMetric,
			Delta: &counterDelta,
		}

		mockDB.EXPECT().Ping().Return(nil).AnyTimes()

		mockSaveCounterStmt.EXPECT().
			Exec(metric.ID, metric.MType, gomock.Any()).
			Return(nil, nil)

		err := storage.Save(context.Background(), metric)
		assert.NoError(t, err)
	})

	t.Run("Ping Error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDB(ctrl)

		storage := &DBStorage{
			db: mockDB,
		}

		mockDB.EXPECT().Ping().Return(errors.New("ping failed")).AnyTimes()

		err := storage.Save(context.Background(), service.Metrics{})
		assert.Error(t, err)
		assert.Equal(t, "ping failed", err.Error())
	})

	t.Run("Invalid Metric Type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDB(ctrl)

		storage := &DBStorage{
			db: mockDB,
		}

		mockDB.EXPECT().Ping().Return(nil).AnyTimes()

		err := storage.Save(context.Background(), service.Metrics{
			ID:    "invalid_metric",
			MType: "unknown_type",
		})
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})

	t.Run("Empty Metric ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockDB := NewMockDB(ctrl)

		storage := &DBStorage{
			db: mockDB,
		}

		mockDB.EXPECT().Ping().Return(nil).AnyTimes()

		err := storage.Save(context.Background(), service.Metrics{
			ID:    "",
			MType: service.GaugeMetric,
		})
		assert.Error(t, err)
		assert.Equal(t, service.ErrInvalidMetricName, err)
	})
}
