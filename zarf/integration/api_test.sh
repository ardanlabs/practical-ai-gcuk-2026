#!/usr/bin/env bash
#
# Runs every integration suite for the api service against one running service,
# sharing a single preflight and a single pass/fail total. The suites drive the
# real HTTP API with curl, so the service (and its database) must already be
# running.
#
# Each suite is a *_test.sh file next to this one and can also be run on its own:
#
#   zarf/integration/userapp_test.sh
#
# Usage:
#   zarf/integration/api_test.sh [base-url]
#
# Environment:
#   BASE_URL        base url of the service (default http://localhost:3000)
#   WAIT_SECONDS    how long to wait for the service to come up (default 15)

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
preflight

# The suites are sourced rather than executed so they share the counters, the
# cleanup registry and the summary printed on exit.
for suite in "$(dirname "${BASH_SOURCE[0]}")"/*_test.sh; do
	if [[ "${suite}" == "${BASH_SOURCE[0]}" ]]; then
		continue
	fi

	source "${suite}"
done
