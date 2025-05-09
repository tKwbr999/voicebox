package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"voicevox/app"

	"github.com/samber/lo"
)

var speakerID = 1

func main() {
	texts, err := app.Create("/Users/tk/Downloads/yoshioka_haruka.md", lo.ToPtr(40))
	if err != nil {
		panic(err)
	}
	outWav := "wav"
	for i, text := range texts {
		fmt.Println(text)
		filename := fmt.Sprintf("%s/%05d.wav", outWav, i)
		err := app.GenerateAndSaveAudio(text, speakerID, filename)
		if err != nil {
			fmt.Println(err)
		}

	}
	err = app.ConcatAllWavFiles(outWav, "yoshioka_haruka")
	if err != nil {
		fmt.Println(err)
	}

	// // outディレクトリにファイルを作成
	// outDir := "in"
	// os.MkdirAll(outDir, 0755)
	// for i, text := range texts {
	// 	fileName := fmt.Sprintf("%s/%d.txt", outDir, i)
	// 	os.WriteFile(fileName, []byte(text), 0644)
	// }

	// files, err := app.ReadInDir(outDir)
	// if err != nil {
	// 	fmt.Println("ファイルを取得中にエラーが発生しました:", err)
	// 	panic(err)
	// }

	// for i, v := range files {
	// 	filename := fmt.Sprintf("%s/%05d.wav", "wav", i)
	// 	fmt.Println("ファイル番号", filename, time.Now().Format("2006-01-02 15:04:05.000"))
	// 	err := app.GenerateAndSaveAudio(v, speakerID, filename)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// }

}

// func Exec() error {

// 	// ファイルごとに音声合成と保存を実行
// 	for _, path := range files {
// 		err := GenerateAndSaveAudio(path)
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}
// 	}
// 	// 最終的にoutディレクトリにあるwavファイルを昇順で結合
// 	err = app.ConcatAllWavFiles("out", "all")
// 	if err != nil {
// 		fmt.Println("音声ファイルの結合中にエラーが発生しました:", err)
// 		return err
// 	}

// 	return nil
// }

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
