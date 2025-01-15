package core

import (
	"path"
	"testing"
)

func Test_getOutputFilePath(t *testing.T) {
	type args struct {
		appName string
		srcName string
		srcType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should return output file path",
			args: args{
				appName: "cert-manager",
				srcName: "crds",
				srcType: "bundle",
			},
			want: path.Join("cert-manager", "crds.bundle.manifest.yaml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOutputFilePath(tt.args.appName, tt.args.srcName, tt.args.srcType); got != tt.want {
				t.Errorf("getOutputFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
