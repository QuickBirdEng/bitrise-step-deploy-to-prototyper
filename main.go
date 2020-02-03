package main

import (
	"bytes"
	"io"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"net/http"
	"os"
	"os/exec"
)

// Creates a new file upload http request with optional extra params
// Source: https://matt.aimonetti.net/posts/2013-07-golang-multipart-file-upload-example/
func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func main() {

	fmt.Println("The following Inputs where specified:")
	upload_url := os.Getenv("upload_url")
	fmt.Println("'upload_url':", upload_url)
	build_key := os.Getenv("build_key")
	fmt.Println("'build_key':", build_key)
	git_branch_name := os.Getenv("git_branch_name")
	fmt.Println("'git_branch_name:", git_branch_name)
	git_commit_hash := os.Getenv("git_commit_hash")
	fmt.Println("'git_commit_hash:", git_commit_hash)
	project_type := os.Getenv("project_type")
	fmt.Println("'project_type':", project_type)
	meta := os.Getenv("meta")
	fmt.Println("'meta':", meta)
	ipa_path := os.Getenv("ipa_path")
	fmt.Println("'ipa_path':", ipa_path)
	upload_key := os.Getenv("upload_key")
	fmt.Println("'upload_key':", upload_key)

	if len(git_branch_name) > 0 && len(git_commit_hash) > 0 {
		fmt.Println("Creating build key out of git_branch_name & git_commit_hash")
		build_key = git_branch_name + "-" + git_commit_hash
		fmt.Println("Using created build_key: ", build_key)
	}

	fmt.Println("Creating Request ...")

	extraParams := map[string]string{
		"beta[type]":			project_type,
		"beta[build_key]":		build_key,
		"beta[meta]":			meta,
		"upload_key":			upload_key,
	}

	request, err := newfileUploadRequest(upload_url, extraParams, "beta[build]", ipa_path)
	if err != nil {
		fmt.Printf("Failed to create Upload Request: ", err)
		os.Exit(1)
	} else {
		fmt.Printf("Created Request to %s\n", request.URL)
	}

	fmt.Println("Send Request ...")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("Request failed: %s", err)
		os.Exit(1)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
    	if err != nil {
			fmt.Printf("Request failed: %s", err)
			os.Exit(1)
		}
		resp.Body.Close()
		
		fmt.Println("Request completed successfully!")
		fmt.Println("Status Code: ", resp.StatusCode)
		fmt.Println("Body:", body)
	}

	// --- Step Outputs: Export Environment Variables for other Steps:
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "PROTOTYPER_BUILD_KEY", "--value", build_key).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		os.Exit(1)
	}

	os.Exit(0)
}
