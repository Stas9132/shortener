package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"reflect"
	"testing"
)

func Test_run(t *testing.T) {
	type args struct {
		pass *analysis.Pass
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{{
		name:    "Empty",
		args:    args{pass: &analysis.Pass{}},
		want:    nil,
		wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := run(tt.args.pass)
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("run() got = %v, want %v", got, tt.want)
			}
		})
	}
}
