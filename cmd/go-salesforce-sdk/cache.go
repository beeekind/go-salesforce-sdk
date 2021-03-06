package main

import (
	"sync"

	"github.com/beeekind/go-salesforce-sdk/codegen"
)

type structCache struct {
	m map[string]codegen.Structs
	l *sync.Mutex
}

func (cache *structCache) get(key string) (structs codegen.Structs, exists bool) {
	cache.l.Lock()
	structs, exists = cache.m[key]
	cache.l.Unlock()
	return structs, exists
}

func (cache *structCache) set(key string, value codegen.Structs) {
	cache.l.Lock()
	cache.m[key] = value
	cache.l.Unlock()
}

var describeCache = &structCache{
	m: make(map[string]codegen.Structs),
	l: &sync.Mutex{},
}

var referenceCache = &structCache{
	m: make(map[string]codegen.Structs),
	l: &sync.Mutex{},
}

var descriptionCache = &structCache{
	m: make(map[string]codegen.Structs),
	l: &sync.Mutex{},
}
