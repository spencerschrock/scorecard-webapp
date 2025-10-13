// postresult is a utility for developers to test the scorecard API by posting
// scorecard results to a local or remote instance of the service.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func post(payload []byte, repo, apiURL string) error {
	url := fmt.Sprintf("%s/projects/github.com/%s", apiURL, repo)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Execute request.
	log.Println("Sending POST to " + url)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing scorecard-api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}
		return fmt.Errorf("http response %d, status: %v, error: %v", resp.StatusCode, resp.Status, string(bodyBytes))
	}

	return nil
}
func main() {
	resultFile := flag.String("file", "", "Path to the results file to upload.")
	repo := flag.String("repo", "", "GitHub repository the results belong to, e.g. owner/repo .")
	tlogIndex := flag.Int64("tlog-index", 0, "Rekor tlog index corresponding to results file.")
	branch := flag.String("branch", "main", "Branch in the repository where the results were generated.")
	apiURL := flag.String("api-url", "http://127.0.0.1:8081", "URL for the scorecard service.")
	flag.Parse()

	if *resultFile == "" {
		log.Fatalln("--file must be set")
	}
	if *repo == "" {
		log.Fatalln("--repo must be set")
	}
	token, ok := os.LookupEnv("GITHUB_AUTH_TOKEN")
	if !ok {
		log.Fatalln("GITHUB_AUTH_TOKEN must be set")
	}

	jsonPayload, err := os.ReadFile(*resultFile)
	if err != nil {
		log.Fatalf("reading payload file: %v", err)
	}
	resultsPayload := struct {
		Result      string `json:"result"`
		Branch      string `json:"branch"`
		AccessToken string `json:"accessToken"`
		TlogIndex   int64  `json:"tlogIndex"`
	}{
		Result:      string(jsonPayload),
		Branch:      *branch,
		AccessToken: token,
		TlogIndex:   *tlogIndex,
	}

	payloadBytes, err := json.Marshal(resultsPayload)
	if err != nil {
		log.Fatalln(err)
	}
	if err := post(payloadBytes, *repo, *apiURL); err != nil {
		log.Fatalln(err)
	}
}
