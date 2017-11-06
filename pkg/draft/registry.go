package draft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/docker/docker/api/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	log "github.com/Sirupsen/logrus"
)

const (
	// name of the docker pull secret draftd will create in the desired destination namespace
	pullSecretName = "draftd-pullsecret"
	// name of the default service account draftd will modify with the imagepullsecret
	svcAcctNameDefault = "default"
)

type (
	// RegistryConfig specifies configuration for the image repository.
	RegistryConfig struct {
		// Auth is the authorization token used to push images up to the registry.
		Auth string
		// URL is the URL of the registry (e.g. quay.io/myuser, docker.io/myuser, myregistry.azurecr.io)
		URL string
	}

	// RegistryAuth is the registry authentication credentials
	RegistryAuth struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		Email         string `json:"email"`
		RegistryToken string `json:"registrytoken"`
	}

	// DockerAuth is a container for the registry authentication credentials wrapped
	// by the registry server name.
	DockerAuth map[string]RegistryAuth
)

func getRegistryAuthToken(cfg *RegistryConfig) (RegistryAuth, error) {
	var regauth RegistryAuth

	log.Printf("no registry auth tokens provided, will attempt to auto-detect (url: %q)", cfg.URL)
	registryURL, err := url.Parse(cfg.URL)
	if err != nil {
		return regauth, fmt.Errorf("could not parse registry url: %v", err)
	}
	if registryURL.Host == "" && registryURL.Path == "" {
		return regauth, fmt.Errorf("empty registry url")
	}
	if registryURL.Host == "" { // If there's no https:// prefix, it won't parse the host.
		registryURL, err = url.Parse(fmt.Sprintf("https://%s/", cfg.URL))
		if err != nil {
			return regauth, fmt.Errorf("could not parse registry url after adding protocol schema prefix: %v", err)
		}
	}

	log.Printf("parsed URL: %#v", registryURL)
	switch registryURL.Host {
	case "gcr.io":
		log.Println("will attempt to auto-detect gcr auth credentials")
		return getGCPOauthToken()
	default:
		return regauth, fmt.Errorf("could not auto-detect registry token for %q", registryURL.Host)
	}

	return regauth, nil
}

func configureRegistryAuth(cfg *RegistryConfig) (RegistryAuth, error) {
	var regauth RegistryAuth

	if cfg.Auth == "" {
		return getRegistryAuthToken(cfg)
	}

	// base64 decode the registryauth string.
	b64dec, err := base64.StdEncoding.DecodeString(cfg.Auth)
	if err != nil {
		return regauth, fmt.Errorf("could not base64 decode registry authentication string: %v", err)
	}

	// break up registry auth json string into a RegistryAuth object.
	if err := json.Unmarshal(b64dec, &regauth); err != nil {
		return regauth, fmt.Errorf("could not json decode registry authentication string: %v", err)
	}

	if regauth.Username == "" && regauth.Password == "" {
		return getRegistryAuthToken(cfg)
	}

	return regauth, nil
}

func getRegistryAuthAsPushOptions(cfg *RegistryConfig) (types.ImagePushOptions, error) {
	var pushopts types.ImagePushOptions

	regauth, err := configureRegistryAuth(cfg)
	if err != nil {
		return pushopts, err
	}
	jsbytes, err := json.Marshal(regauth)
	if err != nil {
		return pushopts, fmt.Errorf("could not json encode docker authentication string: %v", err)
	}

	pushopts = types.ImagePushOptions{RegistryAuth: base64.URLEncoding.EncodeToString(jsbytes)}
	log.Printf("pushopts=%#v", pushopts)

	return pushopts, nil
}

func getRegistryAuthAsPullSecret(cfg *RegistryConfig, meta metav1.ObjectMeta) (*v1.Secret, error) {
	var secret *v1.Secret

	regauth, err := configureRegistryAuth(cfg)
	if err != nil {
		return secret, err
	}
	js, err := json.Marshal(DockerAuth{cfg.URL: regauth})
	if err != nil {
		return secret, fmt.Errorf("could not json encode docker authentication string: %v", err)
	}

	secret = &v1.Secret{
		ObjectMeta: meta,
		Type:       v1.SecretTypeDockercfg,
		StringData: map[string]string{
			".dockercfg": string(js),
		},
	}
	log.Printf("secret=%#v", secret)

	return secret, nil
}
