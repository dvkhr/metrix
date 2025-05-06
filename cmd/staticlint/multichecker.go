// Package main предоставляет multichecker для статического анализа Go-кода.
//
// Программа читает конфигурационный файл, содержащий список анализаторов,
// и запускает их с помощью multichecker.Main. Поддерживаются как стандартные
// анализаторы из пакета golang.org/x/tools/go/analysis/passes,
// так и сторонние анализаторы из honnef.co/go/tools (staticcheck, simple, stylecheck).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dvkhr/metrix.git/internal/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Config — имя файла конфигурации.
const Config = "config.json"

// ConfigData описывает структуру файла конфигурации.
type ConfigData struct {
	Analyzers []string `json:"analyzers"`
}

// main реализует запуск multichecker для статического анализа Go-кода.
//
// Программа выполняет следующие шаги:
// 1. Определяет путь к исполняемому файлу.
// 2. Читает конфигурационный файл config.json, содержащий список анализаторов.
// 3. Создает карту всех доступных анализаторов.
// 4. Формирует список анализаторов на основе конфигурации.
// 5. Запускает multichecker.Main с выбранными анализаторами.
func main() {
	// Определяем путь к исполняемому файлу
	appfile, err := os.Executable()
	if err != nil {
		panic(fmt.Errorf("failed to get executable path: %w", err))
	}

	// Читаем конфигурационный файл
	configPath := filepath.Join(filepath.Dir(appfile), Config)
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("failed to read config file: %w", err))
	}

	// Парсим конфигурацию
	var cfg ConfigData
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Errorf("failed to parse config file: %w", err))
	}

	allAnalyzers := initAnalyzersMap()

	// Формируем список анализаторов на основе конфигурации
	var mychecks []*analysis.Analyzer
	checks := make(map[string]bool)
	for _, v := range cfg.Analyzers {
		checks[v] = true
	}
	for name, analyzer := range allAnalyzers {
		if checks[name] {
			mychecks = append(mychecks, analyzer)
		}
	}

	// Запускаем multichecker
	multichecker.Main(mychecks...)
}

// initAnalyzersMap создает и возвращает карту всех доступных анализаторов.
func initAnalyzersMap() map[string]*analysis.Analyzer {
	allAnalyzers := map[string]*analysis.Analyzer{
		"asmdecl":         asmdecl.Analyzer,
		"assign":          assign.Analyzer,
		"atomic":          atomic.Analyzer,
		"bools":           bools.Analyzer,
		"buildtag":        buildtag.Analyzer,
		"cgocall":         cgocall.Analyzer,
		"composite":       composite.Analyzer,
		"copylock":        copylock.Analyzer,
		"deepequalerrors": deepequalerrors.Analyzer,
		"errorsas":        errorsas.Analyzer,
		"httpresponse":    httpresponse.Analyzer,
		"ifaceassert":     ifaceassert.Analyzer,
		"loopclosure":     loopclosure.Analyzer,
		"lostcancel":      lostcancel.Analyzer,
		"nilfunc":         nilfunc.Analyzer,
		"printf":          printf.Analyzer,
		"shift":           shift.Analyzer,
		"stdmethods":      stdmethods.Analyzer,
		"stringintconv":   stringintconv.Analyzer,
		"structtag":       structtag.Analyzer,
		"tests":           tests.Analyzer,
		"unmarshal":       unmarshal.Analyzer,
		"unreachable":     unreachable.Analyzer,
		"unusedresult":    unusedresult.Analyzer,
		"nilness":         nilness.Analyzer,
		"noosexit":        analyzers.NoOsExitAnalyzer,
	}

	// Добавляем анализаторы staticcheck
	for _, a := range staticcheck.Analyzers {
		allAnalyzers[a.Analyzer.Name] = a.Analyzer
	}
	for _, a := range simple.Analyzers {
		allAnalyzers[a.Analyzer.Name] = a.Analyzer
	}
	for _, a := range stylecheck.Analyzers {
		allAnalyzers[a.Analyzer.Name] = a.Analyzer
	}

	return allAnalyzers
}
