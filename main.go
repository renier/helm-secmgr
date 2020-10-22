package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/httpclient"
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
	instanceID string
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

	endpoint = os.Getenv("KEYPROTECT_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://us-south.kms.cloud.ibm.com"
	}

	token = os.Getenv("IAM_TOKEN")
	if token == "" {
		log.Fatalln("no IAM_TOKEN found in env")
	}

	instanceID = os.Getenv("SERVICE_INSTANCE_ID")
	if instanceID == "" {
		log.Fatalln("no SERVICE_INSTANCE_ID found in env")
	}

	// Substitute the necessary values in it.
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln("reading input stream:", err)
	}
	tmpl, err := template.New("").Delims("[[", "]]").Funcs(template.FuncMap{
		"unwrap": unwrap,
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

type unwrapResponse struct {
	Plaintext string
}

func unwrap(keyID, ciphertext string) string {
	url := endpoint + "/api/v2/keys"
	body := `{"ciphertext":"` + ciphertext + `"}`
	tokenPrefix := "bearer "
	if strings.ToLower(token[:7]) == tokenPrefix {
		tokenPrefix = ""
	}

	resp, err := client.Post(url+"/"+keyID+"?action=unwrap", strings.NewReader(body), http.Header{
		"authorization": []string{tokenPrefix + token},
		"bluemix-instance": []string{instanceID},
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

	var response unwrapResponse
	err = json.Unmarshal(b, &response)
	if err != nil {
		log.Fatalln("unmarshalling:", err, string(b))
	}

	decodedPlaintext, err := b64.StdEncoding.DecodeString(response.Plaintext)
	if err != nil {
		log.Fatalln("base64 decoding:", err)
	}

	return string(decodedPlaintext)
}