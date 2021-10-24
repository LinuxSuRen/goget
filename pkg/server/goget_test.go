package server

import (
	"net/url"
	"testing"
)

func Test_getBranch(t *testing.T) {
	type args struct {
		httpURL *url.URL
	}
	tests := []struct {
		name       string
		args       args
		wantBranch string
	}{{
		name: "no query",
		args: args{
			httpURL: &url.URL{},
		},
		wantBranch: "master",
	}, {
		name: "with valid branch",
		args: args{
			httpURL: &url.URL{
				RawQuery: "branch=test",
			},
		},
		wantBranch: "test",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBranch := getBranch(tt.args.httpURL); gotBranch != tt.wantBranch {
				t.Errorf("getBranch() = %v, want %v", gotBranch, tt.wantBranch)
			}
		})
	}
}

func Test_pair(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{{
		name: "valid",
		args: args{
			key:   "key",
			value: "value",
		},
		want: "key=value",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pair(tt.args.key, tt.args.value); got != tt.want {
				t.Errorf("pair() = %v, want %v", got, tt.want)
			}
		})
	}
}
