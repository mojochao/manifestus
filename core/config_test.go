package core

import (
	"path"
	"reflect"
	"testing"
)

var goodRenderfilePath = path.Join("..", "testdata", "renderfile.yaml")

var goodRendermanConfig = Config{
	Path: goodRenderfilePath,
	Renderfile: Renderfile{
		Schema: "v1",
		Apps: []App{
			{
				Name:     "cert-manager",
				Disabled: false,
				Releases: []Release{
					{
						Name: "cert-manager",
					},
				},
				Bundles: []Bundle{
					{
						Name: "crds",
						Data: map[string]string{
							"app_version": "v1.16.2",
							"base_uri":    "github.com/cert-manager/cert-manager/releases/download",
						},
						Sources: []string{
							"https://{base_uri}/{app_version}/cert-manager.crds.yaml",
						},
					},
				},
			},
			{
				Name:     "external-dns",
				Disabled: false,
				Releases: []Release{
					{
						Name: "external-dns",
					},
				},
			},
		},
	},
}

func TestLoadConfig(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{
			name:    "should load config file",
			args:    args{filePath: goodRenderfilePath},
			want:    goodRendermanConfig,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("LoadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBundle_URLs(t *testing.T) {
	type fields struct {
		Name    string
		Data    map[string]string
		Sources []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "should return expanded paths",
			fields: fields{
				Name: "crds",
				Data: map[string]string{
					"app_version": "v1.16.2",
					"base_uri":    "github.com/cert-manager/cert-manager/releases/download",
				},
				Sources: []string{
					"https://{base_uri}/{app_version}/cert-manager.crds.yaml",
				},
			},
			want: []string{
				"https://github.com/cert-manager/cert-manager/releases/download/v1.16.2/cert-manager.crds.yaml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Bundle{
				Name:    tt.fields.Name,
				Data:    tt.fields.Data,
				Sources: tt.fields.Sources,
			}
			got, err := b.URLs()
			if (err != nil) != tt.wantErr {
				t.Errorf("URLs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("URLs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
