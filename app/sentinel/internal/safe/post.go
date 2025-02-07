/*
|    Protect your secrets, protect your sensitive data.
:    Explore VMware Secrets Manager docs at https://vsecm.com/
</
<>/  keep your secrets… secret
>/
<>/' Copyright 2023–present VMware Secrets Manager contributors.
>/'  SPDX-License-Identifier: BSD-2-Clause
*/

package safe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/vmware-tanzu/secrets-manager/app/sentinel/logger"
	"github.com/vmware-tanzu/secrets-manager/core/crypto"
	data "github.com/vmware-tanzu/secrets-manager/core/entity/data/v1"
	entity "github.com/vmware-tanzu/secrets-manager/core/entity/data/v1"
	reqres "github.com/vmware-tanzu/secrets-manager/core/entity/reqres/safe/v1"
	"github.com/vmware-tanzu/secrets-manager/core/env"
	"github.com/vmware-tanzu/secrets-manager/core/validation"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func createAuthorizer() tlsconfig.Authorizer {
	return tlsconfig.AdaptMatcher(func(id spiffeid.ID) error {
		if validation.IsSafe(id.String()) {
			return nil
		}

		return errors.New("Post: I don’t know you, and it’s crazy: '" +
			id.String() + "'",
		)
	})
}

func decideBackingStore(backingStore string) data.BackingStore {
	switch data.BackingStore(backingStore) {
	case data.File:
		return data.File
	case data.Memory:
		return data.Memory
	default:
		return env.SafeBackingStore()
	}
}

func decideSecretFormat(format string) data.SecretFormat {
	switch data.SecretFormat(format) {
	case data.Json:
		return data.Json
	case data.Yaml:
		return data.Yaml
	default:
		return data.Json
	}
}

func newInputKeysRequest(ageSecretKey, agePublicKey, aesCipherKey string,
) reqres.KeyInputRequest {
	return reqres.KeyInputRequest{
		AgeSecretKey: ageSecretKey,
		AgePublicKey: agePublicKey,
		AesCipherKey: aesCipherKey,
	}
}

func newInitCompletedRequest() reqres.SentinelInitCompleteRequest {
	return reqres.SentinelInitCompleteRequest{}
}

func newSecretUpsertRequest(workloadId, secret string, namespaces []string,
	backingStore string, useKubernetes bool, template string, format string,
	encrypt, appendSecret bool, notBefore string, expires string,
) reqres.SecretUpsertRequest {
	bs := decideBackingStore(backingStore)
	f := decideSecretFormat(format)

	if notBefore == "" {
		notBefore = "now"
	}

	if expires == "" {
		expires = "never"
	}

	return reqres.SecretUpsertRequest{
		WorkloadId:    workloadId,
		BackingStore:  bs,
		Namespaces:    namespaces,
		UseKubernetes: useKubernetes,
		Template:      template,
		Format:        f,
		Encrypt:       encrypt,
		AppendValue:   appendSecret,
		Value:         secret,
		NotBefore:     notBefore,
		Expires:       expires,
	}
}

func respond(r *http.Response) {
	if r == nil {
		return
	}

	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		err := b.Close()
		if err != nil {
			logger.ErrorLn("Post: Problem closing request body.", err.Error())
		}
	}(r.Body)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.ErrorLn("Post: Unable to read the response body from VSecM Safe.", err.Error())
		return
	}

	fmt.Println("")
	fmt.Println(string(body))
	fmt.Println("")
}

func printEndpointError(err error) {
	logger.ErrorLn("Post: I am having problem generating VSecM Safe "+
		"secrets api endpoint URL.", err.Error())
}

func printPayloadError(err error) {
	logger.ErrorLn("Post: I am having problem generating the payload.", err.Error())
}

