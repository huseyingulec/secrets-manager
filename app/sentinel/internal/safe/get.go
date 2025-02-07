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
	"context"
	"errors"
	"fmt"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/vmware-tanzu/secrets-manager/app/sentinel/logger"
	"github.com/vmware-tanzu/secrets-manager/core/env"
	"github.com/vmware-tanzu/secrets-manager/core/validation"
	"io"
	"net/http"
	"net/url"
)

func acquireSource(ctx context.Context) (*workloadapi.X509Source, bool) {
	resultChan := make(chan *workloadapi.X509Source)
	errorChan := make(chan error)

	go func() {
		source, err := workloadapi.NewX509Source(
			ctx, workloadapi.WithClientOptions(
				workloadapi.WithAddr(env.SpiffeSocketUrl()),
			),
		)

		if err != nil {
			errorChan <- err
			return
		}

		svid, err := source.GetX509SVID()
		if err != nil {
			logger.ErrorLn("acquireSource: I am having trouble fetching my identity from SPIRE.")
			logger.ErrorLn("acquireSource: I won’t proceed until you put me in a secured container.")
			errorChan <- err
			return
		}

		// Make sure that the binary is enclosed in a Pod that we trust.
		if !validation.IsSentinel(svid.ID.String()) {
			logger.ErrorLn("acquireSource: I don’t know you, and it’s crazy: '" + svid.ID.String() + "'")
			logger.ErrorLn("acquireSource: `safe` can only run from within the Sentinel container.")
			errorChan <- errors.New("acquireSource: I don’t know you, and it’s crazy: '" + svid.ID.String() + "'")
			return
		}

		resultChan <- source
	}()

	select {
	case source := <-resultChan:
		return source, true
	case err := <-errorChan:
		logger.ErrorLn("acquireSource: I cannot execute command because I cannot talk to SPIRE.", err.Error())
		return nil, false
	case <-ctx.Done():
		logger.ErrorLn("acquireSource: Operation was cancelled.")
		return nil, false
	}
}

func Get(showEncryptedSecrets bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, proceed := acquireSource(ctx)
	defer func() {
		if source == nil {
			return
		}
		err := source.Close()
		if err != nil {
			logger.ErrorLn("Get: Problem closing the workload source.")
		}
	}()
	if !proceed {
		return
	}

	authorizer := tlsconfig.AdaptMatcher(func(id spiffeid.ID) error {
		if validation.IsSafe(id.String()) {
			return nil
		}

		return errors.New("I don’t know you, and it’s crazy: '" + id.String() + "'")
	})

	safeUrl := "/sentinel/v1/secrets"
	if showEncryptedSecrets {
		safeUrl = "/sentinel/v1/secrets?reveal=true"
	}

	p, err := url.JoinPath(env.SafeEndpointUrl(), safeUrl)
	if err != nil {
		logger.ErrorLn("Get: I am having problem generating VSecM Safe secrets api endpoint URL.")
		return
	}

	tlsConfig := tlsconfig.MTLSClientConfig(source, source, authorizer)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	r, err := client.Get(p)
	if err != nil {
		logger.ErrorLn("Get: Problem connecting to VSecM Safe API endpoint URL.", err.Error())
		return
	}

	defer func(b io.ReadCloser) {
		if b == nil {
			return
		}
		err := b.Close()
		if err != nil {
			logger.ErrorLn("Get: Problem closing request body.")
		}
	}(r.Body)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.ErrorLn("Get: Unable to read the response body from VSecM Safe.")
		return
	}

	fmt.Println("")
	fmt.Println(string(body))
	fmt.Println("")
}
