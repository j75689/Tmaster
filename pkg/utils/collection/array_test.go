package collection

import (
	"testing"

	"github.com/j75689/Tmaster/pkg/graph/model"
)

func TestContainsError(t *testing.T) {
	failed := model.ErrorCodeTaskfailed
	all := model.ErrorCodeAll
	type args struct {
		target *model.ErrorCode
		source []*model.ErrorCode
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test Case 1",
			args: args{
				target: &failed,
				source: []*model.ErrorCode{&all},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsError(tt.args.target, tt.args.source); got != tt.want {
				t.Errorf("ContainsError() = %v, want %v", got, tt.want)
			}
		})
	}
}
