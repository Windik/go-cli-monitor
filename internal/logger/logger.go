package logger

import (
	"fmt"
	"os"
	"time"
)

// 1. Создаем собственный тип для уровней логирования
type LogLevel string

const (
	LevelInfo     LogLevel = "INFO"
	LevelWarning  LogLevel = "WARNING"
	LevelError    LogLevel = "ERROR"
	LevelCritical LogLevel = "CRITICAL"
)

// 2. Имя единого файла логов
const LogFileName = "monitor.log"

// 3. Универсальная функция для записи ЛЮБОГО уровня лога
func Log(level LogLevel, message string) {
	// Открываем файл. Флаги говорят: дописывать в конец (APPEND) и создать, если нет (CREATE)
	f, err := os.OpenFile(LogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Используем стандартный %v для вывода ошибки в консоль
		fmt.Printf("Can't open log file: %v\n", err)
		return
	}
	defer f.Close()

	// Формируем строку: подставляем текущее время, динамический уровень [%s] и само сообщение
	logLine := fmt.Sprintf("%s [%s] %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		level,
		message,
	)

	// Записываем строку в файл
	if _, err := f.WriteString(logLine); err != nil {
		fmt.Printf("Can't write to log file: %v\n", err)
	}
}
