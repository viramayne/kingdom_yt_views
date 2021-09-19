package main

import (
	"fmt"
	"time"
)

// Вызов переданной функции раз в сутки в указанное время.
func callAt(hour, min, sec int, f func()) error {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}

	// Вычисляем время первого запуска.
	now := time.Now().Local()
	firstCallTime := time.Date(
		now.Year(), now.Month(), now.Day(), hour, min, sec, 0, loc)
	if firstCallTime.Before(now) {
		// Если получилось время раньше текущего, прибавляем сутки.
		firstCallTime = firstCallTime.Add(time.Hour * 24)
	}

	// Вычисляем временной промежуток до запуска.
	duration := firstCallTime.Sub(time.Now().Local())

	go func() {
		time.Sleep(duration)
		for {
			f()
			// Следующий запуск через сутки.
			time.Sleep(time.Hour * 24)
		}
	}()

	return nil
}

// Ваша функция.
func myfunc() {
	fmt.Printf("+ %v\n", time.Now())
}

// Пример использования.
func mainShedule() {
	err := callAt(0, 0, 0, myfunc)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// Эмуляция дальнейшей работы программы.
	time.Sleep(time.Hour * 24)
}
