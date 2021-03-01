package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/mitchellh/go-homedir"
)

const (
	timeout               = 10 * time.Second
	retries               = 3
	initialTimeout        = 250 * time.Millisecond
	maxTimeout            = 2 * time.Second
	exponentFactor        = 2
	maximumJitterInterval = 10 * time.Millisecond
)

var (
	client *httpclient.Client

	// KeyProtect
	endpoint string
	token string
)

func main() {
	// Setup http client
	backoff := heimdall.NewExponentialBackoff(
		initialTimeout, maxTimeout, exponentFactor, maximumJitterInterval)
	retrier := heimdall.NewRetrier(backoff)

	client = httpclient.NewClient(
		httpclient.WithHTTPTimeout(timeout),
		httpclient.WithRetrier(retrier),
		httpclient.WithRetryCount(retries),
	)

	endpoint = os.Getenv("VAULT_ADDR")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8200"
	}

	token = os.Getenv("VAULT_TOKEN")
	if token == "" {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatalln("getting home:", err)
		}

		b, err := ioutil.ReadFile(filepath.Join(home, "/.vault-token"))
		if err != nil {
			log.Fatalln("reading vault token:", err)
		}

		token = string(b)
	}


	// Substitute the necessary values in it.
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("reading input stream:", err)
	}
	tmpl, err := template.New("").Delims("<<", ">>").Funcs(template.FuncMap{
		"secret_ref": fetch,
	}).Parse(string(input))
	if err != nil {
		log.Fatalln("parse yml:", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		log.Fatalln("tmpl execute:", err)
	}
	fmt.Print(buf.String())
}

func fetch(path string) string {
	url := endpoint + "/v1/ibmcloud/arbitrary/secrets/" + path

	resp, err := client.Get(url, http.Header{
		"x-vault-token": []string{token},
	})
	if err != nil {
		log.Fatalln(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("bad status code:", resp.StatusCode, string(b))
	}

	var response map[string]interface{}
	err = json.Unmarshal(b, &response)
	if err != nil {
		log.Fatalln("unmarshalling:", err, string(b))
	}

	responsePath := "data.secret_data.payload"
	val, ok := Grab(response, responsePath)
	if !ok {
		log.Fatalf("did not find path %s in the response", responsePath)
	}

	result, ok := val.(string)
	if !ok {
		log.Fatalln("bad result:", val)
	}

	return result
}

func Grab(s map[string]interface{}, path string) (interface{}, bool) {
	dotIndex := strings.Index(path, ".")
	last := false
	if dotIndex == -1 {
		dotIndex = len(path)
		last = true
	}

	fieldName := path[0:dotIndex]
	val, ok := s[fieldName]
	if !ok || val == nil {
		return nil, false
	}

	if last {
		return val, true
	}

	result, ok := val.(map[string]interface{})
	if !ok {
		return result, ok
	}

	return Grab(result, path[dotIndex+1:])
}
