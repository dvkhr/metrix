package handlers

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/dvkhr/metrix.git/internal/config"
	"github.com/dvkhr/metrix.git/internal/crypto"
	"github.com/dvkhr/metrix.git/internal/logging"
	"github.com/dvkhr/metrix.git/internal/service"
	"github.com/dvkhr/metrix.git/internal/storage"
)

// MetricStorage представляет интерфейс для работы с хранилищем метрик.
//
// Он определяет набор методов для сохранения, получения и управления метриками.
// Реализации этого интерфейса могут использовать различные типы хранилищ,
// такие как база данных, файловое хранилище или оперативная память.
//
// Методы:
// - Save: Сохраняет одну метрику в хранилище.
// - SaveAll: Сохраняет массив метрик в хранилище.
// - Get: Получает метрику по её имени.
// - List: Возвращает все метрики в виде мапы, где ключ — имя метрики.
// - ListSlice: Возвращает все метрики в виде слайса.
// - NewStorage: Инициализирует хранилище.
// - FreeStorage: Освобождает ресурсы, связанные с хранилищем.
// - CheckStorage: Проверяет доступность хранилища.
type MetricStorage interface {
	Save(ctx context.Context, mt service.Metrics) error
	SaveAll(ctx context.Context, mt *[]service.Metrics) error
	Get(ctx context.Context, metricName string) (*service.Metrics, error)
	List(ctx context.Context) (*map[string]service.Metrics, error)
	ListSlice(ctx context.Context) ([]service.Metrics, error)
	NewStorage() error
	FreeStorage() error
	CheckStorage() error
}

//mockgen -source=internal/handlers/handlers.go -destination=internal/mocks/mock_storage.go -package=mocks

// MetricsServer представляет сервер для обработки метрик.
// Он управляет хранилищем метрик и предоставляет методы для их сохранения, получения и обработки.
// Поля:
//   - MetricStorage: Интерфейс хранилища метрик (база данных, файловое хранилище или память).
//     Используется для выполнения операций с метриками.
//   - Config: Конфигурация сервера, содержащая параметры подключения и настройки.
//   - syncMutex: Мьютекс для обеспечения потокобезопасности при работе с общими ресурсами.
type MetricsServer struct {
	MetricStorage MetricStorage
	Config        config.ConfigServ
	syncMutex     sync.Mutex
}

// NewMetricsServer создает новый экземпляр MetricsServer с выбранным хранилищем метрик.
//
// Выбор хранилища зависит от конфигурации:
// - Если Config.DBDsn не пустой, используется хранилище на основе PostgreSQL (DBStorage).
// - Если Config.FileStoragePath не пустой, используется файловое хранилище (FileStorage).
// - Если ни один из вышеперечисленных параметров не задан, используется хранилище в оперативной памяти (MemStorage).
//
// Параметры:
// - Config: Конфигурация сервера, содержащая параметры для подключения к хранилищу.

// Возвращаемые значения:
// - *MetricsServer: Указатель на созданный экземпляр MetricsServer.
// - error: Ошибка, если произошла проблема при инициализации хранилища.
func NewMetricsServer(Config config.ConfigServ) (*MetricsServer, error) {
	var ms MetricStorage
	if len(Config.DBDsn) > 0 {
		ms = &storage.DBStorage{DBDSN: Config.DBDsn}

	} else if len(Config.FileStoragePath) > 0 {
		ms = &storage.FileStorage{FileStoragePath: Config.FileStoragePath}

	} else {
		ms = &storage.MemStorage{}
	}

	if err := ms.NewStorage(); err != nil {
		return nil, err
	}

	return &MetricsServer{MetricStorage: ms, Config: Config}, nil
}

// IncorrectMetricRq обрабатывает некорректные запросы на обновление метрик.
//
// В ответ на запрос отправляется HTTP-ошибка с кодом 400 (Bad Request) и сообщением:
// "Incorrect update metric request!".
func (ms *MetricsServer) IncorrectMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Incorrect update metric request!", http.StatusBadRequest)
}

