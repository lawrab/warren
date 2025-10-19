#!/bin/bash
# Setup script for git hooks
# Run this once to configure git to use .githooks directory

set -e

HOOKS_DIR=".githooks"

echo "Setting up git hooks..."

# Configure git to use .githooks directory
git config core.hooksPath "$HOOKS_DIR"

# Make hooks executable
chmod +x "$HOOKS_DIR"/pre-commit

echo "✅ Git hooks configured successfully!"
echo ""
echo "Git will now use hooks from $HOOKS_DIR/"
echo ""
echo "Available hooks:"
echo "  • pre-commit: Runs format, lint, and test checks"
echo ""
echo "To skip hooks (not recommended), use:"
echo "  git commit --no-verify"
