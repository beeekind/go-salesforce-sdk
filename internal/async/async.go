// Package async provides a Promise.All() style interface for concurrent processing
//
// It includes support for ratelimiting, logging, and retries.
//
// All work is defined as a closure wrapped with a func() error type. This allows any function
// to be enqueued in a channel to be processed later - no matter the method signature.
//
// Further, because any unit of work can return an error, we can direct failed work into
// a separate dead letter queue to be retried, or halt all remaining work if an error
// occurs half-way through processing.
//
// These features require managing multiple channels in a fan-in-then-fan-out pattern, and
// with that complexity will come potential for deadlock and other related bugs. Please ensure
// all work passed to this package is non-blocking or will timeout and return an error.
//
// Finally logging is done through the rs/zerolog library as the main purpose of logging
// is for internal development and debugging - and I found zerolog to fully cover my usecase.
// If enough users need to piggy back on the logging functionality I can convert it to a broader
// interface that supports more logging libraries. Just leave an issue on the github page for this
// project.
package async

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
)

const (
	startingInfiniteLoop = int32(iota)
	actionWait
	actionReceived
	actionListen
	closureReceived
	closureSleep
	closureWake
	closureErr
	closureCall
	resultSend
	resultSent
)

const DefaultTimeout = time.Minute 

type (
	// Pool manages a slice of workers that process units of work
	Pool struct {
		workers []*worker
		queue   chan action
		logger  *zerolog.Logger
	}

	// limiter provides a mechanism for ratelimiting worker actions
	limiter interface {
		Allow(key string) (nextAllowed time.Duration, err error)
	}

	// action is a logical grouping of work received through
	// an input channel. All work is processed from the in channel before
	// the worker can move on to the next action. The in channel is expected
	// to be closed by its sender.
	action func() (index int, in chan Closure, out chan Output)

	// worker processes units of work encapsulated via an action as a func()error
	worker struct {
		idx     int
		limiter limiter
		queue   chan action
		exit    chan struct{}
		logger  *zerolog.Logger
		state   *atomic.Int32
	}

	// Closure which wraps another closure, returning the
	// error outcome of the inner closure and a function to repeat
	// the action
	//
	// This obfuscation allows any method to be wrapped and enqueued for
	// execution by a worker process in a different go routine
	Closure func() (err error)

	// Output ...
	Output struct {
		Retry Closure
		Err   error
	}
)

// work loops infinitely on the w.queue and w.exit channels and represents a worker working.
func (w *worker) work() {
	w.state.Store(startingInfiniteLoop)
	for {
		w.logger.Trace().Str("activity", "worker.work").Int("worker", w.idx).Str("status", "WAITING").Msg("")
		w.state.Store(actionWait)
		select {
		case <-w.exit:
			return
		case fn := <-w.queue:
			w.state.Store(actionReceived)
			idx, in, out := fn()
			w.logger.Trace().Str("activity", "worker.work").Int("worker", w.idx).Str("status", "WORKING").Int("job", idx).Msg("")
			w.state.Store(actionListen)
			idx2 := 0
			for fn := range in {
				w.state.Store(closureReceived)
				var err error

				if w.limiter != nil {
					// attempt to get past the limiter, else sleep by the returned duration
					for {
						var nextAllowed time.Duration
						nextAllowed, err = w.limiter.Allow("")
						if err != nil {
							w.logger.Error().Stack().Err(err).Str("activity", "worker.work").Int("worker", w.idx).Str("status", "FAILED_RATELIMIT").Int("job", idx).Err(err).Msg("")
							break
						}

						// limite.Allow() returns a time.Duration, if 0 it means an action is allowed
						if nextAllowed <= 0 {
							break
						}

						w.logger.Trace().Int("worker", w.idx).Str("activity", "worker.work").Str("status", "SLEEP_RATELIMIT").Int("job", idx).Int64("ms", nextAllowed.Milliseconds()).Msg("")
						w.state.Store(closureSleep)
						time.Sleep(nextAllowed)
						w.state.Store(closureWake)
					}
				}

				if err != nil {
					w.state.Store(closureErr)
					w.logger.Error().Err(err).Str("activity", "worker.work").Int("worker", w.idx).Str("status", "FAILED_RATELIMIT").Int("job", idx).Msg("")
					w.state.Store(resultSend)
					out <- Output{fn, err}

				} else {
					w.logger.Trace().Str("activity", "worker.work").Int("worker", w.idx).Str("status", "SENDING").Int("job", idx).Msg("")
					w.state.Store(closureCall)
					err := fn()
					w.state.Store(resultSend)
					out <- Output{fn, err}
					w.logger.Trace().Str("activity", "worker.work").Int("worker", w.idx).Str("status", "SENT").Int("job", idx).Msg("")
				}
				idx2++
				w.state.Store(resultSent)
			}
		}
	}
}

