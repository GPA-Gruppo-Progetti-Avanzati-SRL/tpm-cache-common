package redislks_test

import (
	"context"
	"fmt"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-cache-common/cachelks/redislks"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestNewInstanceWithConfig(t *testing.T) {

	cfg := redislks.Config{
		Name: redislks.RedisDefaultBrokerName,
		Addr: "localhost:6379",
	}

	lks, err := redislks.NewInstanceWithConfig(cfg)
	require.NoError(t, err)

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go set(t, lks, fmt.Sprintf("MSG-%2d", i), fmt.Sprintf("MSG-%2d-Value", i), &wg)
	}

	wg.Add(1)
	go set(t, lks, fmt.Sprintf("num-messages"), 10, &wg)

	t.Log("Waiting for goroutines  put to finish...")
	wg.Wait()

	wg.Add(1)
	go get(t, lks, fmt.Sprintf("num-messages"), &wg)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go get(t, lks, fmt.Sprintf("MSG-%2d", i), &wg)
	}

	t.Log("Waiting for goroutines  put to finish...")
	wg.Wait()

}

func set(t *testing.T, lks *redislks.LinkedService, k string, v interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	err := lks.Set(context.Background(), k, v, cachelks.CacheOptions{})
	if err != nil {
		t.Error(err)
	}

	t.Logf("cached %s --> %v", k, v)
}

func get(t *testing.T, lks *redislks.LinkedService, k string, wg *sync.WaitGroup) {
	defer wg.Done()
	v, err := lks.Get(context.Background(), k, cachelks.CacheOptions{})
	if err != nil {
		t.Error(err)
	}

	if v == nil {
		t.Errorf("no value found for %s", k)
	} else {
		t.Logf("retrieved val %s --> %v of type %T", k, v, v)
	}
}
