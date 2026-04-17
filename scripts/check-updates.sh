#!/usr/bin/env bash
# hexstrike-update-checker.sh
# Checks for updates to hexstrike-defense and dependencies

set -euo pipefail

# Configuration
REPO_OWNER="teodorbreajen"
REPO_NAME="hexstrike-defense"
GITHUB_API="https://api.github.com"

# Find project root
PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$PROJECT_ROOT"

# Get current version from CHANGELOG
CURRENT_VERSION=$(grep -oP '(?<=## \[)\d+\.\d+\.\d+' CHANGELOG.md 2>/dev/null | head -1 || echo "unknown")
GO_VERSION="1.21"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  HexStrike Defense Update Checker${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to get latest release
get_latest_release() {
    curl -s -H "Accept: application/vnd.github+json" \
        ${GITHUB_TOKEN:+-H "Authorization: Bearer $GITHUB_TOKEN"} \
        "$GITHUB_API/repos/$REPO_OWNER/$REPO_NAME/releases/latest" | \
        grep -oP '"tag_name":\s*"\K[^"]+' | sed 's/^v//' | head -1
}

# Check 1: Repository Version
echo -e "${YELLOW}[1/5] Checking repository version...${NC}"
LATEST_RELEASE=$(get_latest_release || echo "unknown")

if [[ "$LATEST_RELEASE" != "unknown" && "$CURRENT_VERSION" != "unknown" ]]; then
    if [[ "$LATEST_RELEASE" != "$CURRENT_VERSION" ]]; then
        echo -e "${GREEN}  Current:  v$CURRENT_VERSION${NC}"
        echo -e "${YELLOW}  Latest:   v$LATEST_RELEASE${NC}"
        echo -e "${RED}  Update available!${NC}"
    else
        echo -e "${GREEN}  ✓ Repository is up to date (v$CURRENT_VERSION)${NC}"
    fi
else
    echo -e "${YELLOW}  ? Could not determine version${NC}"
fi
echo ""

# Check 2: Go Dependencies
echo -e "${YELLOW}[2/5] Checking Go dependencies...${NC}"
cd "$PROJECT_ROOT/src/mcp-policy-proxy" 2>/dev/null || cd "$PROJECT_ROOT"
if [[ -f "go.mod" ]]; then
    OUTDATED=$(go list -m -u all 2>/dev/null | grep "\[upgradable\]" || true)
    if [[ -n "$OUTDATED" ]]; then
        echo -e "${YELLOW}  ⚠ Outdated dependencies:${NC}"
        echo "$OUTDATED" | while read -r line; do
            MODULE=$(echo "$line" | awk '{print $1}')
            echo -e "    ${YELLOW}  - $MODULE${NC}"
        done
    else
        echo -e "${GREEN}  ✓ All dependencies are current${NC}"
    fi
else
    echo -e "${YELLOW}  ? go.mod not found${NC}"
fi
echo ""

# Check 3: Docker Base Image
echo -e "${YELLOW}[3/5] Checking Docker base image...${NC}"
cd "$PROJECT_ROOT"
if [[ -f "src/mcp-policy-proxy/Dockerfile" ]]; then
    BASE_IMAGE=$(grep -oP 'FROM \K[^ ]+' src/mcp-policy-proxy/Dockerfile | head -1 || echo "unknown")
    echo -e "${BLUE}  Current base image: $BASE_IMAGE${NC}"
    
    if [[ "$BASE_IMAGE" == *"golang"* ]]; then
        IMAGE_GO_VERSION=$(echo "$BASE_IMAGE" | grep -oP '(?<=golang:)[0-9.]+' || echo "unknown")
        if [[ "$IMAGE_GO_VERSION" == "$GO_VERSION" ]]; then
            echo -e "${GREEN}  ✓ Go version matches ($GO_VERSION)${NC}"
        else
            echo -e "${YELLOW}  ⚠ Go version: image has $IMAGE_GO_VERSION, project uses $GO_VERSION${NC}"
        fi
    fi
else
    echo -e "${YELLOW}  ? Dockerfile not found${NC}"
fi
echo ""

# Check 4: GitHub Actions
echo -e "${YELLOW}[4/5] Checking GitHub Actions...${NC}"
if [[ -d ".github/workflows" ]]; then
    WORKFLOW_COUNT=$(find .github/workflows -name "*.yml" -o -name "*.yaml" 2>/dev/null | wc -l)
    echo -e "${BLUE}  Total workflows: $WORKFLOW_COUNT${NC}"
    
    # Check for unpinned actions
    UNPINNED=$(grep -r "uses:.*@master\|uses:.*@main" .github/workflows/ 2>/dev/null || true)
    if [[ -n "$UNPINNED" ]]; then
        echo -e "${YELLOW}  ⚠ Found unpinned actions (recommend using @vX or SHA)${NC}"
    else
        echo -e "${GREEN}  ✓ All actions are pinned${NC}"
    fi
else
    echo -e "${YELLOW}  ? .github/workflows not found${NC}"
fi
echo ""

# Check 5: Security Configuration
echo -e "${YELLOW}[5/5] Checking security configuration...${NC}"
if [[ -f ".github/dependabot.yml" ]]; then
    echo -e "${GREEN}  ✓ Dependabot configured${NC}"
else
    echo -e "${RED}  ⚠ Dependabot not configured${NC}"
fi

if grep -q "codeql-action" .github/workflows/*.yml 2>/dev/null; then
    echo -e "${GREEN}  ✓ CodeQL scanning enabled${NC}"
else
    echo -e "${YELLOW}  ⚠ CodeQL not found${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  Current Version: v$CURRENT_VERSION"
echo -e "  Latest Release:  v$LATEST_RELEASE"
echo ""

# Commands to run
echo -e "${BLUE}Commands to update:${NC}"
echo ""
echo -e "  ${YELLOW}# Update Go dependencies:${NC}"
echo -e "  cd src/mcp-policy-proxy && go get -u all && go mod tidy"
echo ""
echo -e "  ${YELLOW}# Check for repository updates:${NC}"
echo -e "  gh release list"
echo ""
echo -e "  ${YELLOW}# Update Docker base image:${NC}"
echo -e "  # Edit src/mcp-policy-proxy/Dockerfile and update golang version"
echo ""
