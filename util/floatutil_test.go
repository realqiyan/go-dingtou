package util

import (
	"testing"
)

func TestFloatMul(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "0.1*0.2",
			args: args{
				x: 0.1,
				y: 0.2,
			},
			want: 0.02,
		},
		{
			name: "10*0.2",
			args: args{
				x: 10,
				y: 0.2,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatMul(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("FloatMul() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloatAdd(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "0.1+0.2",
			args: args{
				x: 0.1,
				y: 0.2,
			},
			want: 0.3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatAdd(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("FloatAdd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloatSub(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "0.3-0.2",
			args: args{
				x: 0.3,
				y: 0.2,
			},
			want: 0.1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatSub(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("FloatSub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloatDiv(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "0.1/0.5",
			args: args{
				x: 0.1,
				y: 0.5,
			},
			want: 0.2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatDiv(tt.args.x, tt.args.y); got != tt.want {
				t.Errorf("FloatDiv() = %v, want %v", got, tt.want)
			}
		})
	}
}
