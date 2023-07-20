package handler

import "testing"

func TestFile_exists(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "test1",
			args:    args{path: "../conf"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "test2",
			args:    args{path: "test"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := exists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("File.exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("File.exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
