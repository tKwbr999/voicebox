package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"voicevox/app"
)

const (
	qName     = "g3_202401_q"
	speakerID = 1 // ずんだもんのノーマル声

)

func init() {
	// 出力ディレクトリを作成
	err := app.CreateDirAndRemoveFiles("out")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 一時ディレクトリを作成
	err = app.CreateDirAndRemoveFiles("tmp")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	startTime := time.Now() // 処理開始時間を記録

	err := Exec()
	if err != nil {
		fmt.Println(err)
		return

	}

	// 処理時間を計測
	elapsedTime := time.Since(startTime)
	fmt.Printf("処理時間: %s\n", elapsedTime)

}

func Exec() error {
	// INディレクトリにあるファイルを読み込む
	files, err := app.ReadInDir(qName)
	if err != nil {
		fmt.Println("ファイルを取得中にエラーが発生しました:", err)
		return err
	}

	// ファイルごとに音声合成と保存を実行
	for _, path := range files {
		err := GenerateAndSaveAudio(path)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	// 最終的にoutディレクトリにあるwavファイルを昇順で結合
	err = app.ConcatAllWavFiles("out", "all")
	if err != nil {
		fmt.Println("音声ファイルの結合中にエラーが発生しました:", err)
		return err
	}

	return nil
}

func GenerateAndSaveAudio(path string) error {
	// ファイルから台本を抽出
	scripts, err := app.ExtractLines(path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// filepathからディレクトリを削除したファイル名だけを取得し、拡張子を取り除く
	filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

	//　文字列を台本としてファイル出力する
	scriptpath := fmt.Sprintf("out/%s_script.txt", filename)
	err = app.WriteScriptFile(scripts, scriptpath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tmpDir := fmt.Sprintf("tmp/%s", filename)
	err = app.CreateDirAndRemoveFiles(tmpDir)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 音声合成と保存を実行 (マルチスレッド化)
	var wg sync.WaitGroup
	// 最大3スレッドまで同時に実行
	// voicevox自体にそれほど処理スピードがないため
	sem := make(chan struct{}, 3)

	for i, v := range scripts {
		wg.Add(1)
		sem <- struct{}{} // スレッドを制限

		go func(i int, v string) {
			defer wg.Done()
			defer func() { <-sem }()

			filename := fmt.Sprintf("%s/%05d.wav", tmpDir, i)
			fmt.Println("ファイル番号", filename, time.Now().Format("2006-01-02 15:04:05.000"))
			err := app.GenerateAndSaveAudio(v, speakerID, filename)
			if err != nil {
				fmt.Println(err)
			}
		}(i, v)
	}

	wg.Wait() // 全てのゴルーチンが完了するのを待つ
	// 例: out/output_0.wav, out/output_1.wav, ... -> out/output.wav
	err = app.ConcatAllWavFiles(tmpDir, filename)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
