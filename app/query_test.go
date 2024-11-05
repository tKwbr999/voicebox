package app

import (
	"testing"
)

func TestAudio(t *testing.T) {
	type args struct {
		text      string
		speakerID int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test Empty Wav Generation",
			args: args{text: "、、、、、、、、、、", speakerID: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Audio(tt.args.text, tt.args.speakerID)
			if err != nil {
				t.Log(err)
				t.Fail()
			}
			if len(got) == 0 {
				t.Log("Empty wav")
				t.Fail()
			}
			wav, err := Synthesize(got, tt.args.speakerID)
			if err != nil {
				t.Log(err)
				t.Fail()
			}
			// wavをファイル保存
			err = SaveFile(wav, "./test_audio.wav")
			if err != nil {
				t.Log(err)
				t.Fail()
			}
		})
	}
}
