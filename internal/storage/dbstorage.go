package storage

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/dvkhr/metrix.git/internal/service"
)

type DBStorage struct {
	DBDSN           string
	db              *sql.DB
	saveGaugeStmt   *sql.Stmt
	saveCounterStmt *sql.Stmt
	getStmt         *sql.Stmt
	listStmt        *sql.Stmt
}

func (ms *DBStorage) NewStorage() error {
	var err error
	if ms.db, err = sql.Open("pgx", ms.DBDSN); err != nil {
		return nil
	}

	var createStmt = "create table if not exists metrix (id varchar(32) PRIMARY KEY, value jsonb not null)"
	if _, err = ms.db.Exec(createStmt); err != nil {
		return nil
	}

	if ms.saveGaugeStmt, err = ms.db.Prepare("insert into metrix " +
		"values($1::varchar, " +
		"jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'value', $3::double precision)) " +
		"on conflict(id) do update " +
		"set value = jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'value', $3::double precision) " +
		"where metrix.id = $1::varchar;"); err != nil {
		return err
	}
	if ms.saveCounterStmt, err = ms.db.Prepare("insert into metrix " +
		"values($1::varchar, " +
		"jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'delta', $3::bigint)) " +
		"on conflict(id) do update " +
		"set value = jsonb_set(metrix.value, '{delta}', ((metrix.value ->> 'delta')::bigint + $3::bigint)::text::jsonb, false) " +
		"where metrix.id = $1::varchar;"); err != nil {
		return err
	}
	if ms.getStmt, err = ms.db.Prepare("select value from metrix where id = $1::varchar;"); err != nil {
		return err
	}
	if ms.listStmt, err = ms.db.Prepare("select jsonb_object_agg(k,v) from metrix, jsonb_each(jsonb_build_object(id, value)) as t(k,v);"); err != nil {
		return err
	}
	return nil
}

func (ms *DBStorage) Save(ctx context.Context, mt service.Metrics) error {
	if ms.db.Ping() != nil {
		return service.ErrUninitializedStorage
	}

	if len(mt.ID) == 0 {
		return service.ErrInvalidMetricName
	}

	if mt.MType == service.GaugeMetric {
		if _, err := ms.saveGaugeStmt.Exec(mt.ID, mt.MType, mt.Value); err != nil {
			return err
		}
	} else if mt.MType == service.CounterMetric {
		if _, err := ms.saveCounterStmt.Exec(mt.ID, mt.MType, mt.Delta); err != nil {
			return err
		}
	} else {
		return service.ErrInvalidMetricName
	}
	return nil
}

func (ms *DBStorage) SaveAll(ctx context.Context, mt *[]service.Metrics) error {
	if ms.db.Ping() != nil {
		return service.ErrUninitializedStorage
	}

	if len(*mt) == 0 {
		return service.ErrInvalidMetricName
	}

	pgTx, err := ms.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for _, metric := range *mt {
		if metric.MType == service.GaugeMetric {
			if _, err := ms.saveGaugeStmt.Exec(metric.ID, metric.MType, metric.Value); err != nil {
				pgTx.Rollback()
				return err
			}
		} else if metric.MType == service.CounterMetric {
			if _, err := ms.saveCounterStmt.Exec(metric.ID, metric.MType, metric.Delta); err != nil {
				pgTx.Rollback()
				return err
			}
		} else {
			pgTx.Rollback()
			return service.ErrInvalidMetricName
		}
	}
	pgTx.Commit()

	return nil
}

func (ms *DBStorage) Get(ctx context.Context, metricName string) (*service.Metrics, error) {
	if ms.db.Ping() != nil {
		return nil, service.ErrUninitializedStorage
	}

	if len(metricName) == 0 {
		return nil, service.ErrInvalidMetricName
	}

	var data []byte
	var mtrx service.Metrics

	if err := ms.getStmt.QueryRow(metricName).Scan(&data); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &mtrx); err != nil {
		return nil, err
	}

	return &mtrx, nil
}

func (ms *DBStorage) List(ctx context.Context) (*map[string]service.Metrics, error) {
	if ms.db.Ping() != nil {
		return nil, service.ErrUninitializedStorage
	}

	var data []byte
	var mtrx map[string]service.Metrics

	if err := ms.listStmt.QueryRow().Scan(&data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		mtrx = make(map[string]service.Metrics)
		return &mtrx, nil
	}

	if err := json.Unmarshal(data, &mtrx); err != nil {
		return nil, err
	}
	return &mtrx, nil
}

func (ms *DBStorage) FreeStorage() error {
	return ms.db.Close()
}

func (ms *DBStorage) CheckStorage() error {
	if ms.db.Ping() != nil {
		return service.ErrUninitializedStorage
	}
	return nil
}

func (ms *DBStorage) ListSlice(ctx context.Context) ([]service.Metrics, error) {
	if ms.db.Ping() != nil {
		return nil, service.ErrUninitializedStorage
	}

	var data []byte
	var mtrx []service.Metrics

	if err := ms.listStmt.QueryRow().Scan(&data); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		mtrx = make([]service.Metrics, 0, len(data))
		return mtrx, nil
	}

	if err := json.Unmarshal(data, &mtrx); err != nil {
		return nil, err
	}
	return mtrx, nil
}
