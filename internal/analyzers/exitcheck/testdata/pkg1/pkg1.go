//lint:file-ignore test data
package main

import "os"

func main() {
	os.Exit(0) // want "direct os.Exit call in main func"
}

func dummy() {
	os.Exit(0)
}
