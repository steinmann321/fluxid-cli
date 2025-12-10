#!/usr/bin/env bash
if [[ -n "${FLUXID_SESSION_ID:-}" ]]; then
  echo "ENV_CHECK_PASSED:${FLUXID_SESSION_ID}"
else
  echo "ENV_CHECK_FAILED"
fi
