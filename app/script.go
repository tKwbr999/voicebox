package app

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func Trim(s string) string {
	s = strings.TrimSpace(s)
	// 全角スペースをトリム
	s = strings.Replace(s, "　", "", -1)
	return s
}

// processAndFormatSegment は単一のテキストセグメントを処理し、
// 形態素解析、指定文字数での改行、10文字ルール、句読点の保持と直後の改行を行います。
// 返される文字列スライスの各要素は、最終的にファイルに書き込まれる1行を表し、末尾に \n を含みます。
func processAndFormatSegment(segmentText string, maxLength int, t *tokenizer.Tokenizer) []string {
	trimmedSegment := Trim(segmentText)
	if trimmedSegment == "" {
		// 元のセグメントが句読点のみ（例："。"）の場合、Trimしても空にはならない。
		// 元が空白のみならここで空スライスが返る。
		return []string{}
	}

	var formattedLines []string
	currentLineBuilder := strings.Builder{}
	var bufferForThisSegment []string // 形態素解析と10文字ルールで生成された行（\nなし）

	segmentEndsWithPunctuation := false
	punctuationChar := ""
	textForTokenization := trimmedSegment // Trim済みのセグメントを使用

	if strings.HasSuffix(trimmedSegment, "。") {
		segmentEndsWithPunctuation = true
		punctuationChar = "。"
		textForTokenization = strings.TrimSuffix(trimmedSegment, "。")
	} else if strings.HasSuffix(trimmedSegment, "、") {
		segmentEndsWithPunctuation = true
		punctuationChar = "、"
		textForTokenization = strings.TrimSuffix(trimmedSegment, "、")
	}

	// 句読点を除いた部分が空になる場合（元が "。" や "　。" など）
	// この場合、textForTokenization は Trim すると空になる。
	trimmedTextForTokenization := Trim(textForTokenization)

	if trimmedTextForTokenization != "" {
		tokens := t.Tokenize(trimmedTextForTokenization)
		for i := 0; i < len(tokens); i++ {
			token := tokens[i]
			if token.Class == tokenizer.DUMMY {
				continue
			}
			word := token.Surface
			currentLineRuneCount := utf8.RuneCountInString(currentLineBuilder.String())
			prospectiveRuneCount := currentLineRuneCount + utf8.RuneCountInString(word)

			if currentLineBuilder.Len() > 0 && prospectiveRuneCount > maxLength {
				remainingLength := 0
				for j := i; j < len(tokens); j++ {
					if tokens[j].Class == tokenizer.DUMMY {
						continue
					}
					remainingLength += utf8.RuneCountInString(tokens[j].Surface)
				}

				if remainingLength <= 10 {
					for j := i; j < len(tokens); j++ {
						if tokens[j].Class == tokenizer.DUMMY {
							continue
						}
						currentLineBuilder.WriteString(tokens[j].Surface)
					}
					i = len(tokens) - 1 // Consume all remaining tokens
				} else {
					bufferForThisSegment = append(bufferForThisSegment, currentLineBuilder.String())
					currentLineBuilder.Reset()
					currentLineBuilder.WriteString(word)
				}
			} else {
				currentLineBuilder.WriteString(word)
			}
		}
		if currentLineBuilder.Len() > 0 {
			bufferForThisSegment = append(bufferForThisSegment, currentLineBuilder.String())
		}
		if len(bufferForThisSegment) == 0 && trimmedTextForTokenization != "" {
			// 形態素解析の結果が空でも、元のテキストがあった場合はそれを採用
			bufferForThisSegment = append(bufferForThisSegment, trimmedTextForTokenization)
		}
	}

	if len(bufferForThisSegment) > 0 {
		for idx, line := range bufferForThisSegment {
			if idx == len(bufferForThisSegment)-1 && segmentEndsWithPunctuation {
				formattedLines = append(formattedLines, line+punctuationChar+"\n")
			} else {
				// 句読点で終わるセグメント内の途中行も、単に改行する
				formattedLines = append(formattedLines, line+"\n")
			}
		}
	} else if segmentEndsWithPunctuation { // bufferForThisSegmentが空だが、句読点はある場合 (例: 元が "。" のみ、または "　。" など)
		formattedLines = append(formattedLines, punctuationChar+"\n")
	} else if trimmedTextForTokenization == "" && !segmentEndsWithPunctuation && trimmedSegment != "" {
		// 句読点なし、Trimしたら空になったが、元のセグメントは空ではなかった場合。
		// (例: "   " のような空白のみのセグメント)
		// この場合は、空行として扱われるべきだが、冒頭の Trim(segmentText)=="" で処理されるため、
		// ここには到達しない想定。もし到達した場合は、元のセグメントをそのまま1行とする。
		// ただし、現状のロジックでは `trimmedSegment` を使っているので、この分岐は実質不要。
		// formattedLines = append(formattedLines, trimmedSegment+"\n")
	} else if len(formattedLines) == 0 && trimmedSegment != "" && !segmentEndsWithPunctuation {
		// 句読点なし、形態素解析等で何も生成されなかったが、元のセグメントは空ではなかった場合
		// (例: 形態素解析器が解釈できない特殊文字のみなど)
		formattedLines = append(formattedLines, trimmedSegment+"\n")
	}
	return formattedLines
}

