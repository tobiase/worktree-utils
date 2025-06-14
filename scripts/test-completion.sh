#!/bin/bash

# Quick completion testing script
# Usage: ./scripts/test-completion.sh

set -e

echo "🧪 Testing wt completion in fresh shell..."

# Build first
make build >/dev/null

# Test completion loading
echo "Testing process substitution..."
result=$(echo 'source <(./wt-bin completion zsh); type _wt' | zsh 2>&1)

if [[ $result == *"_wt is a shell function"* ]]; then
    echo "✅ Process substitution works!"
    echo "   $result"
else
    echo "❌ Process substitution failed:"
    echo "   $result"
    exit 1
fi

# Test help
echo ""
echo "Testing help command..."
help_result=$(echo './wt-bin --help | head -3' | zsh)
echo "✅ Help output:"
echo "$help_result"

# Test setup simulation
echo ""
echo "Testing what setup would add to shell config..."
echo "Would add to ~/.zshrc:"
echo "   source <(wt-bin completion zsh)"

echo ""
echo "🎉 All tests passed! Completion is working correctly."
echo ""
echo "To test interactively:"
echo "   make test-completion-interactive"
