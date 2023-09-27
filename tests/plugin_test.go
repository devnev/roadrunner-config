package tests

import (
	"syscall"
	"testing"

	"github.com/roadrunner-server/config/v4"
	"github.com/stretchr/testify/require"
)

func TestEnvArr(t *testing.T) {
	require.NoError(t, syscall.Setenv("REDIS_HOST_1", "localhost:2999"))
	require.NoError(t, syscall.Setenv("REDIS_HOST_2", "localhost:2998"))
	p := &config.Plugin{
		Prefix:  "rr",
		Path:    "configs/.rr-env-arr.yaml",
		Version: "2.11.3",
	}

	err := p.Init()
	require.NoError(t, err)

	str := p.Get("redis.addrs")
	if _, ok := str.([]string); !ok {
		t.Fatal("not a slice")
	}

	require.Len(t, str.([]string), 2)

	if str.([]string)[0] != "localhost:2999" && str.([]string)[0] != "localhost:2998" {
		t.Fatalf("not expanded")
	}

	if str.([]string)[1] != "localhost:2999" && str.([]string)[1] != "localhost:2998" {
		t.Fatalf("not expanded")
	}
}

func TestVersions(t *testing.T) {
	// rr 2.10, config no version
	p := &config.Plugin{
		Prefix:  "rr",
		Path:    "configs/.rr-no-version.yaml",
		Version: "2.10",
	}

	err := p.Init()
	require.Error(t, err)
}