package altnrslog

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"testing"
)

func TestFromContext(t *testing.T) {
	txHandler := &TransactionalHandler{}
	txLogger := slog.New(txHandler)
	jsonHandler := &slog.JSONHandler{}
	jsonLogger := slog.New(jsonHandler)

	type args struct {
		ctx context.Context
	}
	type want struct {
		logger *slog.Logger
		err    error
	}
	type test struct {
		args args
		want want
	}
	tests := map[string]test{
		"happy-path": {
			args: args{
				ctx: context.WithValue(context.Background(), loggerKey{}, txLogger),
			},
			want: want{
				logger: txLogger,
			},
		},
		"unhappy-path: json logger": {
			args: args{
				ctx: context.WithValue(context.Background(), loggerKey{}, jsonLogger),
			},
			want: want{
				err: ErrNotStored,
			},
		},
		"unhappy-path: no logger": {
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: ErrNotStored,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := FromContext(tt.args.ctx)
			if !errors.Is(err, tt.want.err) {
				t.Errorf("FromContext() error = %v, want %v", err, tt.want.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want.logger) {
				t.Errorf("FromContext() got = %v, want %v", got, tt.want.logger)
			}
		})
	}
}

func TestStoreToContext(t *testing.T) {
	txHandler := &TransactionalHandler{}
	txLogger := slog.New(txHandler)
	jsonHandler := &slog.JSONHandler{}
	jsonLogger := slog.New(jsonHandler)

	type args struct {
		ctx    context.Context
		logger *slog.Logger
	}
	type want struct {
		ctx context.Context
		err error
	}
	type test struct {
		args args
		want want
	}

	tests := map[string]test{
		"happy-path": {
			args: args{
				ctx:    context.Background(),
				logger: txLogger,
			},
			want: want{
				ctx: context.WithValue(context.Background(), loggerKey{}, txLogger),
			},
		},
		"unhappy-path: json logger": {
			args: args{
				ctx:    context.Background(),
				logger: jsonLogger,
			},
			want: want{
				ctx: context.Background(),
				err: ErrInvalidHandler,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := StoreToContext(tt.args.ctx, tt.args.logger)
			if !errors.Is(err, tt.want.err) {
				t.Errorf("StoreToContext() error = %v, want %v", err, tt.want.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want.ctx) {
				t.Errorf("StoreToContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}
