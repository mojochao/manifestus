package core

import "testing"

func Test_expandTemplate(t *testing.T) {
	type args struct {
		s    string
		data map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should expand template",
			args: args{
				s: "{greeting}, {name}!",
				data: map[string]string{
					"greeting": "Hello",
					"name":     "World",
				},
			},
			want: "Hello, World!",
		},
		{
			name: "should expand template",
			args: args{
				s: "{greeting}, {name}!",
				data: map[string]string{
					"greeting": "Hello",
				},
			},
			want:    "Hello, {name}!",
			wantErr: true,
		},
		{
			name: "should expand template with multiple parts",
			args: args{
				s: "https://{base_uri}/{version}/resource/foo",
				data: map[string]string{
					"version":  "v1.0.0",
					"base_uri": "example.com/api",
				},
			},
			want: "https://example.com/api/v1.0.0/resource/foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandTemplate(tt.args.s, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("expandTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
