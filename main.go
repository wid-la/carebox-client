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
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	deployerURL   string
	deployerToken string
	branch        string // Compose File Path
}

func main() {
	cfg = &config{}
	payload = &deployerPayload{}

	initViper()
	checkConfig()
	setupPayload()
	readComposeFile()
	Deploy()

	fmt.Println("\n-> Deploy in progres... ")
}

func initViper() {
	flag.StringP("project", "p", "", "Project name")
	flag.StringP("branch", "b", "", "Branch name. Compose File Location")
	flag.StringP("deployerUrl", "u", "", "Deployer URL")
	flag.StringP("deployerToken", "t", "DEPLOY_TOKEN", "Deployer Token Variable Name")
	flag.StringP("extra", "e", "", "set extra vars")
	flag.Bool("showPayload", false, "Secret Feature...")
	flag.Bool("insecure", false, "HTTP insecure connection")
	flag.Bool("dryRun", false, "Test helper without calling the deployer")
	flag.Lookup("showPayload").Hidden = true

	flag.Parse()

	viper.BindPFlag("PROJECT_NAME", flag.Lookup("project"))
	viper.BindPFlag("BRANCH_NAME", flag.Lookup("branch"))
	viper.BindPFlag("DEPLOY_URL", flag.Lookup("deployerUrl"))
	viper.BindPFlag("DEPLOY_TOKEN_VAR", flag.Lookup("deployerToken"))
	viper.BindPFlag("EXTRA", flag.Lookup("extra"))
	viper.BindPFlag("SHOW_PAYLOAD", flag.Lookup("showPayload"))
	viper.BindPFlag("INSECURE", flag.Lookup("insecure"))
	viper.BindPFlag("DRY_RUN", flag.Lookup("dryRun"))

	viper.AutomaticEnv()
}

func checkConfig() {
	cfg.dryRun = viper.GetBool("DRY_RUN")

	if cfg.dryRun {
		dryRunConfig()
	}

	tokenVar := viper.GetString("DEPLOY_TOKEN_VAR")

	cfg.deployerURL = viper.GetString("DEPLOY_URL")
	cfg.deployerToken = viper.GetString(tokenVar)
	cfg.showPayload = viper.GetBool("SHOW_PAYLOAD")
	cfg.insecure = viper.GetBool("INSECURE")
	cfg.branch = viper.GetString("BRANCH_NAME")

	if cfg.deployerURL == "" {
		fmt.Printf("Missing Deployer URL\n")
		os.Exit(-1)
	}
	if cfg.deployerToken == "" {
		fmt.Printf("Missing Deployer Token\n")
		os.Exit(-1)
	}
}

func dryRunConfig() {
	viper.Set("PROJECT_NAME", "my-app")
	viper.Set("DEPLOY_URL", "deployer.app.com")
	viper.Set("REGISTRY_URL", "registry.test.com")
	viper.Set("REGISTRY_LOGIN", "username")
	viper.Set("REGISTRY_EMAIL", "contact@test.com")
	viper.Set("DEPLOY_TOKEN", "xxxXXXxxx")
}

func setupPayload() {
	fmt.Println(">> Extra ", viper.GetString("EXTRA"))
	payload.Project = viper.GetString("PROJECT_NAME")
	payload.Registry.URL = viper.GetString("REGISTRY_URL")
	payload.Registry.Login = viper.GetString("REGISTRY_LOGIN")
	payload.Registry.Password = viper.GetString("REGISTRY_PASSWORD")
	payload.Registry.Email = viper.GetString("REGISTRY_EMAIL")
	payload.Extra = extraStringToMap(viper.GetString("EXTRA"))

	if payload.Registry.URL == "" {
		fmt.Printf("Missing Registry URL\n")
		os.Exit(-1)
	}
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
}
