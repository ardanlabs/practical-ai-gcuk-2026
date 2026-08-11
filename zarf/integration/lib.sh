#!/usr/bin/env bash
#
# Shared harness for the integration test suites. Every *_test.sh file sources
# this, either directly when run on its own or by way of api_test.sh when the
# whole suite runs.
#
# Environment:
#   BASE_URL        base url of the service (default http://localhost:3000)
#   WAIT_SECONDS    how long to wait for the service to come up (default 15)

# Sourcing more than once would fail on the readonly assignments below and reset
# the counters, so the second and later attempts are no-ops.
if [[ -n "${INTEGRATION_LIB_SOURCED:-}" ]]; then
	return 0
fi
INTEGRATION_LIB_SOURCED=1

set -u -o pipefail

# When a suite is run on its own the first argument selects the service, so the
# arguments the sourcing script received are honoured here.
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

# skip reports a test that could not run, without counting it either way.
skip() {
	printf '  %s!%s %s\n' "${YELLOW}" "${RESET}" "$1"
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

	# The trailing sentinel keeps command substitution from eating the metadata:
	# it strips trailing newlines, so an empty body with an empty content type
	# (a 204, for instance) would otherwise shift the fields by one.
	local raw
	raw="$(curl --silent --show-error --max-time 10 \
		--request "${method}" \
		--write-out '\n%{http_code}\n%{content_type}\n#' \
		"$@" "${BASE_URL}${path}" 2>&1)"

	local rc=$?
	raw="${raw%$'\n'#}"
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

# suite_id returns a value unique to this process, so the suites can be run
# repeatedly against the same database without colliding on unique columns.
suite_id() {
	printf '%s-%s-%s' "$$" "${RANDOM}" "$(date +%s)"
}

# -----------------------------------------------------------------------------
# Lifecycle

# on_exit registers a cleanup function to run when the process exits. A single
# EXIT trap is shared, so a suite can clean up after itself without stomping on
# the trap another suite (or the runner) installed.
cleanups=()

on_exit() {
	cleanups+=("$1")
}

_exit_handler() {
	local fn
	for fn in ${cleanups[@]+"${cleanups[@]}"}; do
		"${fn}" || true
	done

	summary
}

trap _exit_handler EXIT

# preflight checks the tooling and waits for the service to answer. It runs once
# per process: a suite sourced by api_test.sh inherits the runner's check.
preflight() {
	if [[ -n "${INTEGRATION_PREFLIGHT_DONE:-}" ]]; then
		return 0
	fi
	INTEGRATION_PREFLIGHT_DONE=1

	local tool
	for tool in curl jq; do
		if ! command -v "${tool}" >/dev/null 2>&1; then
			printf '%serror:%s %s is required\n' "${RED}" "${RESET}" "${tool}" >&2
			exit 1
		fi
	done

	printf '%sintegration tests against %s%s\n\n' "${BOLD}" "${BASE_URL}" "${RESET}"

	printf 'waiting for the service'
	local i
	for ((i = 0; i < WAIT_SECONDS; i++)); do
		if curl --silent --fail --max-time 2 "${BASE_URL}/healthcheck" >/dev/null 2>&1; then
			printf ' ready\n\n'
			return 0
		fi

		printf '.'
		sleep 1
	done

	printf '\n%serror:%s service not reachable at %s after %ss\n' "${RED}" "${RESET}" "${BASE_URL}" "${WAIT_SECONDS}" >&2
	printf '       start it with: make run\n' >&2
	exit 1
}

# summary reports the totals and exits non zero when anything failed.
summary() {
	printf '\n%s%d passed, %d failed%s\n' "${BOLD}" "${passed}" "${failed}" "${RESET}"

	if [[ ${failed} -gt 0 ]]; then
		exit 1
	fi

	exit 0
}