func doDelete(client *http.Client, p string, md []byte) {
	req, err := http.NewRequest(http.MethodDelete, p, bytes.NewBuffer(md))
	if err != nil {
		logger.ErrorLn("Post:Delete: Problem connecting to VSecM Safe API endpoint URL.", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := client.Do(req)
	if err != nil {
		logger.ErrorLn("Post:Delete: Problem connecting to VSecM Safe API endpoint URL.", err.Error())
		return
	}
	respond(r)
}

func doPost(client *http.Client, p string, md []byte) {
	r, err := client.Post(p, "application/json", bytes.NewBuffer(md))
	if err != nil {
		logger.ErrorLn("Post: Problem connecting to VSecM Safe API endpoint URL.", err.Error())
		return
	}
	respond(r)
}

func PostInitializationComplete(parentContext context.Context) {
	ctxWithTimeout, cancel := context.WithTimeout(
		parentContext,
		env.SafeSourceAcquisitionTimeout(),
	)
	defer cancel()

	sourceChan := make(chan *workloadapi.X509Source)
	proceedChan := make(chan bool)

	go func() {
		source, proceed := acquireSource(ctxWithTimeout)
		sourceChan <- source
		proceedChan <- proceed
	}()

	select {
	case <-ctxWithTimeout.Done():
		if errors.Is(ctxWithTimeout.Err(), context.DeadlineExceeded) {
			logger.ErrorLn("PostInit: I cannot execute command because I cannot talk to SPIRE.")
			return
		}

		logger.ErrorLn("PostInit: Operation was cancelled due to an unknown reason.")
	case source := <-sourceChan:
		defer func() {
			if source == nil {
				return
			}
			err := source.Close()
			if err != nil {
				logger.ErrorLn("Post: Problem closing the workload source.")
			}
		}()

		proceed := <-proceedChan

		if !proceed {
			return
		}

		authorizer := createAuthorizer()

		p, err := url.JoinPath(env.SafeEndpointUrl(), "/sentinel/v1/init-completed")
		if err != nil {
			printEndpointError(err)
			return
		}

		tlsConfig := tlsconfig.MTLSClientConfig(source, source, authorizer)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}

		sr := newInitCompletedRequest()

		md, err := json.Marshal(sr)
		if err != nil {
			printPayloadError(err)
			return
		}

		doPost(client, p, md)
	}
}

func Post(parentContext context.Context,
	sc entity.SentinelCommand,
) {
	ctxWithTimeout, cancel := context.WithTimeout(
		parentContext,
		env.SafeSourceAcquisitionTimeout(),
	)
	defer cancel()

	sourceChan := make(chan *workloadapi.X509Source)
	proceedChan := make(chan bool)

	go func() {
		source, proceed := acquireSource(ctxWithTimeout)
		sourceChan <- source
		proceedChan <- proceed
	}()

	select {
	case <-ctxWithTimeout.Done():
		if errors.Is(ctxWithTimeout.Err(), context.DeadlineExceeded) {
			logger.ErrorLn("Post: I cannot execute command because I cannot talk to SPIRE.")
			return
		}

		logger.ErrorLn("Post: Operation was cancelled due to an unknown reason.")
	case source := <-sourceChan:
		defer func() {
			if source == nil {
				return
			}
			err := source.Close()
			if err != nil {
				logger.ErrorLn("Post: Problem closing the workload source.")
			}
		}()

		proceed := <-proceedChan

		if !proceed {
			return
		}

		authorizer := createAuthorizer()

		if sc.InputKeys != "" {
			p, err := url.JoinPath(env.SafeEndpointUrl(), "/sentinel/v1/keys")
			if err != nil {
				printEndpointError(err)
				return
			}

			tlsConfig := tlsconfig.MTLSClientConfig(source, source, authorizer)
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsConfig,
				},
			}

			parts := strings.Split(sc.InputKeys, "\n")
			if len(parts) != 3 {
				printPayloadError(errors.New("post: Bad data! Very bad data"))
				return
			}

			sr := newInputKeysRequest(parts[0], parts[1], parts[2])
			md, err := json.Marshal(sr)
			if err != nil {
				printPayloadError(err)
				return
			}

			doPost(client, p, md)
			return
		}

		// Generate pattern-based random secrets if the secret has the prefix.
		if strings.HasPrefix(sc.Secret, env.SecretGenerationPrefix()) {
			sc.Secret = strings.Replace(
				sc.Secret, env.SecretGenerationPrefix(), "", 1,
			)
			newSecret, err := crypto.GenerateValue(sc.Secret)
			if err != nil {
				sc.Secret = "ParseError:" + sc.Secret
			} else {
				sc.Secret = newSecret
			}
		}

		p, err := url.JoinPath(env.SafeEndpointUrl(), "/sentinel/v1/secrets")
		if err != nil {
			printEndpointError(err)
			return
		}

		tlsConfig := tlsconfig.MTLSClientConfig(source, source, authorizer)
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}

		sr := newSecretUpsertRequest(sc.WorkloadId, sc.Secret, sc.Namespaces,
			sc.BackingStore, sc.UseKubernetes, sc.Template, sc.Format,
			sc.Encrypt, sc.AppendSecret, sc.NotBefore, sc.Expires)

		md, err := json.Marshal(sr)
		if err != nil {
			printPayloadError(err)
			return
		}

		if sc.DeleteSecret {
			doDelete(client, p, md)
			return
		}

		doPost(client, p, md)
	}
}
