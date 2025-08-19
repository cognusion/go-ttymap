package ttymap

func Example() {

	var (
		runeP rune
		runeQ rune
	)

	z := New()
	go z.Run() // launch it async

	// closer stops Gnome and Vinz Clortho
	closer := func(r rune) {
		// close(someChan) or something.Stop() to shut your app down
		z.Close() // Stupid important that this gets called, but waiting for a defer is generally late.
	}

	pause := func(r rune) {
		// pause your app?
	}

	z.Upsert(runeP, pause)  // toggles pause status
	z.Upsert(runeQ, closer) // closes all the things

	// start your app up

	// z.Run() // blocks
}
