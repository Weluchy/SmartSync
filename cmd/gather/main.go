package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	outputFile := "smartsync_context.txt"
	out, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Не удалось создать файл: %v\n", err)
		return
	}
	defer out.Close()

	// Папки, которые нам не нужны в контексте
	ignoreDirs := map[string]bool{
		".git": true, "node_modules": true, "dist": true,
		"build": true, ".vscode": true, "db": false, // db оставляем, если там sql
	}

	// Расширения, которые мы пропускаем (бинарники, картинки, линтеры)
	ignoreExt := map[string]bool{
		".exe": true, ".png": true, ".jpg": true, ".svg": true,
		".sum": true, ".json": true, // package-lock.json тоже бывает огромным
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if ignoreDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		// Пропускаем игнорируемые расширения и сам этот скрипт
		if ignoreExt[ext] || info.Name() == outputFile || info.Name() == "gather.go" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Записываем разделитель с именем файла, чтобы мне было легко ориентироваться
		out.WriteString(fmt.Sprintf("\n\n--- FILE: %s ---\n\n", path))
		out.Write(content)

		return nil
	})

	if err != nil {
		fmt.Printf("Произошла ошибка при обходе: %v\n", err)
	} else {
		fmt.Printf("Готово! Все файлы собраны в %s. Можешь загружать его в чат.\n", outputFile)
	}
}
