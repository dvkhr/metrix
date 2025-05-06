// Package service предоставляет функции для работы с метриками.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"

	//----
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// MetricType представляет тип метрики.
// Возможные значения: "gauge" или "counter".
type MetricType string

const (
	// GaugeMetric указывает на метрику типа "gauge".
	GaugeMetric MetricType = "gauge"

	// CounterMetric указывает на метрику типа "counter".
	CounterMetric MetricType = "counter"
)

// GaugeMetricValue представляет значение метрики типа "gauge".
type GaugeMetricValue float64

// CounterMetricValue представляет значение метрики типа "counter".
type CounterMetricValue int64

// Metrics описывает структуру метрики.
// Используется для передачи данных о метриках в формате JSON.
type Metrics struct {
	// ID — уникальное имя метрики.
	ID string `json:"id"`

	// MType — тип метрики (gauge или counter).
	MType MetricType `json:"type"`

	// Delta — значение метрики в случае, если тип метрики — counter.
	// Может быть nil, если метрика имеет тип gauge.
	Delta *CounterMetricValue `json:"delta,omitempty"`

	// Value — значение метрики в случае, если тип метрики — gauge.
	// Может быть nil, если метрика имеет тип counter.
	Value *GaugeMetricValue `json:"value,omitempty"`
}

var (
	// ErrUninitializedStorage возвращается, если хранилище метрик не было инициализировано.
	ErrUninitializedStorage = errors.New("storage is not initialized")

	// ErrInvalidMetricName возвращается, если имя метрики некорректно.
	ErrInvalidMetricName = errors.New("invalid metric name")

	// ErrUnknownMetric возвращается, если запрашиваемая метрика не найдена в хранилище.
	ErrUnknownMetric = errors.New("unknown metric")
)

type MetricStorage interface {
	// Save сохраняет метрику в хранилище.
	// Возвращает ошибку, если сохранение завершилось неудачей.
	Save(ctx context.Context, mt Metrics) error

	// List возвращает все метрики из хранилища в виде карты,
	// где ключ — имя метрики, а значение — объект Metrics.
	List(ctx context.Context) (*map[string]Metrics, error)
}

// DumpMetrics извлекает все метрики из хранилища и записывает их в указанный writer в формате JSON.
//
// Функция используется для экспорта текущего состояния метрик в удобочитаемом формате.
//
// Параметры:
// - ms: Интерфейс MetricStorage, предоставляющий доступ к хранилищу метрик.
// - wr: Writer, куда будут записаны сериализованные метрики (например, файл или HTTP-ответ).
//
// Возвращаемое значение:
//   - error: Ошибка, если произошла проблема при получении, сериализации или записи метрик.
//     Если операция выполнена успешно, возвращается nil.
func DumpMetrics(ms MetricStorage, wr io.Writer) error {

	ctx := context.TODO()
	mtrx, err := ms.List(ctx)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(mtrx, "", "  ")
	if err != nil {
		return err
	}
	_, err = wr.Write(data)
	return err
}

// RestoreMetrics восстанавливает метрики из данных, прочитанных из указанного reader.
//
// Функция используется для импорта состояния метрик из JSON-данных в хранилище.
//
// Параметры:
// - ms: Интерфейс MetricStorage, предоставляющий доступ к хранилищу метрик.
// - rd: Reader, откуда будут считаны сериализованные метрики (например, файл или HTTP-запрос).
//
// Возвращаемое значение:
//   - error: Ошибка, если произошла проблема при чтении, получении состояния хранилища или десериализации данных.
//     Если операция выполнена успешно, возвращается nil.
func RestoreMetrics(ms MetricStorage, rd io.Reader) error {
	ctx := context.TODO()
	var data []byte

	data, err := io.ReadAll(rd)
	if err != nil {
		return err
	}

	stor, err := ms.List(ctx)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, stor)

	return err
}

// CollectMetricsOS собирает метрики операционной системы и отправляет их в канал.
//
// Логика работы:
//  1. Определяется вспомогательная функция collectMetric, которая создает объект Metrics
//     на основе типа метрики (GaugeMetric или CounterMetric) и отправляет его в канал.
//  2. Собираются метрики памяти (TotalMemory и FreeMemory) с помощью библиотеки mem.VirtualMemory.
//     Если возникает ошибка при сборе метрик памяти, она логируется.
//  3. Собираются метрики загрузки CPU (CPUutilization) с помощью библиотеки cpu.Percent.
//     Для каждого ядра CPU создается отдельная метрика.
//     Если возникает ошибка при сборе метрик CPU, она выводится в stderr.
//  4. Все собранные метрики отправляются в канал metrics для дальнейшей обработки.
//
// Параметры:
// - ctx: Контекст для управления жизненным циклом функции.
// - metrics: Канал, куда отправляются собранные метрики.
//
// Примечание:
// - Метрики памяти (TotalMemory и FreeMemory) имеют тип "gauge".
// - Метрики загрузки CPU (CPUutilization) также имеют тип "gauge".
// - В случае ошибок при сборе метрик они логируются, но выполнение функции продолжается.
func CollectMetricsOS(ctx context.Context, metrics chan Metrics) {
	logging.Logg.Info("+++Run CollectMetricsOS+++\n")

	collectMetric := func(metricType MetricType, metricName string, metricValue any) {
		var mt Metrics
		switch metricType {
		case GaugeMetric:
			temp := metricValue.(GaugeMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Value: &temp}
		case CounterMetric:
			temp := metricValue.(CounterMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Delta: &temp}
		}
		metrics <- mt
	}

	vMem, err := mem.VirtualMemory()
	if err != nil {
		logging.Logg.Error("collecting memory metrics: %v\n", err)
	}

	collectMetric(GaugeMetric, "TotalMemory", GaugeMetricValue(vMem.Total))
	collectMetric(GaugeMetric, "FreeMemory", GaugeMetricValue(vMem.Free))

	CPUutilization, err := cpu.Percent(0, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collecting CPU metrics: %v\n", err)
	}
	for i, iCPUtil := range CPUutilization {
		collectMetric(GaugeMetric, fmt.Sprintf("CPUutilization%d", i+1), GaugeMetricValue(iCPUtil))
	}
}

