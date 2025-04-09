package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/jackc/pgx/v5/pgconn"
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
		return err
	}

	err = ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return err
	}

	var createStmt = "create table if not exists metrix (id varchar(32) PRIMARY KEY, value jsonb not null)"
	err = ms.retry(func() error {
		_, err = ms.db.Exec(createStmt)
		return err
	}, 3)
	if err != nil {
		return err
	}

	err = ms.retry(func() error {
		ms.saveGaugeStmt, err = ms.db.Prepare("insert into metrix " +
			"values($1::varchar, " +
			"jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'value', $3::double precision)) " +
			"on conflict(id) do update " +
			"set value = jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'value', $3::double precision) " +
			"where metrix.id = $1::varchar;")
		return err
	}, 3)
	if err != nil {
		return err
	}

	err = ms.retry(func() error {
		ms.saveCounterStmt, err = ms.db.Prepare("insert into metrix " +
			"values($1::varchar, " +
			"jsonb_build_object('id', $1::varchar, 'type', $2::varchar, 'delta', $3::bigint)) " +
			"on conflict(id) do update " +
			"set value = jsonb_set(metrix.value, '{delta}', ((metrix.value ->> 'delta')::bigint + $3::bigint)::text::jsonb, false) " +
			"where metrix.id = $1::varchar;")
		return err
	}, 3)
	if err != nil {
		return err
	}

	err = ms.retry(func() error {
		ms.getStmt, err = ms.db.Prepare("select value from metrix where id = $1::varchar;")
		return err
	}, 3)
	if err != nil {
		return err
	}

	err = ms.retry(func() error {
		ms.listStmt, err = ms.db.Prepare("select jsonb_object_agg(k,v) from metrix, jsonb_each(jsonb_build_object(id, value)) as t(k,v);")
		return err
	}, 3)
	if err != nil {
		return err
	}

	return nil
}

func isPgTransportError(err error) bool {
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code[2:] == "08" {
			return true
		}
	}
	return false
}

func (ms *DBStorage) retry(f func() error, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = f()
		if err == nil || !isPgTransportError(err) {
			return err
		}
		logging.Logg.Error("Postgres retry after error %v\n", err)

		time.Sleep(time.Duration(2*i+1) * time.Second)
	}
	return err
}

func (ms *DBStorage) Save(ctx context.Context, mt service.Metrics) error {
	err := ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return err
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
	err := ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return err
	}

	if len(*mt) == 0 {
		return service.ErrInvalidMetricName
	}

	var pgTx *sql.Tx

	err = ms.retry(func() error {
		pgTx, err = ms.db.BeginTx(ctx, nil)
		return err
	}, 3)
	if err != nil {
		return err
	}
	for _, metric := range *mt {
		if metric.MType == service.GaugeMetric {
			err = ms.retry(func() error {
				_, err := ms.saveGaugeStmt.Exec(metric.ID, metric.MType, metric.Value)
				return err
			}, 3)
			if err != nil {
				pgTx.Rollback()
				return err
			}
		} else if metric.MType == service.CounterMetric {
			err = ms.retry(func() error {
				_, err := ms.saveCounterStmt.Exec(metric.ID, metric.MType, metric.Delta)
				return err
			}, 3)
			if err != nil {
				pgTx.Rollback()
				return err
			}
		} else {
			pgTx.Rollback()
			return service.ErrInvalidMetricName
		}
	}

	err = ms.retry(func() error {
		pgTx.Commit()
		return err
	}, 3)
	if err != nil {
		return err
	}

	return nil
}

func (ms *DBStorage) Get(ctx context.Context, metricName string) (*service.Metrics, error) {
	err := ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return nil, err
	}

	if len(metricName) == 0 {
		return nil, service.ErrInvalidMetricName
	}

	var data []byte
	var mtrx service.Metrics
	err = ms.retry(func() error {
		err := ms.getStmt.QueryRow(metricName).Scan(&data)
		return err
	}, 3)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &mtrx); err != nil {
		return nil, err
	}

	return &mtrx, nil
}

func (ms *DBStorage) List(ctx context.Context) (*map[string]service.Metrics, error) {
	err := ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return nil, err
	}

	var data []byte
	var mtrx map[string]service.Metrics
	err = ms.retry(func() error {
		err := ms.listStmt.QueryRow().Scan(&data)
		return err
	}, 3)
	if err != nil {
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
	err := ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return err
	}
	return nil
}

func (ms *DBStorage) ListSlice(ctx context.Context) ([]service.Metrics, error) {
	err := ms.retry(func() error {
		err := ms.db.Ping()
		return err
	}, 3)
	if err != nil {
		return nil, err
	}

	var data []byte
	var mtrx []service.Metrics

	err = ms.retry(func() error {
		err := ms.listStmt.QueryRow().Scan(&data)
		return err
	}, 3)
	if err != nil {
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
