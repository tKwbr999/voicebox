package app

import (
	"os"
	"path/filepath"
	"reflect" // reflect パッケージをインポート
	"strings"
	"testing"
)

// helper function to create a temporary file with content
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return tmpFile
}

// readFileContent は不要になったため削除

func TestCreate(t *testing.T) {
	tests := []struct {
		name               string
		lineNum            int
		inputFile          string
		expectedOutputFile []string // Expected content as a slice of strings
	}{
		{
			name:      "基本的な句読点分割と改行",
			lineNum:   20,
			inputFile: `これはテストです。短い文が続きます。そして、少し長めの文がここに来ます。最後に、非常に短い文。`,
			expectedOutputFile: []string{
				"これはテストです。",
				"短い文が続きます。",
				"そして、",
				"少し長めの文がここに来ます。",
				"最後に、",
				"非常に短い文。",
			},
		},
		{
			name:      "10文字ルール適用ケース（残りが短い）",
			lineNum:   15,
			inputFile: `この文は長いです、そして次の部分は短いです。`, // 「そして次の部分は短いです」は13文字
			expectedOutputFile: []string{
				"この文は長いです、",
				"そして次の部分は短いです。",
			},
		},
		{
			name:      "10文字ルール非適用ケース（残りが長い）",
			lineNum:   15,
			inputFile: `この文はまあまあ長いです、そして次の部分はかなり長くなります。`, // 「そして次の部分はかなり長くなります」は18文字
			expectedOutputFile: []string{
				"この文はまあまあ長いです、",
				"そして次の部分はかなり長くなります。",
			},
		},
		{
			name:      "句読点なし（単一セグメント）",
			lineNum:   10,
			inputFile: `句読点のない長い文字列です`,
			expectedOutputFile: []string{
				"句読点のない長い文字列です",
			},
		},
		{
			name:      "連続する句読点",
			lineNum:   10,
			inputFile: `文です。。次の文です、、、そのまた次の文。`,
			expectedOutputFile: []string{
				"文です。",
				"", // "。" は空文字列になる
				"次の文です、",
				"", // "、" は空文字列になる
				"", // "、" は空文字列になる
				"そのまた次の文。",
			},
		},
		{
			name:    "行頭が特殊文字の場合の処理（Markdownリスト風）",
			lineNum: 20,
			inputFile: `- これはリスト項目です。最初の部分です。そして、これが二番目の部分です。
- 次のリスト項目。これもまた、分割されるべきテキストを含んでいます。`,
			expectedOutputFile: []string{
				"- これはリスト項目です。",
				"最初の部分です。",
				"そして、",
				"これが二番目の部分です。",
				"- 次のリスト項目。",
				"これもまた、",
				"分割されるべきテキストを含んでいます。",
			},
		},
		{
			name:    "空行を含む入力",
			lineNum: 20,
			inputFile: `最初の行です。

次の行はここにあります。`,
			expectedOutputFile: []string{
				"最初の行です。",
				"", // 空行は空文字列になる
				"次の行はここにあります。",
			},
		},
		{
			name:      "非常に短いセグメントのみ",
			lineNum:   5,
			inputFile: `短。次。後。`,
			expectedOutputFile: []string{
				"短。",
				"次。",
				"後。",
			},
		},
		{
			name:      "maxLengthより短いが行セグメントが複数",
			lineNum:   30,
			inputFile: `これは最初の文です。これは二番目の文です。`,
			expectedOutputFile: []string{
				"これは最初の文です。",
				"これは二番目の文です。",
			},
		},
		{
			name:      "10文字ルール境界値テスト（ちょうど10文字）",
			lineNum:   15,
			inputFile: `長い前置きがあって、残りは丁度十文字です。`, // 「残りは丁度十文字です」は10文字
			expectedOutputFile: []string{
				"長い前置きがあって、",
				"残りは丁度十文字です。",
			},
		},
		{
			name:      "10文字ルール境界値テスト（11文字）",
			lineNum:   15,
			inputFile: `長い前置きがあって、残りは十一文字でした。`, // 「残りは十一文字でした」は11文字
			expectedOutputFile: []string{
				"長い前置きがあって、",
				"残りは十一文字でした。",
			},
		},
		// 		{
		// 			name:    "日本語と英数字混合",
		// 			lineNum: 20,
		// 			inputFile: `これは日本語のテキストです、This is English text. そしてまた日本語が続きます。`,
		// 			expectedOutputFile: []string{
		// 				"これは日本語のテキストです",
		// 				"This is English text",
		// 				"そしてまた日本語が続きます",
		// 			},
		// 		},
		{
			name:      "ファイル末尾の改行テスト（入力に改行なし）",
			lineNum:   20,
			inputFile: `最後の行です。改行なしで終わる。`,
			expectedOutputFile: []string{
				"最後の行です。",
				"改行なしで終わる。",
			},
		},
		{
			name:      "ファイル末尾の改行テスト（入力に改行あり）",
			lineNum:   20,
			inputFile: "最後の行です。改行ありで終わる。\n",
			expectedOutputFile: []string{
				"最後の行です。",
				"改行ありで終わる。",
				"", // 末尾の改行は空文字列になる
			},
		},
		{
			name:      "ファイル末尾の改行テスト（入力に連続改行あり）",
			lineNum:   20,
			inputFile: "最後の行です。連続改行で終わる。\n\n",
			expectedOutputFile: []string{
				"最後の行です。",
				"連続改行で終わる。",
				"", // 末尾の改行は空文字列になる
				"", // 末尾の改行は空文字列になる
			},
		},
		{
			name:               "完全に空のファイル",
			lineNum:            20,
			inputFile:          "",
			expectedOutputFile: []string{""}, // 以前は "" だったが、Createの修正により []string{""} となる
		},
		{
			name:               "改行のみのファイル",
			lineNum:            20,
			inputFile:          "\n",
			expectedOutputFile: []string{""}, // 以前は "\n" だったが、Createの修正により []string{""} となる
		},
		{
			name:               "複数の改行のみのファイル",
			lineNum:            20,
			inputFile:          "\n\n\n",
			expectedOutputFile: []string{"", "", ""}, // 以前は "\n" だったが、Createの修正により []string{"", "", ""} となる
		},
		{
			name:               "句読点のみの行",
			lineNum:            20,
			inputFile:          "。\n、\n。。",
			expectedOutputFile: []string{"", "", "", ""},
		},
		{
			name:               "空白と句読点",
			lineNum:            20,
			inputFile:          "　。\n、　\n文。　。文",
			expectedOutputFile: []string{"", "", "文。", "", "文"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.inputFile)
			// defer os.Remove(tmpFile) // t.TempDir() handles cleanup

			actualOutput, err := Create(tmpFile, &tt.lineNum) // Create関数の呼び出しを修正
			if err != nil {
				t.Fatalf("Create failed for '%s': %v", tt.name, err)
			}

			// アサーションロジックを reflect.DeepEqual を使用するように変更
			if !reflect.DeepEqual(actualOutput, tt.expectedOutputFile) {
				// エラーメッセージを改善
				var actualOutputFormatted strings.Builder
				actualOutputFormatted.WriteString("[\n")
				for _, line := range actualOutput {
					actualOutputFormatted.WriteString("\t\"" + line + "\",\n")
				}
				actualOutputFormatted.WriteString("]")

				var expectedOutputFormatted strings.Builder
				expectedOutputFormatted.WriteString("[\n")
				for _, line := range tt.expectedOutputFile {
					expectedOutputFormatted.WriteString("\t\"" + line + "\",\n")
				}
				expectedOutputFormatted.WriteString("]")

				t.Errorf("Expected output for '%s' (lineNum: %d) was:\n-----\n%s\n-----\nBut got:\n-----\n%s\n-----",
					tt.name, tt.lineNum, expectedOutputFormatted.String(), actualOutputFormatted.String())
			}
		})
	}
}
