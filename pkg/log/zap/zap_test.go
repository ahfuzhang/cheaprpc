package zap

import (
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/ahfuzhang/cheaprpc/pkg/config"
)

func TestInit2(t *testing.T) {
	type args struct {
		app    string
		server string
		cfg    *config.Zap
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				app:    "cheaprpc",
				server: "ahfuzhang",
				cfg: &config.Zap{
					Level:          "debug",
					LogPath:        "log/",
					Format:         "text",
					MaxAgeMinutes:  60 * 24,
					RotationSizeMB: 100,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Init(tt.args.app, tt.args.server, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	Debugf("aaa=%d", 123)
	Infof("bbb=%d", 456)
	Warnf("ccc=%s", "789")
	Errorf("ddd=%d", 889)
	SetLogLevel(zapcore.ErrorLevel)
	Debugf("aaa xxxx=%d", 123)
	Infof("bbb xxxx=%d", 456)
	Warnf("ccc xxxx=%s", "789")
	Errorf("ddd xxxx=%d", 889)
}