// MustInput ...
func MustInput(timeout time.Duration, fn interface{}, result interface{}, args ...interface{}) Closure {
	input, err := Wrap(timeout, fn, result, args...)
	if err != nil {
		panic(err)
	}

	return input
}

// Wrap ...
func Wrap(timeout time.Duration, fn interface{}, result interface{}, args ...interface{}) (Closure, error) {
	fnValue := reflect.ValueOf(fn)
	if !fnValue.IsValid() {
		return nil, fmt.Errorf("fn passed to NewInput is not a valid value")
	}

	fnType := fnValue.Type()
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("fn passed to NewInput is not reflect.Func type")
	}

	if fnType.NumIn() != len(args) {
		return nil, fmt.Errorf("the number of function arguments (%v) does not match the number of args (%v) passed to NewInput", fnType.NumIn(), len(args))
	}

	resValue := reflect.ValueOf(result)
	if !resValue.IsValid() {
		return nil, fmt.Errorf("result passed to NewInput is not a valid value")
	}

	resElem := resValue.Elem()
	if resElem.NumField() != fnType.NumOut() {
		return nil, fmt.Errorf("the number of struct fields in result (%v) does not match the number of function return values (%v) in fn passed to NewInput", resValue.Elem().NumField(), fnType.NumOut())
	}

	var vArgs []reflect.Value
	for i := 0; i < len(args); i++ {
		// reflect.Call can not handle zero values such as nil and this fixes it
		if args[i] == nil {
			vArgs = append(vArgs, reflect.ValueOf(&args[i]))
			continue
		}

		vArgs = append(vArgs, reflect.ValueOf(args[i]))
	}

	fnName := runtime.FuncForPC(fnValue.Pointer()).Name()

	return func() error {
		exit := make(chan error, 2)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		// go 1 start context
		go func(exit chan error) {
			values := fnValue.Call(vArgs)
			if len(values) == 0 {
				exit <- nil
				return
			}

			for i := 0; i < len(values); i++ {
				resElem.Field(i).Set(values[i])
			}

			// if lastValue is a non-nil error return the error, else return nil
			lastValue := values[len(values)-1]
			if lastValue.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) && !lastValue.IsNil() {
				exit <- lastValue.Interface().(error)
				return
			}

			exit <- nil
		}(exit)

		select {
		case <-ctx.Done():
			cancel() 
			return fmt.Errorf("timeout %s: %s: %w", timeout.String(), fnName, ctx.Err())
		case err := <-exit:
			cancel() 
			if err == nil {
				return nil 
			}

			return fmt.Errorf("%s: %w", fnName, err)
		}

	}, nil
}

