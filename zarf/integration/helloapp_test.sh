#!/usr/bin/env bash
#
# Integration tests for the hello domain. Runs on its own or as part of the
# suite driven by api_test.sh.
#
# Usage:
#   zarf/integration/helloapp_test.sh [base-url]

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
preflight

# -----------------------------------------------------------------------------
# GET /hello/{user}

describe 'GET /hello/{user} greets the user'
if request GET /hello/dave; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'message' "$(jqv '.message')" 'Hello, dave'
fi

describe 'GET /hello/{user} handles an escaped user segment'
if request GET '/hello/Ada%20Lovelace'; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'message' "$(jqv '.message')" 'Hello, Ada Lovelace'
fi

describe 'GET /hello without a user does not match the route'
if request GET /hello/; then
	assert_eq 'status code' "${RESP_CODE}" '404'
fi

describe 'GET /hello/{user} rejects an unsupported method'
if request POST /hello/dave; then
	assert_eq 'status code' "${RESP_CODE}" '405'
fi
