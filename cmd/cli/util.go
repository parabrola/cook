package main

import (
	"fmt"
	"os"

	app "github.com/parabrola/cook/internal"
)

// handleGlobalOptions executes the global options logic.
func handleGlobalOptions(opts *app.Options, p *app.Parseable) {
	for _, optionHandler := range opts.Handlers(p) {
		if p == nil && optionHandler.NeedsParser {
			continue
		}

		if code, err := optionHandler.Func(p); code != -1 {
			if err != nil {
				fmt.Println(err)
			}

			os.Exit(code)
		}
	}
}
