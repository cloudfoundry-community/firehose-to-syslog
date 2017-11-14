package main

import "os"

func main() {
	cli := &CLI{}
	os.Exit(cli.Run(os.Args))
}
