// Package handlers предоставляет HTTP-обработчики для работы с метриками.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dvkhr/metrix.git/internal/storage"
)

// ReadAndUnmarshal читает тело HTTP-запроса и десериализует его в указанную структуру.
// Функция ограничивает размер тела запроса до 1 МБ для предотвращения переполнения памяти.
//
// Параметры:
// - req: HTTP-запрос, тело которого нужно прочитать и десериализовать.
// - v: Указатель на структуру, в которую будет десериализовано тело запроса.
//
// Возвращаемое значение:
//   - error: Ошибка, если произошла проблема при чтении или декодировании тела запроса.
//     Если операция выполнена успешно, возвращается nil.
func ReadAndUnmarshal(req *http.Request, v interface{}) error {

	const maxBodySize = 1 << 20 // 1 MB
	req.Body = http.MaxBytesReader(nil, req.Body, maxBodySize)

	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		return err
	}

	defer req.Body.Close()

	return nil
}

// CheckImplementations проверяет, что все реализации интерфейса MetricStorage
// соответствуют требованиям интерфейса.
// Функция используется для статической проверки корректности реализации интерфейса.
//
// Примечание:
// - Функция не выполняет никаких действий во время выполнения программы.
// - Она предназначена исключительно для проверки соответствия интерфейсу на этапе компиляции.
func CheckImplementations() {
	var (
		_ MetricStorage = (*storage.DBStorage)(nil)
		_ MetricStorage = (*storage.FileStorage)(nil)
		_ MetricStorage = (*storage.MemStorage)(nil)
	)
}
