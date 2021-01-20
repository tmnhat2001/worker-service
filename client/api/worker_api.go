package api

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// WorkerAPI provides a client-side implementation to call the Worker API
type WorkerAPI struct {
	config WorkerAPIConfig
	client *http.Client
}

const (
	requestTimeout = 10 * time.Second
	endpoint       = "https://localhost:8080"
)

// NewWorkerAPI creates a new WorkerAPI from the config struct
func NewWorkerAPI(config WorkerAPIConfig) (*WorkerAPI, error) {
	client, err := newHTTPClient(config.CertFilePath)
	if err != nil {
		return nil, err
	}

	return &WorkerAPI{
		config: config,
		client: client,
	}, nil
}

func newHTTPClient(certFilePath string) (*http.Client, error) {
	cert, err := parseCertificate(certFilePath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: certPool},
		},
		Timeout: requestTimeout,
	}, nil
}

func parseCertificate(certFilePath string) (*x509.Certificate, error) {
	certBytes, err := ioutil.ReadFile(certFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read server's certificate file")
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return nil, errors.New("no PEM data found in server certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse server's certificate")
	}

	return cert, nil
}

func errorFromResponse(response *http.Response) error {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errorMessage := fmt.Sprintf("error reading response body. Response status: %s", response.Status)
		return errors.Wrap(err, errorMessage)
	}

	var contentMap map[string]string
	err = json.Unmarshal(body, &contentMap)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("error: %s", contentMap["error"])
	return errors.New(message)
}

// StartJob calls the /start endpoint of the Worker API
func (api *WorkerAPI) StartJob(command string) ([]byte, error) {
	requestBody, err := json.Marshal(map[string]string{
		"Command": command,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request body")
	}

	url := endpoint + "/start"
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request")
	}

	return api.executeRequest(request)
}

// StopJob calls the /stop endpoint of the Worker API
func (api *WorkerAPI) StopJob(jobID string) ([]byte, error) {
	requestBody, err := json.Marshal(map[string]string{
		"ID": jobID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request body")
	}

	url := endpoint + "/stop"
	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request")
	}

	return api.executeRequest(request)
}

// GetJob calls the /jobs endpoint of the Worker API
func (api *WorkerAPI) GetJob(jobID string) ([]byte, error) {
	url := endpoint + "/jobs/" + jobID
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create request")
	}

	return api.executeRequest(request)
}

func (api *WorkerAPI) executeRequest(request *http.Request) ([]byte, error) {
	request.SetBasicAuth(api.config.Username, api.config.Password)

	response, err := api.client.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "Error sending request")
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errorFromResponse(response)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading response body")
	}

	return body, nil
}

// WorkerAPIConfig provides configurations to set up a WorkerAPI
type WorkerAPIConfig struct {
	Username     string
	Password     string
	CertFilePath string
}
