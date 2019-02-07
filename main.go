package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var cfg *config
var payload *deployerPayload

type deployerPayload struct {
	Project     string            `json:"project"`
	Registry    registry          `json:"registry"`
	ComposeFile string            `json:"composeFile"`
	Extra       map[string]string `json:"extra"`
}

type registry struct {
	URL      string `json:"url"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type deployerResponse struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type config struct {
	showPayload   bool
	insecure      bool
	dryRun        bool
	debug         bool
	deployerURL   string
	deployerToken string
	branch        string // Compose File Path
}

func main() {
	cfg = &config{}
	payload = &deployerPayload{}

	// initViper()
	checkConfig()
	setupPayload()
	readComposeFile()
	Deploy()

	fmt.Println("\n-> Deploy in progress... ")
}

func initViper() {
	// flag.StringP("name", "n", "", "Project name")
	// flag.StringP("branch", "b", "", "Branch name. Compose File Location")
	// flag.StringP("deployerUrl", "d", "", "Deployer URL")
	// flag.StringP("deployerToken", "t", "", "Deployer Token")
	// flag.StringP("registryUrl", "u", "", "Registry URL")
	// flag.StringP("registryLogin", "l", "", "Registry Login")
	// flag.StringP("registryPassword", "p", "", "Registry Password")
	// flag.StringP("registryEmail", "e", "", "Registry Email")
	// flag.StringP("extra", "x", "", "set extra vars")
	// flag.Bool("showPayload", false, "Secret Feature...")
	// flag.Bool("insecure", false, "HTTP insecure connection")
	// flag.Bool("debug", false, "Show debug info")
	// flag.Bool("dryRun", false, "Test helper without calling the deployer")
	// flag.Lookup("showPayload").Hidden = true
	// flag.Lookup("dryRun").Hidden = true
	// flag.Lookup("debug").Hidden = true

	// flag.Parse()

	// viper.BindPFlag("PROJECT_NAME", flag.Lookup("name"))
	// viper.BindPFlag("BRANCH_NAME", flag.Lookup("branch"))
	// viper.BindPFlag("DEPLOY_URL", flag.Lookup("deployerUrl"))
	// viper.BindPFlag("DEPLOY_TOKEN", flag.Lookup("deployerToken"))
	// viper.BindPFlag("REGISTRY_URL", flag.Lookup("registryUrl"))
	// viper.BindPFlag("REGISTRY_LOGIN", flag.Lookup("registryLogin"))
	// viper.BindPFlag("REGISTRY_PASSWORD", flag.Lookup("registryPassword"))
	// viper.BindPFlag("REGISTRY_EMAIL", flag.Lookup("registryEmail"))
	// viper.BindPFlag("EXTRA", flag.Lookup("extra"))
	// viper.BindPFlag("SHOW_PAYLOAD", flag.Lookup("showPayload"))
	// viper.BindPFlag("INSECURE", flag.Lookup("insecure"))
	// viper.BindPFlag("DRY_RUN", flag.Lookup("dryRun"))
	// viper.BindPFlag("DEBUG", flag.Lookup("debug"))

	// viper.AutomaticEnv()
}

func checkConfig() {
	branchName := os.Args[1]
	cfg.dryRun = EnvBool("DRY_RUN", false)

	argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]

	fmt.Printf("Args: %v \n", argsWithProg)
	fmt.Printf("Args: %v \n", argsWithoutProg)

	if cfg.dryRun {
		dryRunConfig()
	}

	cfg.deployerURL = EnvString("DEPLOY_SERVER", "")
	cfg.deployerToken = EnvString("DEPLOY_TOKEN", "")
	cfg.showPayload = EnvBool("SHOW_PAYLOAD", true)
	cfg.insecure = EnvBool("INSECURE", false)
	cfg.debug = EnvBool("DEBUG", false)
	cfg.branch = branchName

	if cfg.deployerURL == "" {
		fmt.Printf("Missing Deployer URL\n")
		os.Exit(-1)
	}
	if cfg.deployerToken == "" {
		fmt.Printf("Missing Deployer Token\n")
		os.Exit(-1)
	}

	fmt.Printf(">> Branch: %s | Token: %s \n", cfg.branch, cfg.deployerToken)
}

func dryRunConfig() {
	// viper.Set("PROJECT_NAME", "my-app")
	// viper.Set("DEPLOY_URL", "deployer.app.com")
	// viper.Set("REGISTRY_URL", "registry.test.com")
	// viper.Set("REGISTRY_LOGIN", "username")
	// viper.Set("REGISTRY_EMAIL", "contact@test.com")
	// viper.Set("DEPLOY_TOKEN", "xxxXXXxxx")
}

func setupPayload() {
	fmt.Printf(">> Extra: [%s] \n", EnvString("EXTRA", "noExtra"))
	payload.Project = EnvString("REPO_NAME", "")
	payload.Registry.URL = EnvString("REGISTRY_URL", "")
	payload.Registry.Login = EnvString("REGISTRY_LOGIN", "")
	payload.Registry.Password = EnvString("REGISTRY_PASSWORD", "")
	payload.Registry.Email = EnvString("REGISTRY_EMAIL", "")
	payload.Extra = extraStringToMap(EnvString("EXTRA", ""))

	payload.Extra["TAG"] = EnvString("TAG", "noTag")
	payload.Extra["COMMIT"] = EnvString("CI_COMMIT_SHA", "noCommit")
	payload.Extra["PROJECT"] = payload.Project
	payload.Extra["DATABASE_NAME"] = EnvString("DATABASE_NAME", "noDbName")
	payload.Extra["DATABASE_PASSWORD"] = EnvString("DATABASE_PASSWORD", "noDbPass")
	payload.Extra["REGISTRY_NAMESPACE"] = EnvString("REGISTRY_NAMESPACE", "noRegistryNamespace")

	if payload.Registry.URL == "" {
		fmt.Printf("Missing Registry URL\n")
		os.Exit(-1)
	}

	fmt.Printf(">> Payload: %v \n", payload.Extra)
}

func extraVarsToMap(extra string) map[string]string {
	m := make(map[string]string)

	arr := strings.Split(extra, ",")

	for _, val := range arr {
		m[val] = EnvString(val, "empty")
	}

	return m
}

func extraStringToMap(extra string) map[string]string {
	m := make(map[string]string)

	arr := strings.Split(extra, ",")

	for _, val := range arr {
		kv := strings.Split(val, ":")
		if len(kv) > 1 {
			m[kv[0]] = kv[1]
		}
	}

	return m
}

func readComposeFile() {
	composePath := fmt.Sprintf("deploy/%s/docker-compose.yml", cfg.branch)
	data, err := ioutil.ReadFile(composePath)
	if err != nil {
		fmt.Println("Error Read : Docker Compose File")
		if cfg.debug {
			fmt.Printf("Compose Path : %s \n", composePath)
			fmt.Println(err.Error())
		}
		os.Exit(-1)
	}
	fmt.Println("Succefuly read docker compose")
	payload.ComposeFile = base64.StdEncoding.EncodeToString(data)
}

// Deploy is a func
func Deploy() {
	r := callService()
	responseHandler(r)
}

func callService() *http.Response {
	protocol := "https"
	if cfg.insecure {
		protocol = "http"
	}
	serviceURL := fmt.Sprintf("%s://%s/%s/", protocol, cfg.deployerURL, "deploy")

	fmt.Printf("Calling Service : %v \n", serviceURL)

	if cfg.showPayload {
		fmt.Printf("With payload : %v \n", payload)
	}

	var URL *url.URL
	URL, err := url.Parse(serviceURL)
	if err != nil {
		fmt.Printf("Error Parsing URL : %v \n", serviceURL)
		os.Exit(-1)
	}

	bArray, err := json.Marshal(payload)
	req, err := http.NewRequest("POST", URL.String(), bytes.NewBuffer(bArray))
	if err != nil {
		fmt.Printf("Error Creating Request : %v \n", err.Error())
		os.Exit(-1)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("Auth-Token", cfg.deployerToken)
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error Calling Client : %v \n", err.Error())
		fmt.Println(req)
		os.Exit(-1)
	}

	return res
}

func responseHandler(r *http.Response) {
	res := &deployerResponse{}

	if r.StatusCode < 200 || r.StatusCode > 400 {
		fmt.Printf("Status %v: NOT OK\n", r.StatusCode)
		os.Exit(-1)
	}

	if err := json.NewDecoder(r.Body).Decode(res); err != nil {
		fmt.Printf("UNABLE to parse JSON response from service\n")
		os.Exit(-1)
	}

	fmt.Printf("\nResponse From Service : \n\n%s : %s \n", res.Type, res.ID)

	fmt.Printf("Logs are available at [https://%s/job/%s/log].", cfg.deployerURL, res.ID)
}

// EnvString ...
func EnvString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

// EnvBool ...
func EnvBool(env string, fallback bool) bool {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	p, _ := strconv.ParseBool(e)
	return p
}
