package parser

import "testing"

func TestRemoveDoubleQuotes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestRemoveDoubleQuotes Case 1",
			args: args{
				s: `"abcdef"`,
			},
			want: "abcdef",
		},
		{
			name: "TestRemoveDoubleQuotes Case 2",
			args: args{
				s: `"abcdef`,
			},
			want: "\"abcdef",
		},
		{
			name: "TestRemoveDoubleQuotes Case 3",
			args: args{
				s: `"ab"cdef"`,
			},
			want: "ab\"cdef",
		},
		{
			name: "TestRemoveDoubleQuotes Case 4",
			args: args{
				s: `ab"cdef`,
			},
			want: "ab\"cdef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveDoubleQuotes(tt.args.s); got != tt.want {
				t.Errorf("RemoveDoubleQuotes() = %v, want %v", got, tt.want)
			}
		})
	}
}
