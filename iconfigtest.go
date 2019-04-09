package goonce_config_consul

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/untillpro/godif"
	"github.com/untillpro/igoonce/iconfig"
	"testing"
)

type testConfig struct {
	Param1 string
	Param2 int
	Param3 bool
	Param4 []string
	Param5 map[string]float64
}

type minTestConfig struct {
	Param1 string
	Param2 int
	Param4 []string
	Param5 map[string]float64
}

type maxTestConfig struct {
	Param1 string
	Param2 int
	Param3 bool
	Param4 []string
	Param5 map[string]float64
	Param6 map[string]interface{}
}

//InitIConfigImplementation should be impl in all child tests and return ctx to use current interface implementation
var InitIConfigImplementation func() context.Context

func TestIConfig(t *testing.T) {
	require.NotNil(t, InitIConfigImplementation, "Need to provide function InitIConfigImplementation to init "+
		"current iconfig implementation")
	TestPutGet(t)
	TestNilConfig(t)
	TestNotPointerInGet(t)
	TestGetWrongStruct(t)
	TestPutGetDifferentStructs(t)
}

var testConfig1 = testConfig{"ac", 3, true, []string{"assert", "b", "c"},
	map[string]float64{"assert": 1.1, "b": 2.2}}

func TestPutGet(t *testing.T) {
	ctx := InitIConfigImplementation()
	defer godif.Reset()
	err := iconfig.PutCurrentAppConfig(ctx, &testConfig1)
	require.Nil(t, err, "Can't put test config to KV! Config: ", err)
	var b testConfig
	err = iconfig.GetCurrentAppConfig(ctx, &b)
	require.Nil(t, err, "Can't get test config from KV! Config: ", err)
	require.True(t, cmp.Equal(testConfig1, b), "Structs must be equal! ", testConfig1, b)
	require.False(t, cmp.Equal(&ctx, &b))
}

func TestNilConfig(t *testing.T) {
	ctx := InitIConfigImplementation()
	defer godif.Reset()
	var config *testConfig = nil
	err := iconfig.PutCurrentAppConfig(ctx, config)
	require.NotNil(t, err)
}

func TestNotPointerInGet(t *testing.T) {
	ctx := InitIConfigImplementation()
	defer godif.Reset()
	var b testConfig
	err := iconfig.GetCurrentAppConfig(ctx, b)
	require.NotNil(t, err)
}

func TestGetWrongStruct(t *testing.T) {
	ctx := InitIConfigImplementation()
	defer godif.Reset()
	err := iconfig.PutCurrentAppConfig(ctx, &testConfig1)
	require.Nil(t, err, "Can't put test config to KV! Config: ", err)

	//try to unmarshal config to wrong struct
	var b error
	err = iconfig.GetCurrentAppConfig(ctx, &b)
	require.Nil(t, b)
	require.NotNil(t, err)
}

func TestPutGetDifferentStructs(t *testing.T) {
	ctx := InitIConfigImplementation()
	defer godif.Reset()
	err := iconfig.PutCurrentAppConfig(ctx, &testConfig1)
	require.Nil(t, err, "Can't put test config to KV! Config: ", err)

	var b minTestConfig

	err = iconfig.GetCurrentAppConfig(ctx, &b)
	require.Nil(t, err, "Can't get test config from KV! Config: ", err)
	require.True(t, !cmp.Equal(testConfig1, b), "Structs must be unequal! ", testConfig1, b)

	//all presented values in minTestConfig are equal
	require.Equal(t, testConfig1.Param1, b.Param1)
	require.Equal(t, testConfig1.Param2, b.Param2)
	require.Equal(t, testConfig1.Param4, b.Param4)
	require.Equal(t, testConfig1.Param5, b.Param5)

	var c maxTestConfig

	err = iconfig.GetCurrentAppConfig(ctx, &c)
	require.Nil(t, err, "Can't get test config from KV! Config: ", err)
	require.True(t, !cmp.Equal(testConfig1, c), "Structs must be unequal! ", testConfig1, b)

	//all presented values in minTestConfig are equal
	require.Equal(t, testConfig1.Param1, c.Param1)
	require.Equal(t, testConfig1.Param2, c.Param2)
	require.Equal(t, testConfig1.Param4, c.Param4)
	require.Equal(t, testConfig1.Param5, c.Param5)
	require.True(t, len(c.Param6) == 0)
}
