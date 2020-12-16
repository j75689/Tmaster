package parser

import (
	"reflect"
	"testing"
)

func TestReplaceSystemVariables(t *testing.T) {
	type args struct {
		config          []byte
		systemVariables interface{}
	}
	case2Name := "1234"
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "TestReplaceSystemVariables Case 1",
			args: args{
				config: []byte(`abcd: #{Name}`),
				systemVariables: struct{ Name string }{
					Name: "1234",
				},
			},
			want:    []byte(`abcd: \"1234\"`),
			wantErr: false,
		},
		{
			name: "TestReplaceSystemVariables Case 2",
			args: args{
				config: []byte(`abcd: #{Name}`),
				systemVariables: struct{ Name *string }{
					Name: &case2Name,
				},
			},
			want:    []byte(`abcd: \"1234\"`),
			wantErr: false,
		},
		{
			name: "TestReplaceSystemVariables Case 3",
			args: args{
				config: []byte(`abcd: #{Count}`),
				systemVariables: struct{ Count int }{
					Count: 456,
				},
			},
			want:    []byte(`abcd: 456`),
			wantErr: false,
		},
		{
			name: "TestReplaceSystemVariables Case 4",
			args: args{
				config: []byte(`abcd: #{C}`),
				systemVariables: struct{ Count int }{
					Count: 456,
				},
			},
			want:    []byte(`abcd: \"C\"`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReplaceSystemVariables(tt.args.config, tt.args.systemVariables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceSystemVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReplaceSystemVariables() = %s, want %s", got, tt.want)
			}
		})
	}
}
