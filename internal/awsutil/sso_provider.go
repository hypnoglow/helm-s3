package awsutil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sso"
	"gopkg.in/ini.v1"
)

// Profile represents an AWS SSO profile section, as expected in the AWS config file
type Profile struct {
	Name         string
	SSOAccountID string `ini:"sso_account_id"`
	SSORegion    string `ini:"sso_region"`
	SSOStartURL  string `ini:"sso_start_url"`
	SSORoleName  string `ini:"sso_role_name"`
}

// SSOCache represents an AWS SSO cache file
type SSOCache struct {
	StartURL    string `json:"startUrl"`
	Region      string `json:"region"`
	AccessToken string `json:"accessToken"`
	ExpiresAt   string `json:"expiresAt"`
}

// SSOCredentialProvider is a custom AWS credential provider to retrieve credentials which can be used with the SDK from SSO tokens
type SSOCredentialProvider struct {
	Session   *session.Session
	AccountID string
	RoleName  string
	Cache     *SSOCache
}

var awsCachePath string
var awsConfigPath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	awsCachePath = filepath.Join(home, ".aws/sso/cache")
	awsConfigPath = filepath.Join(home, ".aws/config")
}

// NewSSOCredentialProvider returns an initialised SSOCredentialProvider
func NewSSOCredentialProvider(awsProfile string) (*SSOCredentialProvider, error) {
	if awsProfile == "" {
		awsProfile = os.Getenv("AWS_PROFILE")
	}
	profile, err := getProfile(awsProfile)
	if err != nil {
		return nil, err
	}
	provider := &SSOCredentialProvider{
		Session: session.Must(session.NewSession(&aws.Config{
			Region: &profile.SSORegion,
		})),
		AccountID: profile.SSOAccountID,
		RoleName:  profile.SSORoleName,
	}
	provider.Cache, err = getCache(awsCachePath)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// Retrieve retrieves a set of temporary credentials using the access token from the cache
func (p *SSOCredentialProvider) Retrieve() (credentials.Value, error) {
	in := &sso.GetRoleCredentialsInput{
		AccountId:   &p.AccountID,
		RoleName:    &p.RoleName,
		AccessToken: &p.Cache.AccessToken,
	}
	svc := sso.New(p.Session)
	out, err := svc.GetRoleCredentials(in)
	if err != nil {
		return credentials.Value{}, err
	}
	return credentials.Value{
		ProviderName:    "SSOCredentialProvider",
		AccessKeyID:     *out.RoleCredentials.AccessKeyId,
		SecretAccessKey: *out.RoleCredentials.SecretAccessKey,
		SessionToken:    *out.RoleCredentials.SessionToken,
	}, nil
}

// IsExpired determines if the cached SSO token is expired
func (p *SSOCredentialProvider) IsExpired() bool {
	t, err := time.Parse("2006-01-02T15:04:05UTC", p.Cache.ExpiresAt)
	if err != nil {
		log.Fatal(err)
	}
	return t.Before(time.Now())
}

func getCacheFile(path string) (*SSOCache, error) {
	cache := &SSOCache{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, cache)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func getCache(cacheDir string) (*SSOCache, error) {

	cache := &SSOCache{}

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		// handle failure accessing a path
		if err != nil {
			return err
		}
		// skip directories (excluding the cache dir itself)
		if info.IsDir() && path != cacheDir {
			return filepath.SkipDir
		}
		// skip anything that's not a json file
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		// skip the botocore files
		if strings.HasPrefix(filepath.Base(path), "botocore-") {
			return nil
		}
		// get the cache details from file
		cache, err = getCacheFile(path)
		if err != nil {
			return err
		}
		return io.EOF
	})

	if err != nil && err != io.EOF {
		return nil, err
	}

	return cache, nil

}

func getProfile(name string) (*Profile, error) {
	if name == "" {
		name = "default"
	}
	cfg, err := ini.Load(awsConfigPath)
	if err != nil {
		return nil, err
	}
	var section string
	if name == "default" {
		section = "default"
	} else {
		section = fmt.Sprintf("profile %s", name)
	}
	s, err := cfg.GetSection(section)
	if err != nil {
		return nil, err
	}
	return &Profile{
		Name:         name,
		SSOAccountID: s.Key("sso_account_id").String(),
		SSORegion:    s.Key("sso_region").String(),
		SSOStartURL:  s.Key("sso_start_url").String(),
		SSORoleName:  s.Key("sso_role_name").String(),
	}, nil
}
