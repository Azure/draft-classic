package draft

// based on https://github.com/weaveworks/flux/blob/1.0.1/registry/gcp.go

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

const (
	gcpDefaultTokenURL = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
)

type gceToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func getGCPOauthToken() (RegistryAuth, error) {
	var regauth RegistryAuth

	request, err := http.NewRequest("GET", gcpDefaultTokenURL, nil)
	if err != nil {
		return regauth, fmt.Errorf("could not make new request to GCP metadata service: %v", err)
	}

	request.Header.Add("Metadata-Flavor", "Google")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return regauth, fmt.Errorf("could not obtain GCP metadata: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		return regauth, fmt.Errorf("unexpected status from GCP metadata service: %s", response.Status)
	}

	var token gceToken
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&token); err != nil {
		return regauth, fmt.Errorf("could not json decode GCP metadata: %v", err)
	}

	if err := response.Body.Close(); err != nil {
		return regauth, fmt.Errorf("unexpected error while closing GCP metadata request: %v", err)
	}

	log.Info("obtained GCR registry auth via GCP service account")
	regauth.Username = "oauth2accesstoken"
	regauth.Password = token.AccessToken
	return regauth, nil
}
