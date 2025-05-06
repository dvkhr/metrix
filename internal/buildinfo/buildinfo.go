package buildinfo

import "fmt"

// printBuildInfo выводит информацию о сборке приложения в стандартный поток вывода (stdout).
// Информация включает версию сборки (buildVersion), дату сборки (buildDate) и хеш коммита (buildCommit).
// Если какое-либо из значений отсутствует (пустая строка), вместо него выводится "N/A" (Not Available).
func PrintBuildInfo(buildVersion, buildDate, buildCommit string) {
	fmt.Printf("Build version: %s\n", getValueOrDefault(buildVersion))
	fmt.Printf("Build date: %s\n", getValueOrDefault(buildDate))
	fmt.Printf("Build commit: %s\n", getValueOrDefault(buildCommit))
}

// getValueOrDefault проверяет, является ли переданное значение пустой строкой.
// Если значение пустое, функция возвращает "N/A" (Not Available).
// В противном случае возвращается само значение.
func getValueOrDefault(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
