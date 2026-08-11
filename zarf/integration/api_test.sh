#!/usr/bin/env bash
#
# Integration tests for the api service. These drive the real HTTP API with
# curl, so the service (and its database) must already be running.
#
# Usage:
#   zarf/integration/api_test.sh [base-url]
#
# Environment:
#   BASE_URL        base url of the service (default http://localhost:3000)
#   WAIT_SECONDS    how long to wait for the service to come up (default 15)

set -u -o pipefail

BASE_URL="${1:-${BASE_URL:-http://localhost:3000}}"
BASE_URL="${BASE_URL%/}"
WAIT_SECONDS="${WAIT_SECONDS:-15}"

readonly BASE_URL WAIT_SECONDS

if [[ -t 1 ]]; then
	RED=$'\e[31m'; GREEN=$'\e[32m'; YELLOW=$'\e[33m'; BOLD=$'\e[1m'; DIM=$'\e[2m'; RESET=$'\e[0m'
else
	RED=''; GREEN=''; YELLOW=''; BOLD=''; DIM=''; RESET=''
fi

passed=0
failed=0
current=''

# -----------------------------------------------------------------------------
# Test harness

# describe names the test currently executing.
describe() {
	current="$1"
	printf '%s— %s%s\n' "${BOLD}" "${current}" "${RESET}"
}

pass() {
	passed=$((passed + 1))
	printf '  %s✓%s %s\n' "${GREEN}" "${RESET}" "$1"
}

fail() {
	failed=$((failed + 1))
	printf '  %s✗%s %s\n' "${RED}" "${RESET}" "$1"
	if [[ $# -gt 1 ]]; then
		printf '      %s\n' "${@:2}"
	fi
}

# assert_eq compares two values for exact equality.
assert_eq() {
	local what="$1" got="$2" want="$3"

	if [[ "${got}" == "${want}" ]]; then
		pass "${what} = ${want}"
		return 0
	fi

	fail "${what}" "want: ${want}" "got:  ${got}"
	return 1
}

# request performs a request and stores the results in the RESP_* globals. The
# body is captured separately from the status code and content type so each can
# be asserted on individually. The exact curl invocation is echoed first so a
# failing assertion can be reproduced by hand.
request() {
	local method="$1" path="$2"
	shift 2

	# The reported command omits the plumbing flags used to capture the status
	# code and content type, so it can be copy-pasted as-is.
	local -a cmd=(curl --request "${method}" "$@" "${BASE_URL}${path}")
	printf '  %s$ %s%s\n' "${DIM}" "${cmd[*]}" "${RESET}"

	local raw
	raw="$(curl --silent --show-error --max-time 10 \
		--request "${method}" \
		--write-out '\n%{http_code}\n%{content_type}' \
		"$@" "${BASE_URL}${path}" 2>&1)"

	local rc=$?
	if [[ ${rc} -ne 0 ]]; then
		RESP_BODY="${raw}"
		RESP_CODE='000'
		RESP_TYPE=''
		fail "curl ${method} ${path} failed (exit ${rc})" "${raw}"
		return 1
	fi

	RESP_TYPE="${raw##*$'\n'}"
	raw="${raw%$'\n'*}"
	RESP_CODE="${raw##*$'\n'}"
	RESP_BODY="${raw%$'\n'*}"

	return 0
}

# jqv extracts a value from the last response body, printing nothing when the
# body is not valid json or the path is absent.
jqv() {
	printf '%s' "${RESP_BODY}" | jq --exit-status --raw-output "$1" 2>/dev/null
}

# -----------------------------------------------------------------------------
# Preflight

if ! command -v curl >/dev/null 2>&1; then
	printf '%serror:%s curl is required\n' "${RED}" "${RESET}" >&2
	exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
	printf '%serror:%s jq is required (brew install jq)\n' "${RED}" "${RESET}" >&2
	exit 1
fi

printf '%sintegration tests against %s%s\n\n' "${BOLD}" "${BASE_URL}" "${RESET}"

printf 'waiting for the service'
for ((i = 0; i < WAIT_SECONDS; i++)); do
	if curl --silent --fail --max-time 2 "${BASE_URL}/healthcheck" >/dev/null 2>&1; then
		printf ' ready\n\n'
		break
	fi

	printf '.'
	sleep 1
done

if ! curl --silent --fail --max-time 2 "${BASE_URL}/healthcheck" >/dev/null 2>&1; then
	printf '\n%serror:%s service not reachable at %s after %ss\n' "${RED}" "${RESET}" "${BASE_URL}" "${WAIT_SECONDS}" >&2
	printf '       start it with: make run\n' >&2
	exit 1
fi

# -----------------------------------------------------------------------------
# GET /healthcheck

describe 'GET /healthcheck reports liveness'
if request GET /healthcheck; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'status field' "$(jqv '.status')" 'OK'
fi

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

# -----------------------------------------------------------------------------
# GET /users

describe 'GET /users returns the user collection'
if request GET /users; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'body is a json array' "$(jqv 'type')" 'array'

	count="$(jqv 'length')"
	if [[ -n "${count}" ]]; then
		pass "returned ${count} user(s)"
	fi

	if [[ "${count:-0}" -gt 0 ]]; then
		# Every element must carry the full app layer representation with the
		# documented primitive types.
		shape="$(jqv '
			map(
				(.id          | type == "number") and
				(.name        | type == "string") and
				(.email       | type == "string") and
				(.enabled     | type == "boolean") and
				(.dateCreated | type == "string") and
				(.dateUpdated | type == "string")
			) | all')"
		assert_eq 'every user has the expected fields and types' "${shape}" 'true'

		rfc3339='^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([.][0-9]+)?(Z|[+-][0-9]{2}:[0-9]{2})$'
		assert_eq 'dates are RFC3339' \
			"$(jqv "map((.dateCreated | test(\"${rfc3339}\")) and (.dateUpdated | test(\"${rfc3339}\"))) | all")" \
			'true'

		assert_eq 'emails are non-empty' "$(jqv 'map(.email | length > 0) | all')" 'true'
	else
		printf '  %s!%s no users in the database, field assertions skipped\n' "${YELLOW}" "${RESET}"
		printf '      seed some users to exercise the full payload\n'
	fi
fi

describe 'GET /users rejects an unsupported method'
if request POST /users; then
	assert_eq 'status code' "${RESP_CODE}" '405'
fi

# -----------------------------------------------------------------------------
# Unknown routes

describe 'GET an unknown route returns not found'
if request GET /does-not-exist; then
	assert_eq 'status code' "${RESP_CODE}" '404'
fi

# -----------------------------------------------------------------------------
# Summary

printf '\n%s%d passed, %d failed%s\n' "${BOLD}" "${passed}" "${failed}" "${RESET}"

if [[ ${failed} -gt 0 ]]; then
	exit 1
fi
