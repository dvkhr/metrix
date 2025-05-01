// Package analyzers предоставляет статические анализаторы для проверки кода на соответствие определенным правилам.
package analyzers

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer — анализатор, запрещающий использование os.Exit в функции main пакета main.
// Анализатор проверяет все вызовы функций в пакете main и сообщает об использовании os.Exit.
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "Запрещает использование os.Exit в функции main пакета main.",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if !isMainPackage(file) {
			continue
		}

		// Переменная для отслеживания текущей функции
		var currentFuncName string

		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.FuncDecl:
				// Сохраняем имя текущей функции
				currentFuncName = n.Name.Name
			case *ast.CallExpr:
				// Проверяем, что мы находимся внутри функции main
				if currentFuncName == "main" && isOsExit(n, pass.TypesInfo) {
					pass.Reportf(n.Pos(), "прямой вызов os.Exit в функции main пакета main запрещен")
				}
			}
			return true
		})
	}

	return nil, nil
}

// isMainPackage проверяет, является ли файл частью пакета main.
func isMainPackage(file *ast.File) bool {
	return file.Name.Name == "main" && !strings.HasSuffix(file.Name.Name, "_test")
}

// isOsExit проверяет, является ли вызов функцией os.Exit.
func isOsExit(callExpr *ast.CallExpr, typesInfo *types.Info) bool {
	if typesInfo == nil {
		return false
	}

	fun, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	xIdent, ok := fun.X.(*ast.Ident)
	if !ok {
		return false
	}

	pkg, ok := typesInfo.ObjectOf(xIdent).(*types.PkgName)
	if !ok || pkg.Imported().Path() != "os" {
		return false
	}

	return fun.Sel.Name == "Exit"
}
