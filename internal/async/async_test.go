package async_test

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/b3ntly/ratelimit"
	"github.com/b3ntly/ratelimit/memory"
	"github.com/b3ntly/salesforce/internal/async"
	"github.com/stretchr/testify/require"
)

type exampleResult struct {
	Result string
	Err    error
}

func example(val string) (string, error) {
	return val, nil
}

type exampleResult2 struct {
	Result string
	Value  int
	Err    error
}

func example2(val string, v int) (string, int, error) {
	return val, v, nil
}

func example3() (string, error) {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(2) == 1 {
		return "", errors.New("error about 50% of the time")
	}

	return "", nil
}

func TestPool_Retry(t *testing.T) {
	t.Skip()
	p := async.New(3, nil, nil)

	var inputs []async.Closure
	for i := 0; i < 50; i++ {
		inputs = append(inputs, async.MustInput(0, example3, &exampleResult{}))
	}

	remaining, err := p.Retry(inputs,
		time.Millisecond*250,
		time.Millisecond*250,
		time.Millisecond*250,
		time.Millisecond*250,
		time.Millisecond*250,
		time.Millisecond*250,
		time.Millisecond*250,
		time.Millisecond*250,
	)

	if err != nil {
		println(err.Error(), len(remaining), 1)
	}
}

func TestPool_All(t *testing.T) {
	t.Skip()
	p := async.New(1, nil, nil)

	a := &exampleResult{}
	b := &exampleResult{}
	c := &exampleResult2{}

	if _, err := p.All(
		async.MustInput(0, example, a, "two"),
		async.MustInput(0, example, b, "three"),
		async.MustInput(0, example2, c, "four", 5),
	); err != nil {
		require.Nil(t, err)
	}

	// or don't
	require.Equal(t, a.Result, "two")
	require.Equal(t, nil, a.Err)
	require.Equal(t, b.Result, "three")
	require.Equal(t, nil, b.Err)
	require.Equal(t, c.Result, "four")
	require.Equal(t, c.Value, 5)
	require.Equal(t, nil, c.Err)
}

func ExamplePool_All() {
	type fooResult struct {
		Result string
		Err    error
	}

	var foo = func(val string) (string, error) {
		return val, nil
	}

	type barResult struct {
		Result string
		Value  int
		Err    error
	}

	var bar = func(val string, v int) (string, int, error) {
		return val, v, nil
	}

	p := async.New(10, nil, nil)

	a := &fooResult{}
	b := &fooResult{}
	c := &barResult{}

	results := []interface{}{a, b, c}

	if _, err := p.All(
		async.MustInput(0, foo, a, "two"),
		async.MustInput(0, foo, b, "three"),
		async.MustInput(0, bar, c, "four", 5),
	); err != nil {
		println(err.Error())
	}

	// process sequentially
	for i := 0; i < len(results); i++ {
		switch v := results[i].(type) {
		case *fooResult:
			println(v.Result)
		case *barResult:
			println(v.Result, v.Value, v.Err == nil)
		default:
			panic(v)
		}
	}
}

func thing(elem int) (int, error) {
	if elem, exists := s.get(elem); exists {
		println("cache hit", elem)
		return elem, nil
	}

	time.Sleep(time.Second)
	println("elem is", elem)

	s.set(elem, elem)
	return elem, nil
}

type thingResult struct {
	One int
	Two error
}

var s = newSampleCache()

type sampleCache struct {
	m map[int]int
	l sync.Mutex
}

func newSampleCache() *sampleCache {
	return &sampleCache{
		m: make(map[int]int),
		l: sync.Mutex{},
	}
}

func (s *sampleCache) get(key int) (int, bool) {
	s.l.Lock()
	elem, exists := s.m[key]
	s.l.Unlock()
	return elem, exists
}

func (s *sampleCache) set(key, value int) {
	s.l.Lock()
	s.m[key] = value
	s.l.Unlock()
}

func TestThings(t *testing.T) {
	t.Skip()
	pool := async.New(10, nil, ratelimit.New(5, time.Second, 5, memory.New()))
	var inputs []async.Closure
	for i := 0; i < 50; i++ {
		n := i
		inputs = append(inputs, async.MustInput(0, thing, &thingResult{}, n))
	}
	for i := 0; i < 50; i++ {
		n := i
		inputs = append(inputs, async.MustInput(0, thing, &thingResult{}, n))
	}
	if _, err := pool.All(inputs...); err != nil {
		println(err.Error())
		t.FailNow()
	} else {
		//t.Log(len(retries), "retries")
	}

	println("done")
}

func TestManyThings(t *testing.T) {
	pool := async.New(5, nil, ratelimit.New(5, time.Second, 5, memory.New()))
	var inputs [][]async.Closure
	for i := 0; i < 10; i++ {
		inputs = append(inputs, []async.Closure{
			async.MustInput(0, thing, &thingResult{}, i),
		})
	}

	for _, input := range inputs {
		_, err := pool.All(input...)
		if err != nil {
			println(err.Error())
		} else {

		}
	}

	for range time.Tick(time.Second) {
		pool.Debug()
	}
}