// NotfoundMetricRq обрабатывает запросы на получение или обновление несуществующих метрик.
//
// В ответ на запрос отправляется HTTP-ошибка с кодом 404 (Not Found) и сообщением:
// "Metric not found!".
func (ms *MetricsServer) NotfoundMetricRq(res http.ResponseWriter, req *http.Request) {
	http.Error(res, "Metric not found!", http.StatusNotFound)
}

// checkPostMethod проверяет, является ли HTTP-метод запроса POST.
// Если метод отличается от POST, возвращается ошибка.
func (ms *MetricsServer) checkPostMethod(req *http.Request) error {
	if req.Method != http.MethodPost {
		return fmt.Errorf("only POST requests are allowed")
	}
	return nil
}

// readRequestBody читает тело HTTP-запроса и возвращает его в виде массива байтов.
// Если возникает ошибка при чтении тела запроса, она возвращается.
// После чтения тело запроса закрывается.
func (ms *MetricsServer) readRequestBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body")
	}
	defer req.Body.Close()
	return body, nil
}

// loadPrivateKey загружает приватный ключ из файла, указанного в конфигурации.
// Если путь к ключу не указан, возвращается nil.
// В случае ошибки чтения ключа возвращается соответствующая ошибка.
func (ms *MetricsServer) loadPrivateKey() (*rsa.PrivateKey, error) {
	if ms.Config.CryptoKey == "" {
		return nil, nil
	}
	privateKey, err := crypto.ReadPrivateKey(ms.Config.CryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}
	logging.Logg.Info("Private key successfully loaded")
	return privateKey, nil
}

// decryptData расшифровывает данные с использованием предоставленного приватного ключа.
// Если ключ отсутствует, возвращаются исходные данные без изменений.
// В случае ошибки расшифровки возвращается соответствующая ошибка.
func (ms *MetricsServer) decryptData(data []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return data, nil
	}
	decryptedData, err := crypto.DecryptData(string(data), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return decryptedData, nil
}

// parseMetrics десериализует JSON-данные в массив метрик.
// Если данные не могут быть преобразованы в формат []service.Metrics,
// возвращается соответствующая ошибка.
func (ms *MetricsServer) parseMetrics(data []byte) ([]service.Metrics, error) {
	var metrics []service.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics: %w", err)
	}
	return metrics, nil
}

// saveMetrics сохраняет массив метрик в хранилище.
// Если сохранение завершается ошибкой, она возвращается.
func (ms *MetricsServer) saveMetrics(ctx context.Context, metrics []service.Metrics) error {
	if err := ms.MetricStorage.SaveAll(ctx, &metrics); err != nil {
		return fmt.Errorf("failed to save metrics: %w", err)
	}
	return nil
}

// getAllMetrics получает все метрики из хранилища в виде карты (map[string]service.Metrics).
// Если возникает ошибка при получении данных, она возвращается.
func (ms *MetricsServer) getAllMetrics(ctx context.Context) (*map[string]service.Metrics, error) {
	allMetrics, err := ms.MetricStorage.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve metrics: %w", err)
	}
	return allMetrics, nil
}

// prepareResponse подготавливает ответ клиенту:
// 1. Преобразует метрики в формат JSON.
// 2. Генерирует хэш SHA-256 на основе JSON-данных и ключа (если ключ предоставлен).
//
// Возвращает:
// - Сериализованные данные метрик.
// - Хэш (если ключ предоставлен).
// - Ошибку, если что-то пошло не так.
func (ms *MetricsServer) prepareResponse(metrics *map[string]service.Metrics, key string) ([]byte, string, error) {
	// Преобразование данных в JSON
	bufResp, err := json.Marshal(metrics)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal response: %w", err)
	}

	// Если ключ предоставлен, генерируем хэш
	var hash string
	if len(key) > 0 {
		signBuf := append(bufResp, ',')
		signBuf = append(signBuf, key...)
		sign := sha256.Sum256(signBuf)
		hash = hex.EncodeToString(sign[:])
	}

	return bufResp, hash, nil
}
