#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

if command -v supabase >/dev/null 2>&1; then
  SUPABASE_CMD=(supabase)
elif command -v npx >/dev/null 2>&1; then
  SUPABASE_CMD=(npx supabase)
else
  echo "Supabase CLI not found" >&2
  exit 1
fi

PROJECT_REF="${SUPABASE_PROJECT_REF:-}"
if [[ -z "${PROJECT_REF}" && -f "${ROOT_DIR}/supabase/.temp/project-ref" ]]; then
  PROJECT_REF="$(tr -d '[:space:]' < "${ROOT_DIR}/supabase/.temp/project-ref")"
fi

if [[ -z "${PROJECT_REF}" ]]; then
  echo "Missing SUPABASE_PROJECT_REF (or run supabase link --project-ref <ref>)" >&2
  exit 1
fi

echo "Linking Supabase project ${PROJECT_REF}..."
"${SUPABASE_CMD[@]}" link --project-ref "${PROJECT_REF}" >/dev/null

echo "Pushing local database migrations to remote..."
"${SUPABASE_CMD[@]}" db push --linked

echo "Deploying all local edge functions to remote..."
FUNCTION_DIRS=()
for index_file in "${ROOT_DIR}"/supabase/functions/*/index.ts; do
  if [[ ! -f "${index_file}" ]]; then
    continue
  fi
  fn_dir="$(basename "$(dirname "${index_file}")")"
  if [[ "${fn_dir}" == "_shared" ]]; then
    continue
  fi
  FUNCTION_DIRS+=("${fn_dir}")
done

if [[ ${#FUNCTION_DIRS[@]} -eq 0 ]]; then
  echo "No deployable edge functions found under supabase/functions."
else
  for fn in "${FUNCTION_DIRS[@]}"; do
    echo "Deploying function: ${fn}"
    "${SUPABASE_CMD[@]}" functions deploy "${fn}" --project-ref "${PROJECT_REF}" --no-verify-jwt
  done
fi

echo "Supabase full deploy complete."
