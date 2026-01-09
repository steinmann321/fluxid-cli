#!/bin/bash
# Manual test script to verify each agent CLI works correctly

set -e

echo "Testing Claude..."
claude --print --output-format stream-json --verbose "Say Hi" 2>&1 | head -5
echo "✓ Claude works"
echo ""

echo "Testing Codex..."
codex exec --json "Say Hi" 2>&1 | head -5
echo "✓ Codex works"
echo ""

echo "Testing Opencode..."
opencode run --format json "Say Hi" 2>&1 | head -5
echo "✓ Opencode works"
echo ""

echo "Testing Gemini..."
gemini --output-format stream-json "Say Hi" 2>&1 | head -5
echo "✓ Gemini works"
echo ""

echo "All agents tested successfully!"
