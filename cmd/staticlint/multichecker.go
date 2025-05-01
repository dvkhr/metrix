package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	// Карта всех доступных анализаторов
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
