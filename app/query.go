package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func Audio(text string, speakerID int) ([]byte, error) {
	endpoint := fmt.Sprintf("http://localhost:50021/audio_query?text=%s&speaker=%d", url.QueryEscape(text), speakerID)
	resp, err := http.Post(endpoint, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func Synthesize(query []byte, speakerID int) ([]byte, error) {
	endpoint := fmt.Sprintf("http://localhost:50021/synthesis?speaker=%d", speakerID)
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
