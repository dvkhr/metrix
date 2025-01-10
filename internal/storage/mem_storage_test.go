package storage

import (
	"reflect"
	"testing"
)

/*
func TestMemStorage_NewMemStroage(t *testing.T) {
	ms := MemStorage{}
	ms.NewMemStroage()
	if ms.data == nil {
		t.Errorf("NewMemStroage must initialize MemStroage.data map")
	}
}
*/

func TestMemStorage_NewMemStroage(t *testing.T) {
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
		metricValue GuageMetricValue
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
					"Metric1": GuageMetricValue(10.0),
				},
			},
			expectErr: nil,
		},
		{
			name: "SuccessfulExistingMetric",
			fields: fields{
				data: map[string]interface{}{
					"Metric1": GuageMetricValue(15.0),
				},
			},
			args: args{
				metricName:  "Metric1",
				metricValue: 10.0,
			},
			expect: fields{
				data: map[string]interface{}{
					"Metric1": GuageMetricValue(10.0),
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
