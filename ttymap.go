// Package ttymap is a utility to map keypresses (runes) to callbacks for CLI applications.
// The rune pressed is sent to the callback, but can be ignored.
package ttymap

import (
	"fmt"
	"sync"

	"github.com/mattn/go-tty"
)

// KeyFunc is a type used for callbacks upon keypresses.
type KeyFunc func(rune)

// TtyMap is cool, eh?
type TtyMap struct {
	keyMap   map[rune]KeyFunc
	maplock  sync.RWMutex
	doneChan chan struct{}
	tty      *tty.TTY
	runOnce  func()
}

// New returns a barely initialized TtyMap.
func New() *TtyMap {
	z := &TtyMap{
		keyMap:   make(map[rune]KeyFunc),
		doneChan: make(chan struct{}),
	}
	z.runOnce = sync.OnceFunc(z.run)

	return z
}

// Upsert either updates or adds a func(rune) for keypress of the specified rune.
// This function is safe to call across goros, however if the same rune is used in multiple
// requests, the "last" one "wins".
func (z *TtyMap) Upsert(r rune, f KeyFunc) {
	z.maplock.Lock()
	defer z.maplock.Unlock()

	z.keyMap[r] = f
}

// Remove will delete the key denoted by the specified rune if it exists.
// This function is safe to call across goros. This function is safe to call if one is
// uncertain if the rune has been Upsertted previously.
func (z *TtyMap) Remove(r rune) {
	z.maplock.Lock()
	defer z.maplock.Unlock()

	delete(z.keyMap, r)
}

// Run opens the TTY and waits for a keypress. Run is intended to be executed in its
// own goro, but could be run inline. Will only execute once, but may be called concurrently.
func (z *TtyMap) Run() {
	z.runOnce()
}

func (z *TtyMap) run() {
	var err error
	z.tty, err = tty.Open()
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-z.doneChan:
			return
		default:
			r, err := z.tty.ReadRune()
			if err != nil {
				fmt.Printf("\nError during key read: %s\n", err)
			}
			z.maplock.RLock()
			if f, ok := z.keyMap[r]; ok {
				z.maplock.RUnlock() // immediately, in case f is forever
				f(r)
			} else {
				// doesn't exist
				z.maplock.RUnlock()
			}

		}
	}
}

// Close will signal the Run() loop to end, and close the TTY.
// You must create a New() if you want to continue mapping the keypresses.
func (z *TtyMap) Close() {
	close(z.doneChan)
	z.tty.Close()
}