// New returns a new instance of pool with numWorkers spawned to serve messages sent to pool.queue
func New(numWorkers int, logger *zerolog.Logger, limiter limiter) *Pool {
	p := &Pool{
		workers: make([]*worker, numWorkers),
		queue:   make(chan action),
		logger:  logger,
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// if no logger is set use the global package logger
	if logger == nil {
		p.logger = &log.Logger
	}

	for i := 0; i < numWorkers; i++ {
		uuid := i
		p.workers[i] = &worker{uuid, limiter, p.queue, make(chan struct{}), p.logger, atomic.NewInt32(startingInfiniteLoop)}
		go p.workers[uuid].work()
	}

	go func(p *Pool) {
		for _ = range time.Tick(time.Second * 5) {
			//p.Debug()
		}
	}(p)

	return p
}

// Debug ...
func (p *Pool) Debug() {
	for _, w := range p.workers {
		var state string
		switch w.state.Load() {
		case startingInfiniteLoop:
			state = "starting"
		case actionWait:
			state = "actionWait"
		case actionReceived:
			state = "actionReceived"
		case actionListen:
			state = "actionListen"
		case closureReceived:
			state = "closureReceived"
		case closureSleep:
			state = "closureSleep"
		case closureWake:
			state = "closureWake"
		case closureErr:
			state = "closureErr"
		case closureCall:
			state = "closureCall"
		case resultSend:
			state = "resultSend"
		case resultSent:
			state = "resultSent"
		}
		p.logger.Debug().Int("id", w.idx).Str("state", state).Msg("")
	}
}

// Stop attempts to stop all running workers.
//
// IMPORTANT: if a worker is stuck sending to its out channel or calling its action() method, then it will not read
// from its exit channel and the go routine which called pool.Stop() will deadlock.
func (p *Pool) Stop() {
	for i := 0; i < len(p.workers); i++ {
		w := p.workers[i]
		w.exit <- struct{}{}
	}
}

// All ...
func (p *Pool) All(inputs ...Closure) (failed []Closure, err error) {
	return p.Exec(false, inputs...)
}

// Attempt ...
func (p *Pool) Attempt(inputs ...Closure) (failed []Closure, err error) {
	return p.Exec(true, inputs...)
}

// Retry calls pool.All() once for each item in backoffs.([]time.Duration). If 0 remaining work is returned from
// pool.All() it will return a nil error. If more work remains to be done it will sleep for backoffs[i].(time.Duration)
// until it has run out of backoff elements. If all backoffs have been exhausted it will return the remaining work and the last
// error returned by pool.All()
func (p *Pool) Retry(inputs []Closure, backoffs ...time.Duration) (remaining []Closure, err error) {
	for _, backoff := range backoffs {
		inputs, err = p.All(inputs...)
		if len(inputs) == 0 {
			break
		}

		time.Sleep(backoff)
	}

	return inputs, err
}

// Exec is a blocking method which generates and asynchronously invokes closures in an optionally ratelimited environment.
//
// If one of the produced closures generates an error all subsequent work will be ignored and the error will be returned.
func (p *Pool) Exec(ignoreErrors bool, inputs ...Closure) (failed []Closure, err error) {
	// use of the reflect package can incur panics
	// this defer/recover statement captures and returns a more usable error message
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("recovered in async.All: %s", x)
			case error:
				err = fmt.Errorf("recovered in async.All: %w", err)
			default:
				// Fallback err (per specs, error strings should be lowercase w/o punctuation
				err = fmt.Errorf("recovered unknown in async.All: %v", r)
			}
		}
	}()

	return p.exec(ignoreErrors, inputs...)
}

// exec enqueues a set of functions and sequentially processes their results once they are completed
func (p *Pool) exec(ignoreErrors bool, funcs ...Closure) (retries []Closure, err error) {
	in := make(chan Closure, len(funcs))
	out := make(chan Output, len(funcs))

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	ready := make(chan struct{})

	go func(ready chan struct{}, ctx context.Context){
		for {
			select {
			case <- time.Tick(time.Second):
				available := 0 
				for _, worker := range p.workers {
					if available > 1 {
						ready <- struct{}{}
						return 
					}

					if worker.state.Load() == actionWait {
						available++ 
					}
				}

			case <- ctx.Done():
				return 
			}
		}

	}(ready, ctx)

	select {
	case <- ctx.Done():
		cancel() 
		return funcs, ctx.Err()
	case <-ready:	
		cancel() 
	}

	// 1) push the in channel to each worker so they may process its work
	for i := 0; i < len(p.workers) - 1; i++ {
		index := i
		go func(index int, queue chan action, in chan Closure, out chan Output) {
			queue <- func() (int, chan Closure, chan Output) {
				return index, in, out
			}
		}(index, p.queue, in, out)
	}

	// 2) assign all work through the in queue
	for i := 0; i < len(funcs); i++ {
		fn := funcs[i]
		in <- fn
		p.logger.Trace().Str("activity", "pool.exec").Int("total", len(funcs)).Str("status", "ENQUEUED").Int("index", i).Msg("")
	}

	p.logger.Trace().Str("activity", "pool.exec").Int("total", len(funcs)).Str("status", "ENQUEUED_ALL").Msg("")
	close(in)

	// 3) process all work
	for i := 0; i < len(funcs); i++ {
		o := <-out

		p.logger.Trace().Str("activity", "pool.exec").Int("index", i).Int("total", len(funcs)).Str("status", "DEQUEUED").Msg("")

		if o.Err != nil {
			retries = append(retries, o.Retry)
		}

		// 3) if an non-nil error is returned drain the remaining work and return the error
		if !ignoreErrors {
			if o.Err != nil {
				//p.logger.Error().Err(o.Err).Str("activity", "pool.exec").Int("index", i).Int("total", len(funcs)).Str("status", "FAILED_ACTION").Msg("")
				//p.logger.Error().Err(o.Err).Str("activity", "pool.exec").Int("index", i).Int("total", len(funcs)).Str("status", "DRAINED").Msg("")
				return retries, fmt.Errorf("%w", o.Err)
			}
		}
	}

	return retries, nil
}
