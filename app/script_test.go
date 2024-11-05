package app

import (
	"fmt"
	"os"
	"testing"
)

func TestTrim(t *testing.T) {
	type args struct {
		s string
	}

	bytes, err := os.ReadFile("./test/g3_202401_q_script.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	s := string(bytes)

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test Trim",
			args: args{s: s},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Trim(tt.args.s)
			fmt.Println(s)
		})
	}
}
