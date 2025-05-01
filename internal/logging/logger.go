// Package logging предоставляет инструменты для настройки и управления логированием в приложении.
//
// Пакет использует библиотеку slog для структурированного логирования
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/dvkhr/metrix.git/internal/config"
)

// Logger — это обертка над slog.Logger
type Logger struct {
	logger *slog.Logger
}

var (
	Logg       *Logger
	loggerOnce sync.Once // Для однократной инициализации
)

// InitLogger инициализирует логгер из конфигурации.
func InitLogger(configPath string) error {
	var err error
	loggerOnce.Do(func() {
		// Загружаем конфигурацию
		cfg, loadErr := config.LoadLoggerConfig(configPath)
		if loadErr != nil {
			err = fmt.Errorf("failed to load logger config: %w", loadErr)
			return
		}

		// Получаем параметры логгера из конфигурации
		logLevel := cfg.LogLevel
		consoleFormat := cfg.ConsoleFormat
		fileFormat := cfg.FileFormat
		destination := cfg.Destination
		filePattern := cfg.FilePattern

		// Определяем уровень логирования
		level := getLogLevel(logLevel)
		if level == slog.Level(-999) {
			Logg = &Logger{
				logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
			}
			return
		}

		// Создаем обработчики для терминала и файла
		var handlers []slog.Handler

		// Обработчик для терминала
		if destination == "console" || destination == "both" {
			var consoleHandler slog.Handler
			switch consoleFormat {
			case "text":
				consoleHandler = newTextHandler(os.Stdout, level)
			case "json":
				consoleHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level: level,
				})
			default:
				consoleHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level: level,
				})
			}
			handlers = append(handlers, consoleHandler)
		}

		// Обработчик для файла
		if destination == "file" || destination == "both" {
			// Генерация имени файла
			fileName := generateFileName(filePattern)
			file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Printf("Failed to open log file: %v\n", err)
			} else {
				var fileHandler slog.Handler
				switch fileFormat {
				case "text":
					fileHandler = newTextHandler(file, level)
				case "json":
					fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{
						Level: level,
					})
				default:
					fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{
						Level: level,
					})
				}
				handlers = append(handlers, fileHandler)
			}
		}

		// Если нет ни одного обработчика, используем вывод в терминал по умолчанию
		if len(handlers) == 0 {
			handlers = append(handlers, slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: level,
			}))
		}

		// Создаем логгер с несколькими обработчиками
		Logg = &Logger{
			logger: slog.New(NewMultiHandler(handlers...)),
		}
	})
	return err
}

// InitTestLogger инициализирует тестовый логгер с уровнем логгирования "none".
func InitTestLogger() error {
	var err error
	loggerOnce.Do(func() {
		// Уровень логгирования "none" — отключаем все логи
		level := slog.Level(-999) // Значение ниже минимального уровня логгирования

		// Создаем обработчик, который игнорирует все логи
		handler := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
			Level: level,
		})

		// Инициализируем глобальный логгер
		Logg = &Logger{
			logger: slog.New(handler),
		}
	})
	return err
}

// getLogLevel преобразует строку в slog.Level
func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "none":
		return slog.Level(-999)

	default:
		// По умолчанию используем LevelInfo
		return slog.LevelInfo
	}
}

// generateFileName генерирует имя файла на основе шаблона и текущей даты
func generateFileName(pattern string) string {
	now := time.Now()

	// Заменяем шаблон на отформатированную дату
	fileName := now.Format(pattern)

	// Убедитесь, что папка существует
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
	}

	return fileName
}

// Info логирует информационное сообщение
func (l *Logger) Info(msg string, attrs ...any) {
	if l == nil || l.logger == nil {
		panic("logger is not initialized")
	}
	l.logger.Info(msg, attrs...)
}

// Warn логирует предупреждение
func (l *Logger) Warn(msg string, attrs ...any) {
	l.logger.Warn(msg, attrs...)
}

// Error логирует сообщение об ошибке
func (l *Logger) Error(msg string, attrs ...any) {
	l.logger.Error(msg, attrs...)
}

// Debug логирует отладочное сообщение
func (l *Logger) Debug(msg string, attrs ...any) {
	l.logger.Debug(msg, attrs...)
}

// Маскировка чувствительных данных
func MaskSensitiveData(body string) string {
	re := regexp.MustCompile(`("(password|token)"\s*:\s*")([^"]*)`)
	return re.ReplaceAllString(body, `$1***`)
}
