/*
Package update replaces the current executable (or any other single-file target) with new contents,
rolling back if the replacement fails.

Checksum and signature verification of the downloaded release is handled by the Validator
of the root selfupdate package, before the asset is decompressed.

# Basic Example

This example shows how to update a program remotely from a URL.

	import (
		"fmt"
		"net/http"

		"github.com/creativeprojects/go-selfupdate/v2/update"
	)

	func doUpdate(url string) error {
		// request the new file
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		err = update.Apply(resp.Body, update.Options{})
		if err != nil {
			if rerr := update.RollbackError(err); rerr != nil {
				fmt.Printf("Failed to rollback from bad update: %v\n", rerr)
			}
		}
		return err
	}
*/
package update
