#!/usr/bin/env bash
#
# Integration tests for the check domain. Runs on its own or as part of the
# suite driven by api_test.sh.
#
# Usage:
#   zarf/integration/checkapp_test.sh [base-url]

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
preflight

# -----------------------------------------------------------------------------
# GET /healthcheck

describe 'GET /healthcheck reports liveness'
if request GET /healthcheck; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'status field' "$(jqv '.status')" 'OK'
fi
