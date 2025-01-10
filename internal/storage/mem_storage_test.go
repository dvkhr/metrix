package storage

import (
	"reflect"
	"sort"
	"testing"
)

func TestMemStorage_NewMemStorage(t *testing.T) {
	type fields struct {
		data map[string]interface{}
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

func TestMemStorage_PutGaugeMetric(t *testing.T) {
	type fields struct {
		data map[string]interface{}
	}
	type args struct {
		metricName  string
		metricValue GaugeMetricValue
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
			expectErr: ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName:  "",
				metricValue: 0.0,
			},
			expectErr: ErrInvalidMetricName,
		},
		{
			name: "SuccessfulNewMetric",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10.0,
			},
			expect: fields{
				data: map[string]interface{}{
					"Metric1": GaugeMetricValue(10.0),
				},
			},
			expectErr: nil,
		},
		{
			name: "SuccessfulExistingMetric",
			fields: fields{
				data: map[string]interface{}{
					"Metric1": GaugeMetricValue(15.0),
				},
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10.0,
			},
			expect: fields{
				data: map[string]interface{}{
					"Metric1": GaugeMetricValue(10.0),
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
	type fields struct {
		data map[string]interface{}
	}
	type args struct {
		metricName  string
		metricValue CounterMetricValue
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
			expectErr: ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName:  "",
				metricValue: 0,
			},
			expectErr: ErrInvalidMetricName,
		},
		{
			name: "SuccessfulNewMetric",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10,
			},
			expect: fields{
				data: map[string]interface{}{
					"Metric1": CounterMetricValue(10.0),
				},
			},
			expectErr: nil,
		},
		{
			name: "SuccessfulExistingMetric",
			fields: fields{
				data: map[string]interface{}{
					"Metric1": CounterMetricValue(15),
				},
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10,
			},
			expect: fields{
				data: map[string]interface{}{
					"Metric1": CounterMetricValue(25),
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
	type fields struct {
		data map[string]interface{}
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    GaugeMetricValue
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName: "",
			},
			expectErr: ErrInvalidMetricName,
		},
		{
			name: "UnknownMetric",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName: "UnknownMetric",
			},
			expectErr: ErrUnkonownMetric,
		},
		{
			name: "Successful",
			fields: fields{
				data: map[string]interface{}{
					"Metric1": GaugeMetricValue(15.0),
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
	type fields struct {
		data map[string]interface{}
	}
	type args struct {
		metricName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		expect    CounterMetricValue
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: ErrUninitializedStorage,
		},
		{
			name: "InvalidMetricNmae",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName: "",
			},
			expectErr: ErrInvalidMetricName,
		},
		{
			name: "UnknownMetric",
			fields: fields{
				data: make(map[string]interface{}),
			},
			args: args{
				metricName: "UnknownMetric",
			},
			expectErr: ErrUnkonownMetric,
		},
		{
			name: "Successful",
			fields: fields{
				data: map[string]interface{}{
					"Metric1": CounterMetricValue(15),
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
}

func TestMemStorage_MetricStrings(t *testing.T) {
	type fields struct {
		data map[string]interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		expect    []string
		expectErr error
	}{
		{
			name: "UninitializedStorage",
			fields: fields{
				data: nil,
			},
			expectErr: ErrUninitializedStorage,
		},
		{
			name: "SuccessfulEmpty",
			fields: fields{
				data: make(map[string]interface{}),
			},
			expectErr: nil,
			expect:    []string{},
		},
		{
			name: "Successful",
			fields: fields{
				data: map[string]interface{}{
					"Metric1": CounterMetricValue(15),
					"Metric2": GaugeMetricValue(23.0),
				},
			},
			expectErr: nil,
			expect: []string{
				"counter/Metric1/15",
				"gauge/Metric2/23",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				data: tt.fields.data,
			}
			got, err := ms.MetricStrings()
			if err != tt.expectErr {
				t.Errorf("MemStorage.MetricStrings() error [%v], expect error [%v]", err, tt.expectErr)
				return
			}

			sort.Strings(got)
			sort.Strings(tt.expect)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("MemStorage.MetricStrings() result [%+v], expect [%+v]", got, tt.expect)
			}
		})
	}
}
