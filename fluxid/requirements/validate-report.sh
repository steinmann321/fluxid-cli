#!/usr/bin/env bash
# Validates YAML report against report-schema.yaml requirements
# Usage: ./validate-report.sh <report-file-path>
# Exit: 0 on success, 1 on validation failure

set -euo pipefail

REPORT_FILE="${1:-}"

if [[ -z "$REPORT_FILE" ]]; then
  echo "Error: No report file specified" >&2
  echo "Usage: $0 <report-file-path>" >&2
  exit 1
fi

if [[ ! -f "$REPORT_FILE" ]]; then
  echo "Error: Report file not found: $REPORT_FILE" >&2
  exit 1
fi

# Get history file path
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HISTORY_FILE=$("$SCRIPT_DIR/files.sh" --history)

# Check if history file exists
if [[ ! -f "$HISTORY_FILE" ]]; then
  echo "Error: History file not found: $HISTORY_FILE" >&2
  echo "The workflow-loop-history.md file must exist for validation to pass" >&2
  exit 1
fi

# Use Python to validate YAML structure
python3 - "$REPORT_FILE" <<'PYTHON_SCRIPT'
import sys
import yaml
from datetime import datetime

def validate_report(file_path):
    errors = []

    # Load YAML
    try:
        with open(file_path, 'r') as f:
            data = yaml.safe_load(f)
    except yaml.YAMLError as e:
        errors.append(f"Invalid YAML format: {e}")
        return errors
    except Exception as e:
        errors.append(f"Failed to read file: {e}")
        return errors

    if not isinstance(data, dict):
        errors.append("Report must be a YAML object")
        return errors

    # Check required fields
    required_fields = ['command', 'artifact', 'timestamp', 'status', 'issues']
    for field in required_fields:
        if field not in data:
            errors.append(f"Missing required field: {field}")

    # Validate command (string)
    if 'command' in data and not isinstance(data['command'], str):
        errors.append("Field 'command' must be a string")

    # Validate artifact (string, single token, no path/extension characters)
    if 'artifact' in data:
        if not isinstance(data['artifact'], str):
            errors.append("Field 'artifact' must be a string")
        else:
            artifact = data['artifact']
            # Disallow path-like or filename-like artifacts (no '/' or '.')
            if '/' in artifact or '.' in artifact:
                errors.append(
                    "Field 'artifact' must be a single token without path or extension characters (no '/' or '.'): "
                    f"got: {artifact}"
                )

    # Validate timestamp (ISO 8601 date-time string or datetime object)
    if 'timestamp' in data:
        if isinstance(data['timestamp'], str):
            try:
                # Parse ISO 8601 format
                datetime.fromisoformat(data['timestamp'].replace('Z', '+00:00'))
            except ValueError:
                errors.append(f"Field 'timestamp' must be ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ), got: {data['timestamp']}")
        elif not isinstance(data['timestamp'], datetime):
            errors.append(f"Field 'timestamp' must be a string or datetime, got: {type(data['timestamp']).__name__}")

    # Validate status (enum: PASS or FAIL)
    if 'status' in data:
        if data['status'] not in ['PASS', 'FAIL']:
            errors.append(f"Field 'status' must be 'PASS' or 'FAIL', got: {data['status']}")

    # Validate issues structure
    if 'issues' in data:
        issues = data['issues']
        if not isinstance(issues, dict):
            errors.append("Field 'issues' must be an object")
        else:
            required_categories = ['blockers', 'defects', 'concerns', 'observations', 'enhancements']
            for category in required_categories:
                if category not in issues:
                    errors.append(f"Missing required issues category: {category}")
                elif not isinstance(issues[category], list):
                    errors.append(f"Issues category '{category}' must be an array")
                else:
                    # Validate each issue in the category
                    for idx, issue in enumerate(issues[category]):
                        if not isinstance(issue, dict):
                            errors.append(f"Issue in '{category}[{idx}]' must be an object")
                            continue

                        if 'message' not in issue:
                            errors.append(f"Issue in '{category}[{idx}]' missing required field: message")
                        elif not isinstance(issue['message'], str):
                            errors.append(f"Issue 'message' in '{category}[{idx}]' must be a string")

                        # Validate optional fields if present
                        for field in ['location', 'code', 'suggestion', 'reference']:
                            if field in issue and not isinstance(issue[field], str):
                                errors.append(f"Issue '{field}' in '{category}[{idx}]' must be a string")

            # Check for extra categories (not allowed per schema: additionalProperties: false)
            extra_categories = set(issues.keys()) - set(required_categories)
            if extra_categories:
                errors.append(f"Unknown issues categories: {', '.join(extra_categories)}")

    # Validate optional fields if present
    if 'next_steps' in data:
        if not isinstance(data['next_steps'], list):
            errors.append("Field 'next_steps' must be an array")
        else:
            for idx, item in enumerate(data['next_steps']):
                if not isinstance(item, str):
                    errors.append(f"Next steps item [{idx}] must be a string")

    if 'summary' in data:
        if not isinstance(data['summary'], str):
            errors.append("Field 'summary' must be a string")

    return errors

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Error: Report file path required", file=sys.stderr)
        sys.exit(1)

    errors = validate_report(sys.argv[1])

    if errors:
        print("Report validation failed:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        print(file=sys.stderr)
        print(f"Report file: {sys.argv[1]}", file=sys.stderr)
        sys.exit(1)

    # Success - no output on valid report
    sys.exit(0)
PYTHON_SCRIPT
