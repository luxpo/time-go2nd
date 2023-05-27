package cache

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/golang/mock/gomock"
	"github.com/luxpo/time-go2nd/cache/redis/mocks"
	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -destination=./mocks/mock_redis_cmdable.go -package=mocks github.com/redis/go-redis/v9 Cmdable

func TestNewRedisCache(t *testing.T) {
	type args struct {
		client redis.Cmdable
	}
	tests := []struct {
		name string
		args args
		want *RedisCache
	}{
		{
			name: "ok",
			args: args{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					cmd := mocks.NewMockCmdable(ctrl)
					return cmd
				}(),
			},
			want: &RedisCache{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					cmd := mocks.NewMockCmdable(ctrl)
					return cmd
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRedisCache(tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRedisCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	type fields struct {
		client redis.Cmdable
	}
	type args struct {
		ctx        context.Context
		key        string
		val        any
		expiration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "ok",
			fields: fields{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					status := redis.NewStatusCmd(context.Background())
					status.SetVal("OK")
					cmd := mocks.NewMockCmdable(ctrl)
					cmd.EXPECT().Set(context.Background(), "k1", "v1", time.Second).Return(status)
					return cmd
				}(),
			},
			args: args{
				ctx:        context.Background(),
				key:        "k1",
				val:        "v1",
				expiration: time.Second,
			},
			wantErr: nil,
		},
		{
			name: "timeout error",
			fields: fields{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					status := redis.NewStatusCmd(context.Background())
					status.SetErr(context.DeadlineExceeded)
					cmd := mocks.NewMockCmdable(ctrl)
					cmd.EXPECT().Set(context.Background(), "k1", "v1", time.Second).Return(status)
					return cmd
				}(),
			},
			args: args{
				ctx:        context.Background(),
				key:        "k1",
				val:        "v1",
				expiration: time.Second,
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "not ok error",
			fields: fields{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					status := redis.NewStatusCmd(context.Background())
					status.SetVal("NOT OK")
					cmd := mocks.NewMockCmdable(ctrl)
					cmd.EXPECT().Set(context.Background(), "k1", "v1", time.Second).Return(status)
					return cmd
				}(),
			},
			args: args{
				ctx:        context.Background(),
				key:        "k1",
				val:        "v1",
				expiration: time.Second,
			},
			wantErr: fmt.Errorf("%w, 返回信息 %s", errFailedToSetCache, "NOT OK"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RedisCache{
				client: tt.fields.client,
			}
			err := c.Set(tt.args.ctx, tt.args.key, tt.args.val, tt.args.expiration)
			assert.Equal(t, err, tt.wantErr)
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	type fields struct {
		client redis.Cmdable
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr error
	}{
		{
			name: "ok",
			fields: fields{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					cmd := mocks.NewMockCmdable(ctrl)
					str := redis.NewStringCmd(context.Background())
					str.SetVal("v1")
					cmd.EXPECT().Get(context.Background(), "k1").Return(str)
					return cmd
				}(),
			},
			args: args{
				ctx: context.Background(),
				key: "k1",
			},
			want:    "v1",
			wantErr: nil,
		},
		{
			name: "timeout error",
			fields: fields{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					cmd := mocks.NewMockCmdable(ctrl)
					str := redis.NewStringCmd(context.Background())
					str.SetErr(context.DeadlineExceeded)
					cmd.EXPECT().Get(context.Background(), "k1").Return(str)
					return cmd
				}(),
			},
			args: args{
				ctx: context.Background(),
				key: "k1",
			},
			wantErr: context.DeadlineExceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RedisCache{
				client: tt.fields.client,
			}
			got, err := c.Get(tt.args.ctx, tt.args.key)
			assert.Equal(t, err, tt.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestRedisCache_Delete(t *testing.T) {
	type fields struct {
		client redis.Cmdable
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				client: func() redis.Cmdable {
					ctrl := gomock.NewController(t)
					cmd := mocks.NewMockCmdable(ctrl)
					int := redis.NewIntCmd(context.Background())
					int.SetVal(1)
					cmd.EXPECT().Del(context.Background(), "k1").Return(int)
					return cmd
				}(),
			},
			args: args{
				ctx: context.Background(),
				key: "k1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RedisCache{
				client: tt.fields.client,
			}
			if err := c.Delete(tt.args.ctx, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RedisCache.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
