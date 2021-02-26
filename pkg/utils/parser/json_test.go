package parser

import (
	"reflect"
	"testing"
)

func TestGetJSONValue(t *testing.T) {
	type args struct {
		cmd       string
		variables interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "TestGetJSONValue Case 1",
			args: args{
				cmd: "${xxx.xx.x}",
				variables: map[string]interface{}{
					"xxx": map[string]interface{}{
						"xx": map[string]interface{}{
							"x": "123",
						},
					},
				},
			},
			want:    "123",
			wantErr: false,
		},
		{
			name: "TestGetJSONValue Case 2",
			args: args{
				cmd: "${xxx.xx.x}",
				variables: map[string]interface{}{
					"xxx": map[string]interface{}{
						"xx": map[string]interface{}{
							"x": true,
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "TestGetJSONValue Case 3",
			args: args{
				cmd: "${xxx.xx.x}",
				variables: map[string]interface{}{
					"xxx": map[string]interface{}{
						"xx": "abc",
					},
				},
			},
			want:    "abc",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJSONValue(tt.args.cmd, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetJSONValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetJSONValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetJSONValue(t *testing.T) {
	type args struct {
		cmd       string
		value     interface{}
		variables interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "TestSetJSONValue Case 1",
			args: args{
				cmd: "${xxx.xx.x}",
				variables: map[string]interface{}{
					"xxx": map[string]interface{}{
						"xx": map[string]interface{}{
							"x": "123",
						},
					},
				},
				value: 123,
			},
			want: map[string]interface{}{
				"xxx": map[string]interface{}{
					"xx": map[string]interface{}{
						"x": 123,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TestSetJSONValue Case 2",
			args: args{
				cmd:   "${xxx.111.x.abc}",
				value: 456,
			},
			want: map[string]interface{}{
				"xxx": map[string]interface{}{
					"111": map[string]interface{}{
						"x": map[string]interface{}{
							"abc": 456,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TestSetJSONValue Case 3",
			args: args{
				cmd: "${xxx.111.x.abc}",
				variables: map[string]interface{}{
					"xxx": map[string]interface{}{
						"222": map[string]interface{}{
							"x": "222x",
						},
					},
				},
				value: 456,
			},
			want: map[string]interface{}{
				"xxx": map[string]interface{}{
					"111": map[string]interface{}{
						"x": map[string]interface{}{
							"abc": 456,
						},
					},
					"222": map[string]interface{}{
						"x": "222x",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "TestSetJSONValue Case 4",
			args: args{
				cmd: "${ttt}",
				variables: map[string]interface{}{
					"xxx": map[string]interface{}{
						"222": map[string]interface{}{
							"x": "222x",
						},
					},
				},
				value: 456,
			},
			want: map[string]interface{}{
				"ttt": 456,
				"xxx": map[string]interface{}{
					"222": map[string]interface{}{
						"x": "222x",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetJSONValue(tt.args.cmd, tt.args.value, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetJSONValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetJSONValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	type args struct {
		config    []byte
		variables interface{}
	}
	tests := []struct {
		name      string
		args      args
		wantReply []byte
		wantErr   bool
	}{
		{
			name: "TestReplaceVariables Case 1",
			args: args{
				config: []byte(`abcd: ${aa}`),
				variables: map[string]interface{}{
					"aa": 1234,
				},
			},
			wantReply: []byte(`abcd: 1234`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 2",
			args: args{
				config: []byte(`abcd: ${xx.xx.abc}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
				},
			},
			wantReply: []byte(`abcd: \"1234\"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 3",
			args: args{
				config: []byte(`abcd: ${xx.kk}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
				},
			},
			wantReply: []byte(`abcd: \"xx.kk\"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 4",
			args: args{
				config: []byte(`abcd: ${xx.xx}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
				},
			},
			wantReply: []byte(`abcd: {\"abc\":\"1234\"}`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 5",
			args: args{
				config: []byte(`abcd: ${xx.xx.ccc.cccc}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": "fdsafasdfasdfsd",
					},
				},
			},
			wantReply: []byte(`abcd: ${xx.xx.ccc.cccc}`),
			wantErr:   true,
		},
		{
			name: "TestReplaceVariables Case 6",
			args: args{
				config: []byte(`abcd: ${xx.xx.ccc.cccc||aa.xx}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": "fadfadsfasd",
					},
					"aa": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
				},
			},
			wantReply: []byte(`abcd: {\"abc\":\"1234\"}`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 6",
			args: args{
				config: []byte(`abcd: ${xx.xx.ccc.cccc || aa.xx|| aaa}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
					"aa": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
				},
			},
			wantReply: []byte(`abcd: {\"abc\":\"1234\"}`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 7",
			args: args{
				config: []byte(`abcd: ${xx.xx.ccc.cccc || aa.xx | replace '\u0026' '&' | replace '\u003c' '<' | replace '\u003e' '>'}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "1234",
						},
					},
					"aa": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "\u003c\u003c\u0026\u003e\u003e",
						},
					},
				},
			},
			wantReply: []byte(`abcd: {\"abc\":\"\\u003c\\u003c\\u0026\\u003e\\u003e\"}`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 8",
			args: args{
				config: []byte(`abcd: ${xx.xx.ccc.cccc || aa.xx.ccc | replace '\u0026' '&' | replace '\u003c' '<' | replace '\u003e' '>' || xx.xx.abc}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "<<&>>",
						},
					},
					"aa": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "\u003c\u003c\u0026\u003e\u003e",
						},
					},
				},
			},
			wantReply: []byte(`abcd: \"<<&>>\"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 9",
			args: args{
				config: []byte(`abcd: ${{xx.xx.abc.cccc || xx.xx.abc | replace '"'}}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "te\"st",
						},
					},
				},
			},
			wantReply: []byte(`abcd: test`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 10",
			args: args{
				config: []byte(`abcd: ${xx.xx.abc.cccc || xx.xx.abc | replace '"'}`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "te\"st",
						},
					},
				},
			},
			wantReply: []byte(`abcd: \"test\"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 11",
			args: args{
				config: []byte(`abcd: "${{xx.xx.abc.cccc || xx.xx.abc | quote }}"`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "te\"st",
						},
					},
				},
			},
			wantReply: []byte(`abcd: "te\"st"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 12",
			args: args{
				config: []byte(`abcd: "${{xx.xx.abc.cccc || xx.xx.abc | unquote }}"`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": `"te\\\"st"`,
						},
					},
				},
			},
			wantReply: []byte(`abcd: "te\"st"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 13",
			args: args{
				config: []byte(`abcd: "${{xx.xx.abc.cccc || xx.xx.abc | base64encode }}"`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "test",
						},
					},
				},
			},
			wantReply: []byte(`abcd: "dGVzdA=="`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 14",
			args: args{
				config: []byte(`abcd: "${{xx.xx.abc.cccc || xx.xx.abc | base64decode | replace t }}"`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "dGVzdA==",
						},
					},
				},
			},
			wantReply: []byte(`abcd: "es"`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 15",
			args: args{
				config: []byte(`abcd: "${{}}"`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "dGVzdA==",
						},
					},
				},
			},
			wantReply: []byte(`abcd: ""`),
			wantErr:   false,
		},
		{
			name: "TestReplaceVariables Case 16",
			args: args{
				config: []byte(`abcd: "${}"`),
				variables: map[string]interface{}{
					"xx": map[string]interface{}{
						"xx": map[string]interface{}{
							"abc": "dGVzdA==",
						},
					},
				},
			},
			wantReply: []byte(`abcd: ""`),
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReply, err := ReplaceVariables(tt.args.config, tt.args.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReplaceVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("ReplaceVariables() = %s, want %s", gotReply, tt.wantReply)
			}
		})
	}
}
