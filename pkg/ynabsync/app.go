package ynabsync

import (
	"flag"
	"fmt"
	"os"

	ynab "github.com/brunomvsouza/ynab.go"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
	yaml "gopkg.in/yaml.v3"
)

var YNAB_TOKEN string = os.Getenv("YNAB_TOKEN")

type BudgetMetadata struct {
	Name     string
	Id       string
	Currency string
}

type Sync struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type Config struct {
	// Budgets map[string]Budget `yaml:"budgets"`
	Syncs map[string]Sync `yaml:"syncs"`
}

func parseArgs() map[string]string {
	config := flag.String("config", "config.yaml", "yaml formatted config file's path")
	flag.Parse()

	flagValues := make(map[string]string)

	flagValues["config"] = *config

	return flagValues
}

func initLogger(d bool, logFileName string) *zap.Logger {
	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic("Log file error")
	}
	pe := zap.NewProductionEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(pe)

	pe.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(pe)

	level := zap.InfoLevel
	if d {
		level = zap.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(f), level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	)

	l := zap.New(core)

	return l
}

func parseConfig(configPath string) *Config {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("ReadFile")
	}
	var config Config
	if err := yaml.Unmarshal(configFile, &config); err != nil {
		fmt.Println(err)
	}
	if !configValidity(config) {
		panic("Config validation process ended with errors.")
	}

	return &config
}

func configValidity(config Config) bool {
	// var issues bool = false
	// Check if all elements from source and target exist in budgets
	// for sync_name, sync := range config.Syncs {
	// 	if _, exists := config.Budgets[sync.Source]; !exists {
	// 		fmt.Printf("Source %s.%s does not exist in budgets\n", sync_name, sync.Source)
	// 		issues = true
	// 	}

	// 	if _, exists := config.Budgets[sync.Target]; !exists {
	// 		fmt.Printf("Target %s.%s does not exist in budgets\n", sync_name, sync.Target)
	// 		issues = true
	// 	}
	// }
	// return !issues
	return true
}

func Main() {
	var log = initLogger(true, "general.log")
	log.Info("Logger has been initialized")
	config := parseArgs()
	_ = parseConfig(config["config"])

	c := ynab.NewClient(YNAB_TOKEN)
	budgets, err := c.Budget().GetBudgets()
	if err != nil {
		panic(err)
	}
	// var budgetMetadata []BudgetMetadata = make([]BudgetMetadata, 0, len(budgets))
	for _, b := range budgets {
		// log.Info(fmt.Sprintf("Requested budget %s %s %s", b.Name, b.CurrencyFormat.ISOCode, b.ID))
		log.Info(
			"Requested budget",
			zap.String("Name", b.Name),
			zap.String("ISOCode", b.CurrencyFormat.ISOCode),
			zap.String("Id", b.ID),
		)

	}
	log.Info(fmt.Sprintf("RateLimit Used/Total %d/%d", c.RateLimit().Used(), c.RateLimit().Total()))

}
