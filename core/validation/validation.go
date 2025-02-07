/*
|    Protect your secrets, protect your sensitive data.
:    Explore VMware Secrets Manager docs at https://vsecm.com/
</
<>/  keep your secrets… secret
>/
<>/' Copyright 2023–present VMware Secrets Manager contributors.
>/'  SPDX-License-Identifier: BSD-2-Clause
*/

package validation

import (
	"github.com/vmware-tanzu/secrets-manager/core/env"
	"strings"
)

// IsSentinel returns true if the given SPIFFEID is a Sentinel ID.
// It does this by checking if the SPIFFEID has the SentinelSpiffeIdPrefix as its prefix.
func IsSentinel(spiffeid string) bool {
	return strings.HasPrefix(spiffeid, env.SentinelSpiffeIdPrefix())
}

// IsSafe returns true if the given SPIFFEID is a Safe ID.
// It does this by checking if the SPIFFEID has the SafeSpiffeIdPrefix as its prefix.
func IsSafe(spiffeid string) bool {
	return strings.HasPrefix(spiffeid, env.SafeSpiffeIdPrefix())
}

// IsWorkload returns true if the given SPIFFEID is a WorkloadId ID.
// It does this by checking if the SPIFFEID has the WorkloadSpiffeIdPrefix as its prefix.
func IsWorkload(spiffeid string) bool {
	return strings.HasPrefix(spiffeid, env.WorkloadSpiffeIdPrefix())
}
