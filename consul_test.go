package config

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/untillpro/godif"
	"github.com/untillpro/igoonce/iconfig"
	"math/rand"
	"testing"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func TestConsul(t *testing.T) {
	InitIConfigImplementation = func() context.Context {
		godif.Reset()
		Declare()
		godif.Require(&iconfig.GetCurrentAppConfig)
		godif.Require(&iconfig.PutCurrentAppConfig)
		errs := godif.ResolveAll()
		if len(errs) != 0 {
			panic(errs)
		}
		var err error
		ctx, err := Init(context.Background(), "127.0.0.1", randStringBytes(8), 8500)
		if err != nil {
			panic(err)
		}
		return ctx
	}
	TestIConfig(t)
}

func TestInit(t *testing.T) {
	tss := []struct {
		ctx          context.Context
		host, prefix string
		port         uint16
	}{
		{context.Background(), "127.0.0.1", "", 8500},
		{context.Background(), "", "a", 8500},
		{context.Background(), "127.0.0.1", "a", 0},
		{nil, "127.0.0.1", "a", 123},
	}
	for _, ts := range tss {
		_, err := Init(ts.ctx, ts.host, ts.prefix, ts.port)
		require.NotNil(t, err)
	}
	config, err := Init(context.Background(), "127.0.0.1", "asd", 8500)
	require.NotNil(t, config)
	require.Nil(t, err)
}
