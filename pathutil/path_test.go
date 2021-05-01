package pathutil_test

import (
	"testing"

	"github.com/eloylp/kit/pathutil"
)

//nolint:scopelint
func TestRelativePath(t *testing.T) {
	type args struct {
		root         string
		requiredPath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test subdirectory of root",
			args: args{
				root:         "/var/www/html",
				requiredPath: "/var/www/html/static",
			},
			want:    "static",
			wantErr: false,
		},
		{
			name: "Test file of subdir of root",
			args: args{
				root:         "/var/www/html",
				requiredPath: "/var/www/html/static/main.js",
			},
			want:    "static/main.js",
			wantErr: false,
		},
		{
			name: "Test file of subdir of root (dir traversal)",
			args: args{
				root:         "/var/www/html",
				requiredPath: "/var/www/html/static/sub/../main.js",
			},
			want:    "static/main.js",
			wantErr: false,
		},
		{
			name: "Test file of subdir of root (dir traversal on both parts)",
			args: args{
				root:         "/var/www/html/sub/..",
				requiredPath: "/var/www/html/static/sub/../main.js",
			},
			want:    "static/main.js",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pathutil.RelativePath(tt.args.root, tt.args.requiredPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("RelativePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RelativePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:scopelint
func TestPathInRoot(t *testing.T) {
	type args struct {
		docRoot string
		path    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Should pass when both paths are absolutes and path is inside the document root",
			args{"/var/www", "/var/www/uploaded"},
			false,
		},
		{
			"Should pass when both paths are relatives and path is inside the document root",
			args{"images", "images/uploaded"},
			false,
		},
		{
			"Should pass when both paths are relatives and path is inside the document root (dir traversing)",
			args{"images", "images/uploaded/.."},
			false,
		},
		{
			"Should pass when both paths are relatives and path is inside the document root (dir traversing)",
			args{"images", "images/uploaded/.."},
			false,
		},

		{
			"Should NOT pass when both paths are absolutes and path is outside the document root",
			args{"/var/www", "/var/non-allowed-dir/uploaded"},
			true,
		},
		{
			"Should NOT pass when both paths are relatives and path is outside the document root",
			args{"images", "photos/upload"},
			true,
		},
		{
			"Should NOT pass when both paths are relatives and path is outside the document root (dir traversing)",
			args{"images", "images/upload/../.."},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := pathutil.PathInRoot(tt.args.docRoot, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("PathInRoot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
