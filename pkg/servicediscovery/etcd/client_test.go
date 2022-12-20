package etcd

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ahfuzhang/cheaprpc/pkg/config"
	"github.com/ahfuzhang/cheaprpc/test"
)

// var addr = []string{"10.129.1xx.xxx:xxx"}

func TestNewClient(t *testing.T) {
	err := test.LoadConfig()
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	cfg := test.GetConfig()
	addrs := cfg.ETCD.Addrs
	type args struct {
		ctx       context.Context
		endpoints []string
		timeout   time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				ctx:       context.Background(),
				endpoints: addrs,
				timeout:   100 * time.Millisecond,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.ctx, tt.args.endpoints, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("%+v", got)
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("NewClient() got = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestClient_RegisterService(t *testing.T) {
	err := test.LoadConfig()
	if err != nil {
		t.Errorf("%+v", err)
		return
	}
	cfg := test.GetConfig()
	addrs := cfg.ETCD.Addrs
	ctx := context.Background()
	ctx1, cancel := context.WithCancel(ctx)
	cli, err := NewClient(ctx1, addrs, 100*time.Millisecond)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	//
	type args struct {
		cfg     *config.Configs
		service string
		addr    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "register first services",
			args: args{
				cfg: &config.Configs{
					Global: config.Global{
						Namespace:     "cheaprpc",
						EnvName:       "test",
						ContainerName: "my-container-1",
						LocalIP:       "127.0.0.1",
					},
					Server: config.Server{
						App:    "ahfuzhang",
						Server: "my_test_server",
					},
				},
				service: "mytestservice",
				addr:    "127.0.0.1:8080",
			},
			wantErr: false,
		},
		{
			name: "register first services-2",
			args: args{
				cfg: &config.Configs{
					Global: config.Global{
						Namespace:     "cheaprpc",
						EnvName:       "test",
						ContainerName: "my-container-3",
						LocalIP:       "127.0.0.1",
					},
					Server: config.Server{
						App:    "ahfuzhang",
						Server: "my_test_server",
					},
				},
				service: "mytestservice",
				addr:    "127.0.0.1:8082",
			},
			wantErr: false,
		},
		{
			name: "register second services",
			args: args{
				cfg: &config.Configs{
					Global: config.Global{
						Namespace:     "cheaprpc",
						EnvName:       "test",
						ContainerName: "my-container-2",
						LocalIP:       "127.0.0.1",
					},
					Server: config.Server{
						App:    "ahfuzhang",
						Server: "my_real_server",
					},
				},
				service: "mytestservice_real",
				addr:    "127.0.0.1:8081",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cli.Register(tt.args.cfg, tt.args.service, tt.args.addr); (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
			out, err := cli.SelectAll(tt.args.cfg, tt.args.service)
			if err != nil {
				t.Errorf("select fail:err=%s", err.Error())
				return
			}
			t.Logf("%+v", out)
			assert.True(t, len(out) > 0)
			inst, err1 := cli.Select(tt.args.cfg, tt.args.service)
			if err1 != nil {
				t.Errorf("select fail:err=%s", err.Error())
				return
			}
			t.Logf("%+v", inst)
		})
	}
	cancel()
	cli.Close()
}

func Test_SyncValue(t *testing.T) {
	a := atomic.Value{}
	a.Load()
	a.Store(nil)
}
