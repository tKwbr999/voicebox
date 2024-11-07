package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CreateDirAndRemoveFiles(path string) error {
	// ディレクトリを作成
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("ディレクトリの作成に失敗しました: %v", err)
	}

	// ディレクトリ内の全てのファイルを削除
	files, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return fmt.Errorf("ディレクトリ内のファイルを取得中にエラーが発生しました: %v", err)
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return fmt.Errorf("ファイルの削除に失敗しました: %v", err)
		}
	}

	return nil
}

func ExtractLines(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("ファイルを開く際にエラーが発生しました: ", err)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "。")
		if len(parts) > 1 {
			// partsの最後の要素が空文字の場合は取り除く
			if parts[len(parts)-1] == "" {
				parts = parts[:len(parts)-1]
			}
		}
		lines = append(lines, parts...)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("ファイルをスキャン中にエラーが発生しました: ", err)
		return nil, err
	}
	return lines, nil
}

func WriteScriptFile(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %v", err)
	}
	defer file.Close()
	// linesを台本に書き込み
	for _, script := range lines {
		// 無音行の場合は改行のみを書き込む
		// 無音行の判定は行の開始が"..."となるかで判定を行う
		if strings.HasPrefix(script, "...") {
			_, err := file.WriteString("\n")
			if err != nil {
				return fmt.Errorf("台本への書き込みに失敗しました: %v", err)
			}
			continue
		}

		_, err := file.WriteString(script + "\n")
		if err != nil {
			return fmt.Errorf("台本への書き込みに失敗しました: %v", err)
		}
	}
	fmt.Println("-----台本出力完了")
	return nil
}

func ReadInDir(qName string) ([]string, error) {
	files, err := filepath.Glob("./in/*.txt")
	if err != nil {
		fmt.Println("ファイルを取得中にエラーが発生しました:", err)
		return nil, err
	}
	return files, nil
}

func SaveFile(contents []byte, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("音声ファイルの作成中にエラーが発生しました: %v", err)
	}
	defer file.Close()

	_, err = file.Write(contents)
	if err != nil {
		return fmt.Errorf("音声データのファイルへの書き込み中にエラーが発生しました: %v", err)
	}

	return nil
}
