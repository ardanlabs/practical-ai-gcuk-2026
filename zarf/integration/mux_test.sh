#!/usr/bin/env bash
#
# Integration tests for behaviour owned by the mux rather than any one domain.
# Runs on its own or as part of the suite driven by api_test.sh.
#
# Usage:
#   zarf/integration/mux_test.sh [base-url]

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
preflight

# -----------------------------------------------------------------------------
# Unknown routes

describe 'GET an unknown route returns not found'
if request GET /does-not-exist; then
	assert_eq 'status code' "${RESP_CODE}" '404'
fi
