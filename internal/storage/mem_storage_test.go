package storage

import (
	"testing"

	"github.com/dvkhr/metrix.git/internal/metric"
)

func TestMemStorage_NewMemStorage(t *testing.T) {
	type fields struct {
		data map[string]metric.Metrics
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Successful",
			fields: fields{
				data: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				data: tt.fields.data,
			}

			ms.NewMemStorage()

			if ms.data == nil {
				t.Errorf("MemStorage.NewMemStroage() must initialize MemStroage.data map")
			}
		})

	}
}

/*
func TestMemStorage_PutGaugeMetric(t *testing.T) {
	gauge10 := metric.GaugeMetricValue(10.0)
	gauge15 := metric.GaugeMetricValue(15.0)
	gauge25 := metric.GaugeMetricValue(25.0)
	type fields struct {
		data map[string]metric.Metrics
	}
	type args struct {
		metricName  string
		metricValue metric.GaugeMetricValue
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    fields
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: metric.ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName:  "",
				metricValue: 0.0,
			},
			expectErr: metric.ErrInvalidMetricName,
		},
		{
			name: "SuccessfulNewMetric",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10.0,
			},
			expect: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge10},
				},
			},
			expectErr: nil,
		},
		{
			name: "SuccessfulExistingMetric",
			fields: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge15},
				},
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10.0,
			},
			expect: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge25},
				},
			},
			expectErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				data: tt.fields.data,
			}

			err := ms.PutGaugeMetric(tt.args.metricName, tt.args.metricValue)
			// Check error code
			if err != tt.expectErr {
				t.Errorf("MemStorage.PutGaugeMetric() error [%v], expect error [%v]", err, tt.expectErr)
				return
			}
			// Check storage data
			if err == nil && len(tt.expect.data) > 0 && !reflect.DeepEqual(ms.data, tt.expect.data) {
				t.Errorf("MemStorage.PutGaugeMetric() result [%+v], expect [%+v]", ms.data, tt.expect.data)
			}
		})
	}
}

func TestMemStorage_PutCouinterMetric(t *testing.T) {
	gauge10 := metric.GaugeMetricValue(10.0)
	gauge15 := metric.GaugeMetricValue(15.0)
	gauge25 := metric.GaugeMetricValue(25.0)
	type fields struct {
		data map[string]metric.Metrics
	}
	type args struct {
		metricName  string
		metricValue metric.CounterMetricValue
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    fields
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: metric.ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName:  "",
				metricValue: 0,
			},
			expectErr: metric.ErrInvalidMetricName,
		},
		{
			name: "SuccessfulNewMetric",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10,
			},
			expect: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge10},
				},
			},
			expectErr: nil,
		},
		{
			name: "SuccessfulExistingMetric",
			fields: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge15},
				},
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10,
			},
			expect: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge25},
				},
			},
			expectErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := MemStorage{
				data: tt.fields.data,
			}

			err := ms.PutCounterMetric(tt.args.metricName, tt.args.metricValue)
			// Check error code
			if err != tt.expectErr {
				t.Errorf("MemStorage.PutCounterMetric() error [%v], expect error [%v]", err, tt.expectErr)
				return
			}
			// Check storage data
			if err == nil && len(tt.expect.data) > 0 && !reflect.DeepEqual(ms.data, tt.expect.data) {
				t.Errorf("MemStorage.PutCounterMetric() result [%+v], expect [%+v]", ms.data, tt.expect.data)
			}
		})
	}
}

func TestMemStorage_GetGaugeMetric(t *testing.T) {
	gauge15 := metric.GaugeMetricValue(15.0)
	//gauge23 := metric.GaugeMetricValue(23.0)
	type fields struct {
		data map[string]metric.Metrics
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    metric.GaugeMetricValue
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: metric.ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName: "",
			},
			expectErr: metric.ErrInvalidMetricName,
		},
		{
			name: "UnknownMetric",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName: "UnknownMetric",
			},
			expectErr: metric.ErrUnkonownMetric,
		},
		{
			name: "Successful",
			fields: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge15},
				},
			},
			args: args{
				metricName: "Metric1",
			},
			expectErr: nil,
			expect:    15.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				data: tt.fields.data,
			}
			got, err := ms.GetGaugeMetric(tt.args.metricName)
			if err != tt.expectErr {
				t.Errorf("MemStorage.GetGaugeMetric() error [%v], expect error [%v]", err, tt.expectErr)
				return
			}
			if got != tt.expect {
				t.Errorf("MemStorage.GetGaugeMetric() result [%v], expect [%v]", got, tt.expect)
			}
		})
	}
}

func TestMemStorage_GetCounterMetric(t *testing.T) {
	gauge15 := metric.GaugeMetricValue(15.0)
	type fields struct {
		data map[string]metric.Metrics
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    metric.CounterMetricValue
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: metric.ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName: "",
			},
			expectErr: metric.ErrInvalidMetricName,
		},
		{
			name: "UnknownMetric",
			fields: fields{
				data: make(map[string]metric.Metrics),
			},
			args: args{
				metricName: "UnknownMetric",
			},
			expectErr: metric.ErrUnkonownMetric,
		},
		{
			name: "Successful",
			fields: fields{
				data: map[string]metric.Metrics{
					"Metric1": metric.Metrics{ID: "Metric1", MType: metric.GaugeMetric, Value: &gauge15},
				},
			},
			args: args{
				metricName: "Metric1",
			},
			expectErr: nil,
			expect:    15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				data: tt.fields.data,
			}
			got, err := ms.GetCounterMetric(tt.args.metricName)
			if err != tt.expectErr {
				t.Errorf("MemStorage.GetCounterMetric() error [%v], expect error [%v]", err, tt.expectErr)
				return
			}
			if got != tt.expect {
				t.Errorf("MemStorage.GetCounterMetric() result [%v], expect [%v]", got, tt.expect)
			}
		})
	}
}*/
