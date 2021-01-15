package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/tmnhat2001/worker-service/internal/worker"
)

func TestStartJob(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	command := "echo \"hello world\""
	response, err := executeStartJobRequest(command, "user1", "thisispasswordforuser1")
	if err != nil {
		t.Error(err)
		return
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", response.StatusCode)
	}

	job, err := getJobFromResponse(response)
	if err != nil {
		t.Error(err)
		return
	}

	if job.ID == "" {
		t.Error("The job ID should not be empty")
	}

	if job.Status != worker.Running {
		t.Errorf("Expected the job status to be '%s', but got '%s'", worker.Running, job.Status)
	}

	if job.Command != command {
		t.Errorf(`The input command is different from the one in the response.
Expected: %s
Got: %s`, command, job.Command)
	}

	if job.ExitCode != "" {
		t.Errorf("Expected the job exit code to be empty, but got %s", job.ExitCode)
	}
}

func TestStopJob(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	username := "user1"
	password := "thisispasswordforuser1"

	command := "sleep 5"
	startResponse, err := executeStartJobRequest(command, username, password)
	if err != nil {
		t.Error(err)
		return
	}

	job1, err := getJobFromResponse(startResponse)
	if err != nil {
		t.Error(err)
		return
	}

	stopResponse, err := executeStopJobRequest(job1.ID, username, password)
	if err != nil {
		t.Error(err)
		return
	}

	if stopResponse.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", stopResponse.StatusCode)
	}

	job2, err := getJobFromResponse(stopResponse)
	if err != nil {
		t.Error(err)
	}

	if job2.Status != worker.Stopped {
		t.Errorf("Expected the job status to be '%s', but got '%s'", worker.Stopped, job2.Status)
	}

	if job2.ExitCode != "-1" {
		t.Errorf("Expected the job exit code to be -1, but got %s", job2.ExitCode)
	}
}

func TestGetJob(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	username := "user1"
	password := "thisispasswordforuser1"

	command := "echo hello world"
	startResponse, err := executeStartJobRequest(command, username, password)
	if err != nil {
		t.Error(err)
		return
	}

	job1, err := getJobFromResponse(startResponse)
	if err != nil {
		t.Error(err)
		return
	}

	response, err := executeGetJobRequest(job1.ID, username, password)
	if err != nil {
		t.Error(err)
		return
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, but got %d", response.StatusCode)
	}

	job2, err := getJobFromResponse(response)
	if err != nil {
		t.Error(err)
		return
	}

	if job2.Status != worker.Completed {
		t.Errorf("Expected the job status to be '%s', but got '%s'", worker.Completed, job2.Status)
	}

	if job2.Stdout != "hello world\n" {
		t.Errorf(`The result for Stdout is not correct.
Expected: %s
Got: %s`, "hello world\n", job2.Stdout)
	}

	if job2.ExitCode != "0" {
		t.Errorf("Expected the job exit code to be 0, but got %s", job2.ExitCode)
	}
}

func TestPlainHTTP(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	response, err := executePlainTextRequest()
	if err != nil {
		t.Error(err)
		return
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code to be 400. Got: %d", response.StatusCode)
	}
}

func TestAuthentication(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	response, err := executeStartJobRequest("echo hello world", "user1", "anIncorrectPassword")
	if err != nil {
		t.Error(err)
		return
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code 401, but got %d", response.StatusCode)
	}

	expectErrorMessage(response, "Unable to authenticate user", t)
}

func TestAuthorization(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	startResponse, err := executeStartJobRequest("echo hello world", "user1", "thisispasswordforuser1")
	if err != nil {
		t.Error(err)
		return
	}

	job1, err := getJobFromResponse(startResponse)
	if err != nil {
		t.Error(err)
		return
	}

	response, err := executeGetJobRequest(job1.ID, "user2", "thisispasswordforuser2")
	if err != nil {
		t.Error(err)
		return
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code 404, but got %d", response.StatusCode)
	}

	expectErrorMessage(response, "Failed to find job", t)
}

func TestInvalidCommand(t *testing.T) {
	server, err := NewServer(8989)
	if err != nil {
		t.Error(err)
		return
	}

	go server.run("../../certs/server.crt", "../../certs/server.key")
	defer server.close()

	command := "an invalid command"
	response, err := executeStartJobRequest(command, "user1", "thisispasswordforuser1")
	if err != nil {
		t.Error(err)
		return
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code 500, but got %d", response.StatusCode)
	}

	expectErrorMessage(response, "Failed to start job", t)
}

func executeStartJobRequest(command, username, password string) (*http.Response, error) {
	requestBody, err := json.Marshal(map[string]string{
		"Command": command,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", makeURL("https", 8989, "start"), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	return executeRequest(request, username, password)
}

func executeStopJobRequest(jobID, username, password string) (*http.Response, error) {
	requestBody, err := json.Marshal(map[string]string{
		"ID": jobID,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("PUT", makeURL("https", 8989, "stop"), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	return executeRequest(request, username, password)
}

func executeGetJobRequest(jobID, username, password string) (*http.Response, error) {
	path := "/jobs/" + jobID

	request, err := http.NewRequest("GET", makeURL("https", 8989, path), nil)
	if err != nil {
		return nil, err
	}

	return executeRequest(request, username, password)
}

func executePlainTextRequest() (*http.Response, error) {
	path := "/jobs/123"

	request, err := http.NewRequest("GET", makeURL("http", 8989, path), nil)
	if err != nil {
		return nil, err
	}

	return executeRequest(request, "user1", "thisispasswordforuser1")
}

func newHTTPSClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &http.Client{Transport: tr}
}

func makeURL(protocol string, port int, path string) string {
	return fmt.Sprintf("%s://localhost:%d/%s", protocol, port, path)
}

func executeRequest(request *http.Request, username, password string) (*http.Response, error) {
	request.SetBasicAuth(username, password)

	client := newHTTPSClient()
	response, err := client.Do(request)

	return response, err
}

func getJobFromResponse(response *http.Response) (*worker.Job, error) {
	responseMap, err := parseResponse(response)
	if err != nil {
		return nil, err
	}

	var job worker.Job
	json.Unmarshal(responseMap, &job)

	return &job, nil
}

func parseResponse(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errorFromResponse(body)
	}

	return body, nil
}

func errorFromResponse(content []byte) error {
	var contentMap map[string]string
	json.Unmarshal(content, &contentMap)

	message := fmt.Sprintf("Error: %s", contentMap["error"])
	return errors.New(message)
}

func expectErrorMessage(response *http.Response, expectedMessage string, t *testing.T) {
	message, err := parseErrorResponse(response)
	if err != nil {
		t.Error(err)
	}

	if message != expectedMessage {
		t.Errorf("Expected error message: %s\nGot: %s", expectedMessage, message)
	}
}

func parseErrorResponse(response *http.Response) (string, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var contentMap map[string]string
	json.Unmarshal(body, &contentMap)

	return contentMap["error"], nil
}
