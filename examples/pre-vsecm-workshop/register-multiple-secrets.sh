#!/usr/bin/env bash

# /*
# |    Protect your secrets, protect your sensitive data.
# :    Explore VMware Secrets Manager docs at https://vsecm.com/
# </
# <>/  keep your secrets… secret
# >/
# <>/' Copyright 2023–present VMware Secrets Manager contributors.
# >/'  SPDX-License-Identifier: BSD-2-Clause
# */

source ./env.sh

kubectl exec "$SENTINEL" -n vsecm-system -- safe \
  -w "example" \
  -n "default" \
  -s '{"name": "USERNAME", "value": "operator"}' \
  -a

kubectl exec "$SENTINEL" -n vsecm-system -- safe \
  -w "example" \
  -n "default" \
  -s '{"name": "PASSWORD", "value": "!KeepYourSecrets!"}' \
  -a