// CollectMetricsCh собирает метрики о состоянии среды выполнения Go (runtime metrics)
// и отправляет их в указанный канал.
//
// Логика работы:
//  1. Определяется вспомогательная функция collectMetric, которая создает объект Metrics
//     на основе типа метрики (GaugeMetric или CounterMetric) и отправляет его в канал.
//  2. Собираются метрики runtime с помощью функции runtime.ReadMemStats.
//     Эти метрики включают информацию об использовании памяти, работе сборщика мусора и других параметрах.
//  3. Для каждой метрики runtime создается объект Metrics типа "gauge" и отправляется в канал.
//  4. Дополнительно генерируется случайное значение (RandomValue) и счетчик PollCount,
//     который увеличивается на единицу при каждом вызове функции.
//
// Параметры:
// - ctx: Контекст для управления жизненным циклом функции.
// - metrics: Канал, куда отправляются собранные метрики.
//
// Примечание:
// - Все метрики runtime имеют тип "gauge".
// - Метрика PollCount имеет тип "counter" и используется для подсчета количества вызовов функции.
func CollectMetricsCh(ctx context.Context, metrics chan Metrics) {
	fmt.Println("Run CollectMetricsCh")

	collectMetric := func(metricType MetricType, metricName string, metricValue any) {
		var mt Metrics
		switch metricType {
		case GaugeMetric:
			temp := metricValue.(GaugeMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Value: &temp}
		case CounterMetric:
			temp := metricValue.(CounterMetricValue)
			mt = Metrics{ID: metricName, MType: metricType, Delta: &temp}
		}
		metrics <- mt
	}

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	gaugeMetrics := []struct {
		Name  string
		Value GaugeMetricValue
	}{
		{"Alloc", GaugeMetricValue(rtm.Alloc)},
		{"BuckHashSys", GaugeMetricValue(rtm.BuckHashSys)},
		{"Frees", GaugeMetricValue(rtm.Frees)},
		{"GCCPUFraction", GaugeMetricValue(rtm.GCCPUFraction)},
		{"GCSys", GaugeMetricValue(rtm.GCSys)},
		{"HeapAlloc", GaugeMetricValue(rtm.HeapAlloc)},
		{"HeapIdle", GaugeMetricValue(rtm.HeapIdle)},
		{"HeapInuse", GaugeMetricValue(rtm.HeapInuse)},
		{"HeapObjects", GaugeMetricValue(rtm.HeapObjects)},
		{"HeapReleased", GaugeMetricValue(rtm.HeapReleased)},
		{"HeapSys", GaugeMetricValue(rtm.HeapSys)},
		{"LastGC", GaugeMetricValue(rtm.LastGC)},
		{"Lookups", GaugeMetricValue(rtm.Lookups)},
		{"MCacheInuse", GaugeMetricValue(rtm.MCacheInuse)},
		{"MCacheSys", GaugeMetricValue(rtm.MCacheSys)},
		{"MSpanInuse", GaugeMetricValue(rtm.MSpanInuse)},
		{"MSpanSys", GaugeMetricValue(rtm.MSpanSys)},
		{"Mallocs", GaugeMetricValue(rtm.Mallocs)},
		{"NextGC", GaugeMetricValue(rtm.NextGC)},
		{"NumForcedGC", GaugeMetricValue(rtm.NumForcedGC)},
		{"NumGC", GaugeMetricValue(rtm.NumGC)},
		{"OtherSys", GaugeMetricValue(rtm.OtherSys)},
		{"PauseTotalNs", GaugeMetricValue(rtm.PauseTotalNs)},
		{"StackInuse", GaugeMetricValue(rtm.StackInuse)},
		{"StackSys", GaugeMetricValue(rtm.StackSys)},
		{"Sys", GaugeMetricValue(rtm.Sys)},
		{"TotalAlloc", GaugeMetricValue(rtm.TotalAlloc)},
		{"RandomValue", GaugeMetricValue(rand.Float64())},
	}
	for _, gm := range gaugeMetrics {
		collectMetric(GaugeMetric, gm.Name, gm.Value)
	}

	collectMetric(CounterMetric, "PollCount", CounterMetricValue(1))
}
