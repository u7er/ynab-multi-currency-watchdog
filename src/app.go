package main

import (
	"flag"
	"fmt"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
	yaml "gopkg.in/yaml.v3"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Budget struct {
	Name     string `yaml:name`
	Id       string `yaml:"id"`
	Currency string `yaml:"currency"`
}

type Sync struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

type Config struct {
	Budgets map[string]Budget `yaml:"budgets"`
	Syncs   map[string]Sync   `yaml:"syncs"`
}

func parseArgs() map[string]string {
	config := flag.String("config", "config.yaml", "yaml formatted config file's path")
	flag.Parse()

	flagValues := make(map[string]string)

	flagValues["config"] = *config

	return flagValues
}

func initLogger(d bool, logFileName string) *zap.SugaredLogger {
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

	return l.Sugar()
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
	var issues bool = false
	// Check if all elements from source and target exist in budgets
	for sync_name, sync := range config.Syncs {
		if _, exists := config.Budgets[sync.Source]; !exists {
			fmt.Printf("Source %s.%s does not exist in budgets\n", sync_name, sync.Source)
			issues = true
		}

		if _, exists := config.Budgets[sync.Target]; !exists {
			fmt.Printf("Target %s.%s does not exist in budgets\n", sync_name, sync.Target)
			issues = true
		}
	}
	return !issues
}

func getBudgetById(budget_id string) {
	var token string = os.Getenv("YNAB_TOKEN")
	// Specify the URL you want to make a request to
	var url string = "https://api.ynab.com/v1/budgets"

	// Make a GET request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating GET request: %v\n", err)
		return
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	client := http.DefaultClient
	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error making GET request: %v\n", err)
		return
	}

	defer response.Body.Close()

	// Check if the response status code is OK (200)
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Error: Unexpected status code %d\n", response.StatusCode)
		return
	}

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	// Print the response body as a string
	fmt.Println(string(body))
}

// func main() {
// 	// yanb-mcw --config config.yaml

// 	var log = initLogger(true, "general.log")
// 	log.Info("test")
// 	config := parseArgs()
// 	configPtr := parseConfig(config["config"])
// 	getBudgetById("123")
// 	fmt.Println("Done", *configPtr)
// }

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("worker", id, "started  job", j)
		time.Sleep(time.Second * 2)
		fmt.Println("worker", id, "finished job", j)
		results <- j * 100
	}
}

func main() {
	const numJobs = 11
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= numJobs/3; w++ {
		go worker(w, jobs, results)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}
