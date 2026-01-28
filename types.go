package miya

import "github.com/zipreport/miya/loader"

type Loader = loader.Loader

type FilterFunc func(value interface{}, args ...interface{}) (interface{}, error)

type TestFunc func(value interface{}, args ...interface{}) (bool, error)
