// Package credshandler handles the /creds endpoint
package credshandler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type creds struct {
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	Token           string    `json:"Token,omitempty"`
	Expiration      time.Time `json:"Expiration,omitempty"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	roleArn, sourceIdentity, ok := r.BasicAuth()

	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	assumecnf, err := config.LoadDefaultConfig(
		r.Context(),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(
				credentials.NewStaticCredentialsProvider(
					awsAccessKeyID,
					awsSecretAccessKey,
					"",
				),
			),
		),
	)

	if err != nil {
		http.Error(w, "error authenticating", http.StatusInternalServerError)
		return
	}

	stsclient := sts.NewFromConfig(assumecnf)
	cb := "current-binding"
	provider := stscreds.NewAssumeRoleProvider(
		stsclient,
		strings.ReplaceAll(roleArn, "*", ":"),
		func(o *stscreds.AssumeRoleOptions) {
			o.Tags = []types.Tag{{Key: &cb, Value: &sourceIdentity}}
			o.TransitiveTagKeys = []string{cb}
		},
	)

	newcreds, err := provider.Retrieve(r.Context())
	if err != nil {
		http.Error(w, "error retrieving new creds using", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(creds{AccessKeyID: newcreds.AccessKeyID, SecretAccessKey: newcreds.SecretAccessKey, Token: newcreds.SessionToken, Expiration: newcreds.Expires})
	if err != nil {
		http.Error(w, "error marshalling creds", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, "error writing response", http.StatusInternalServerError)
		return
	}
}
