package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ConcatAllWavFiles は指定されたディレクトリ内の音声ファイルを結合し、1つのファイルに保存します。
func ConcatAllWavFiles(dir string, id string) error {
	// ディレクトリ内のファイルを取得
	audios := filepath.Join(dir, "*.wav")
	files, err := filepath.Glob(audios)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	// ファイル名をソート
	sort.Strings(files)

	// 出力ファイルを作成
	outputFile, err := os.Create(fmt.Sprintf("out/output_%s.wav", id))
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	var totalDataSize int64

	// 各ファイルを結合
	for i, file := range files {
		err = appendFile(outputFile, file, i == 0)
		if err != nil {
			return fmt.Errorf("error appending file %s: %v", file, err)
		}

		// ファイルサイズを取得してデータサイズを更新
		fileInfo, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("error getting file info: %v", err)
		}

		// 最初のファイルはヘッダーを含むので、44バイトを引く
		if i == 0 {
			totalDataSize += fileInfo.Size()
		} else {
			totalDataSize += fileInfo.Size() - 44
		}
	}

	// 出力ファイルのWAVヘッダーを更新
	err = updateWavHeader(outputFile, totalDataSize)
	if err != nil {
		return fmt.Errorf("error updating WAV header: %v", err)
	}

	fmt.Println("All audio files have been concatenated successfully.")
	return nil
}

// appendFile は指定されたファイルを出力ファイルに追加します。
// 最初のファイルの場合はヘッダーを含め、以降のファイルはデータ部分のみを追加します。
func appendFile(outputFile *os.File, inputFilename string, includeHeader bool) error {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer inputFile.Close()

	if includeHeader {
		// 最初のファイルのヘッダーを出力ファイルに書き込む
		_, err = io.Copy(outputFile, inputFile)
		if err != nil {
			return fmt.Errorf("error copying data: %v", err)
		}
	} else {
		// ヘッダーをスキップしてデータ部分のみをコピー
		_, err = inputFile.Seek(44, io.SeekStart)
		if err != nil {
			return fmt.Errorf("error seeking input file: %v", err)
		}
		_, err = io.Copy(outputFile, inputFile)
		if err != nil {
			return fmt.Errorf("error copying data: %v", err)
		}
	}

	return nil
}

// updateWavHeader は出力ファイルのWAVヘッダーを更新して、正しいデータサイズを反映させます。
func updateWavHeader(outputFile *os.File, dataSize int64) error {
	// RIFFチャンクのサイズを更新
	_, err := outputFile.Seek(4, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error seeking output file: %v", err)
	}
	_, err = outputFile.Write([]byte{
		byte(dataSize + 36), byte((dataSize + 36) >> 8),
		byte((dataSize + 36) >> 16), byte((dataSize + 36) >> 24),
	})
	if err != nil {
		return fmt.Errorf("error writing RIFF size: %v", err)
	}

	// dataチャンクのサイズを更新
	_, err = outputFile.Seek(40, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error seeking output file: %v", err)
	}
	_, err = outputFile.Write([]byte{
		byte(dataSize), byte(dataSize >> 8),
		byte(dataSize >> 16), byte(dataSize >> 24),
	})
	if err != nil {
		return fmt.Errorf("error writing data size: %v", err)
	}

	return nil
}

func GenerateAndSaveAudio(text string, speakerID int, outputPath string) error {
	// 音声合成用のクエリを生成
	qa, err := Audio(text, speakerID)
	if err != nil {
		return fmt.Errorf("error generating audio query: %v", err)
	}

	// 音声を合成
	audioData, err := Synthesize(qa, speakerID)
	if err != nil {
		return fmt.Errorf("error synthesizing audio: %v", err)
	}

	// 音声ファイルを保存
	// outディレクトリを作成
	err = os.MkdirAll("out", os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating out directory: %v", err)
	}

	// プロジェクトルートのoutディレクトリにoutput.wavとして保存
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating audio file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(audioData)
	if err != nil {
		return fmt.Errorf("error writing audio data to file: %v", err)
	}

	return nil
}