func Create(path string, l *int) ([]string, error) {
	if l == nil {
		defaultLine := 20 // デフォルト値を20に変更（テストケースに合わせる）
		l = &defaultLine
	}
	maxLength := *l

	// ファイルの内容を読み込む
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}
	entireText := string(content)
	// Windowsの改行コード CRLF を LF に統一しておく（テストの一貫性のため）
	// Split前に実行することで、Split結果が \r を含まないようにする
	entireText = strings.ReplaceAll(entireText, "\r\n", "\n")


	// 形態素解析器の準備
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}

	var allFormattedLinesFromSegments []string // 各要素は \n を含む行

	// 元のファイルの行区切りを維持するため、\n で分割
	originalLines := strings.Split(entireText, "\n")

	for _, originalLine := range originalLines {
		trimmedOriginalLine := Trim(originalLine)

		if trimmedOriginalLine == "" {
			allFormattedLinesFromSegments = append(allFormattedLinesFromSegments, "\n") // 空行として追加
			continue
		}

		currentProcessingLine := trimmedOriginalLine // この行の処理対象部分
		for len(currentProcessingLine) > 0 {
			idxJaKuten := -1
			kutenLen := 0
			tempIdxJaKuten := strings.Index(currentProcessingLine, "。")
			if tempIdxJaKuten != -1 {
				idxJaKuten = tempIdxJaKuten
				kutenLen = len("。")
			}

			idxJaTouten := -1
			toutenLen := 0
			tempIdxJaTouten := strings.Index(currentProcessingLine, "、")
			if tempIdxJaTouten != -1 {
				idxJaTouten = tempIdxJaTouten
				toutenLen = len("、")
			}

			var segmentToProcess string

			if idxJaKuten != -1 && (idxJaTouten == -1 || idxJaKuten < idxJaTouten) {
				// 「。」が最初に見つかった句読点
				segmentToProcess = currentProcessingLine[:idxJaKuten+kutenLen]
				currentProcessingLine = currentProcessingLine[idxJaKuten+kutenLen:]
			} else if idxJaTouten != -1 {
				// 「、」が最初に見つかった句読点
				segmentToProcess = currentProcessingLine[:idxJaTouten+toutenLen]
				currentProcessingLine = currentProcessingLine[idxJaTouten+toutenLen:]
			} else {
				// この行にはもう句読点がない
				segmentToProcess = currentProcessingLine
				currentProcessingLine = "" // ループを抜ける
			}

			// segmentToProcess が空白のみや空文字列の場合もあるので Trim する
			// ただし、processAndFormatSegment 内部でも Trim されるので、ここではそのままでも良い。
			// processAndFormatSegment は Trim 後の結果が空なら空スライスを返す。
			if segmentToProcess != "" { // 空セグメントは処理しない
				processedSegmentLines := processAndFormatSegment(segmentToProcess, maxLength, t)
				allFormattedLinesFromSegments = append(allFormattedLinesFromSegments, processedSegmentLines...)
			}
		}
	}

	var resultLines []string
	for _, lineWithNewline := range allFormattedLinesFromSegments {
		line := strings.TrimSuffix(lineWithNewline, "\n")
		// 指示: 「改行のみの行は空文字列として扱う」
		// 指示: 「句読点の連続などによって、実質的に文字を含まない行（改行コードのみに相当する行）が生成される場合、その行はスライス内で空文字列 ("") として表現してください」
		// これらを考慮し、line が "" (元が "\n")、"。"、"、" の場合に空文字列 "" を結果スライスに追加する。
		if line == "" || line == "。" || line == "、" {
			resultLines = append(resultLines, "")
		} else {
			resultLines = append(resultLines, line)
		}
	}

	// 完全に空のファイルの場合、entireText == "" であり、originalLines == []string{""} となる。
	// このとき、allFormattedLinesFromSegments == []string{"\n"} となり、
	// resultLines == []string{""} となる。これは期待される動作。

	// 入力が改行のみで構成される場合の調整
	// 例: 入力 "\n" -> Split で ["", ""], resultLines が ["", ""] になる。期待は [""]。
	// 例: 入力 "\n\n" -> Split で ["", "", ""], resultLines が ["", "", ""] になる。期待は ["", ""]。
	if entireText != "" {
		isOnlyNewlines := true
		for _, r := range entireText {
			if r != '\n' {
				isOnlyNewlines = false
				break
			}
		}
		if isOnlyNewlines && len(resultLines) > 0 {
			// resultLines の末尾の余分な "" を1つ削除する
			resultLines = resultLines[:len(resultLines)-1]
		}
	}

	return resultLines, nil
}
