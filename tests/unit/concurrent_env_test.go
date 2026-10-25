package miya_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/extensions"
	"github.com/zipreport/miya/loader"
)

func TestConcurrentEnvironmentMutation(t *testing.T) {
	env := miya.NewEnvironment()
	env.AddGlobal("g1", 1)

	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			env.AddGlobal(fmt.Sprint("g", i), i)
		}
	}()
	go func() {
		defer wg.Done()
		_ = env.AddExtension(extensions.NewHelloExtension())
		_ = env.AddExtension(extensions.NewHighlightExtension())
	}()
	for r := 0; r < 2; r++ {
		go func(r int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_, _ = env.RenderString(fmt.Sprint("{{ g1 }}", r, "-", i), miya.NewContext())
			}
		}(r)
	}
	wg.Wait()
}

func TestConfigureInheritanceCacheConcurrent(t *testing.T) {
	sl := loader.NewStringLoader(loader.NewDirectTemplateParser())
	sl.AddTemplate("base.html", `[{% block b %}{% endblock %}]`)
	sl.AddTemplate("child.html", `{% extends "base.html" %}{% block b %}x{% endblock %}`)
	env := miya.NewEnvironment(miya.WithLoader(sl))

	done := make(chan struct{})
	var cfg sync.WaitGroup
	cfg.Add(1)
	go func() {
		defer cfg.Done()
		for {
			select {
			case <-done:
				return
			default:
				env.ConfigureInheritanceCache(time.Minute, time.Minute, 1000)
			}
		}
	}()

	var wg sync.WaitGroup
	for g := 0; g < 2; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 300; i++ {
				got, err := env.RenderTemplate("child.html", miya.NewContext())
				if err != nil || got != "[x]" {
					t.Errorf("got %q, err %v; want [x]", got, err)
					return
				}
			}
		}()
	}
	wg.Wait()
	close(done)
	cfg.Wait()
}
