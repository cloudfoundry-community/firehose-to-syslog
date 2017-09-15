package main

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/uaago"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args[1:]) != 3 {
		fmt.Fprintf(os.Stderr, "Usage %s [URL] [USERNAME] [PASS]", args[0])
		return 1
	}

	uaa, err := uaago.NewClient(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %s", err.Error())
		return 1
	}

	token, err := uaa.GetAuthToken(args[2], args[3], false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Faild to get auth token: %s", err.Error())
		return 1
	}

	fmt.Fprintf(os.Stdout, "TOKEN: %s\n", token)
	return 0
}
