package main

import (
	"flag"
	"fmt"
)

// Read the config with budgets mapped to currency & target budget
// rubbles -> eur, euro
// Get transactions from ynab's secondary budget
// Get today's currency rate and store it into a cache (async)
//

type Sync struct {
	Name   string
	Source budget
	Target budget
}

type Budget struct {
	Name     string
	Id       string
	Currency string
}

type Config struct {
	Budgets []budget
	Syncs   []sync
}

func parseArgs() {
	configArg := flag.String("config", "config.yaml", "yaml formatted config file's path")
	flag.Parse()
	fmt.Println(*configArg, flag.Args())
}

func main() {
	// yanb-mcw --config config.yaml
	parseArgs()
	fmt.Println("Done")
}
