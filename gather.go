package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	outputFileName := "smartsync_context.txt"
	outFile, err := os.Create(outputFileName)
	if err != nil {
		fmt.Printf("Ошибка при создании файла: %v\n", err)
		return
	}
	defer outFile.Close()

	// Папки, которые скрипт будет пропускать
	ignoreDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"dist":         true,
		"build":        true,
		".vscode":      true,
		"docs":         true, // если сваггер докс слишком большой
	}

	// Расширения файлов, которые не нужно читать (бинарники, картинки)
	ignoreExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true,
		".png": true, ".jpg": true, ".svg": true, ".ico": true,
		".zip": true, ".pdf": true, ".tar": true, ".gz": true,
		".sum": true, // go.sum обычно не нужен для анализа логики
	}

	fmt.Println("Начинаю сборку проекта...")

	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Пропускаем сам файл скрипта и итоговый файл
		if path == outputFileName || path == "gather.go" {
			return nil
		}

		if d.IsDir() {
			if ignoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ignoreExts[ext] {
			return nil
		}

		// Читаем содержимое файла
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Записываем разделитель с названием файла
		separator := fmt.Sprintf("\n\n========================================\nФАЙЛ: %s\n========================================\n\n", path)
		outFile.WriteString(separator)
		outFile.Write(content)

		return nil
	})

	if err != nil {
		fmt.Printf("Ошибка при обходе директорий: %v\n", err)
	} else {
		fmt.Printf("Готово! Весь код собран в файл: %s\n", outputFileName)
	}
}
