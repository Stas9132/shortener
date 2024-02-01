package logger

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestDummy_Debug(t *testing.T) {
	type args struct {
		in0 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Debug(tt.args.in0...)
		})
	}
}

func TestDummy_Debugf(t *testing.T) {
	type args struct {
		in0 string
		in1 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{"", nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Debugf(tt.args.in0, tt.args.in1...)
		})
	}
}

func TestDummy_Error(t *testing.T) {
	type args struct {
		in0 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Error(tt.args.in0...)
		})
	}
}

func TestDummy_Errorf(t *testing.T) {
	type args struct {
		in0 string
		in1 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{"", nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Errorf(tt.args.in0, tt.args.in1...)
		})
	}
}

func TestDummy_Info(t *testing.T) {
	type args struct {
		in0 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Info(tt.args.in0...)
		})
	}
}

func TestDummy_Infof(t *testing.T) {
	type args struct {
		in0 string
		in1 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{"", nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Infof(tt.args.in0, tt.args.in1...)
		})
	}
}

func TestDummy_Trace(t *testing.T) {
	type args struct {
		in0 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Trace(tt.args.in0...)
		})
	}
}

func TestDummy_Tracef(t *testing.T) {
	type args struct {
		in0 string
		in1 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{"", nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Tracef(tt.args.in0, tt.args.in1...)
		})
	}
}

func TestDummy_Warn(t *testing.T) {
	type args struct {
		in0 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Warn(tt.args.in0...)
		})
	}
}

func TestDummy_Warnf(t *testing.T) {
	type args struct {
		in0 string
		in1 []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Empty", args: args{"", nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			d.Warnf(tt.args.in0, tt.args.in1...)
		})
	}
}

func TestDummy_WithField(t *testing.T) {
	type args struct {
		key   string
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want *logrus.Entry
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			if got := d.WithField(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDummy_WithFields(t *testing.T) {
	type args struct {
		fields logrus.Fields
	}
	tests := []struct {
		name string
		args args
		want *logrus.Entry
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dummy{}
			if got := d.WithFields(tt.args.fields); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDummy(t *testing.T) {
	tests := []struct {
		name string
		want *Dummy
	}{
		{name: "Ok", want: NewDummy()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDummy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDummy() = %v, want %v", got, tt.want)
			}
		})
	}
}
