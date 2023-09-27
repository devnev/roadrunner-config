package tests

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	configImpl "github.com/roadrunner-server/config/v4"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/rpc/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViperProvider_Init(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = nil

	err := cont.Register(vp)
	require.NoError(t, err)

	err = cont.Register(&Foo{})
	require.NoError(t, err)

	err = cont.Init()
	require.NoError(t, err)

	ch, err := cont.Serve()
	require.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestViperProvider_OldConfig(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{
		Path:   "configs/.rr-old.yaml",
		Prefix: "rr",
	}

	err := cont.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Foo{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigOverwriteExpandEnv(t *testing.T) {
	container := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen=tcp://${RPC_VAL:-127.0.0.1:6001}"}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	_, err = container.Serve()
	assert.NoError(t, err)
	_ = container.Stop()
}

func TestConfigOverwriteFail(t *testing.T) {
	container := endure.New(slog.LevelDebug)
	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen=tcp//not_exist"}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteFail_2(t *testing.T) {
	container := endure.New(slog.LevelDebug)
	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen="}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteFail_3(t *testing.T) {
	container := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"="}

	err := container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.Error(t, err)
}

func TestConfigOverwriteValid(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = []string{"rpc.listen=tcp://127.0.0.1:36643"}

	err := cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariables(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	err := os.Setenv("SUPER_RPC_ENV", "tcp://127.0.0.1:36643")
	assert.NoError(t, err)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env.yaml"
	vp.Prefix = "rr"

	err = cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariables2(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env2.yaml"
	vp.Prefix = "rr"

	err := cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo4{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariables3(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	_ = os.Setenv("RPC_PORT", "6001")

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env3.yaml"
	vp.Prefix = "rr"

	err := cont.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo5{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}

func TestConfigEnvVariablesFail(t *testing.T) {
	container := endure.New(slog.LevelDebug)

	err := os.Setenv("SUPER_RPC_ENV", "tcp://127.0.0.1:6065")
	assert.NoError(t, err)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr-env.yaml"
	vp.Prefix = "rr"

	err = container.RegisterAll(
		&logger.Plugin{},
		&rpc.Plugin{},
		vp,
		&Foo2{},
	)
	assert.NoError(t, err)

	err = container.Init()
	assert.NoError(t, err)

	_, err = container.Serve()
	assert.Error(t, err)
}

func TestConfigProvider_GeneralSection(t *testing.T) {
	cont := endure.New(slog.LevelDebug)

	vp := &configImpl.Plugin{}
	vp.Path = "configs/.rr.yaml"
	vp.Prefix = "rr"
	vp.Flags = nil
	vp.Timeout = time.Second * 10

	err := cont.Register(vp)
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Register(&Foo3{})
	if err != nil {
		t.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		t.Fatal(err)
	}

	// stop by CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	stopCh <- struct{}{}
	wg.Wait()
}