#!/usr/bin/env bash
# hexstrike-update-checker.sh
# Checks for updates to hexstrike-defense and dependencies

set -euo pipefail

# Configuration
REPO_OWNER="teodorbreajen"
REPO_NAME="hexstrike-defense"
GITHUB_API="https://api.github.com"
CURRENT_VERSION=$(grep -oP '(?<=version:\s*v)\d+\.\d+\.\d+' CHANGELOG.md 2>/dev/null || echo "unknown")
GO_VERSION="1.21"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  HexStrike Defense Update Checker${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Function to check GitHub API rate limit
check_rate_limit() {
    if [[ -n "${GITHUB_TOKEN:-}" ]]; then
        RESPONSE=$(curl -s -H "Authorization: Bearer $GITHUB_TOKEN" \
            -H "Accept: application/vnd.github+json" \
            "$GITHUB_API/rate_limit")
        REMAINING=$(echo "$RESPONSE" | grep -oP '"remaining":\s*\K\d+' | head -1)
        echo -e "${BLUE}GitHub API Rate Limit: $REMAINING remaining${NC}"
    fi
}

# Function to get latest release from GitHub
get_latest_release() {
    local endpoint="$1"
    curl -s -H "Accept: application/vnd.github+json" \
        ${GITHUB_TOKEN:+-H "Authorization: Bearer $GITHUB_TOKEN"} \
        "$GITHUB_API/repos/$REPO_OWNER/$REPO_NAME/$endpoint" | \
        grep -oP '"tag_name":\s*"\K[^"]+' | head -1 | sed 's/^v//'
}

# Function to compare versions
version_compare() {
    local v1="$1"
    local v2="$2"
    
    if [[ "$v1" == "$v2" ]]; then
        echo "equal"
    elif [[ "$v1" == "unknown" || "$v1" == "" ]]; then
        echo "unknown"
    else
        # Using sort -V for version comparison
        local highest=$(printf '%s\n%s' "$v1" "$v2" | sort -V | tail -n1)
        if [[ "$highest" == "$v2" && "$v1" != "$v2" ]]; then
            echo "update_available"
        else
            echo "current"
        fi
    fi
}

# Check 1: Repository Version
echo -e "${YELLOW}[1/5] Checking repository version...${NC}"
LATEST_RELEASE=$(get_latest_release "releases/latest")
if [[ -z "$LATEST_RELEASE" ]]; then
    LATEST_RELEASE=$(get_latest_release "releases")
fi

if [[ -n "$LATEST_RELEASE" ]]; then
    VERSION_STATUS=$(version_compare "$CURRENT_VERSION" "$LATEST_RELEASE")
    case "$VERSION_STATUS" in
        "update_available")
            echo -e "${RED}  ⚠ Update available: v$CURRENT_VERSION → v$LATEST_RELEASE${NC}"
            echo -e "${RED}    Run: gh release create v$LATEST_RELEASE${NC}"
            ;;
        "current")
            echo -e "${GREEN}  ✓ Repository is up to date (v$CURRENT_VERSION)${NC}"
            ;;
        *)
            echo -e "${YELLOW}  ? Could not determine version${NC}"
            ;;
    esac
else
    echo -e "${YELLOW}  ? Could not fetch latest release${NC}"
fi
echo ""

# Check 2: Go Dependencies
echo -e "${YELLOW}[2/5] Checking Go dependencies...${NC}"
cd src/mcp-policy-proxy 2>/dev/null || cd .
if [[ -f "go.mod" ]]; then
    echo -e "${BLUE}  Running go mod outdated...${NC}"
    OUTDATED=$(go list -m -u all 2>/dev/null | grep "\[upgradable\]" | head -10 || true)
    if [[ -n "$OUTDATED" ]]; then
        echo -e "${YELLOW}  ⚠ Outdated dependencies found:${NC}"
        echo "$OUTDATED" | while read -r line; do
            MODULE=$(echo "$line" | awk '{print $1}')
            OLD=$(echo "$line" | grep -oP '\bv\K[0-9.]+' | head -1)
            NEW=$(echo "$line" | grep -oP '\[upgradable\] v\K[0-9.]+' | sed 's/\[upgradable\] //' || echo "latest")
            echo -e "    ${YELLOW}  - $MODULE: $OLD → $NEW${NC}"
        done
        echo -e "${YELLOW}    Run: go get -u all${NC}"
    else
        echo -e "${GREEN}  ✓ All dependencies are current${NC}"
    fi
else
    echo -e "${YELLOW}  ? go.mod not found${NC}"
fi
echo ""

# Check 3: Docker Base Image
echo -e "${YELLOW}[3/5] Checking Docker base image...${NC}"
if [[ -f "Dockerfile" ]]; then
    BASE_IMAGE=$(grep -oP 'FROM \K[^ ]+' Dockerfile | head -1 || echo "unknown")
    echo -e "${BLUE}  Current base image: $BASE_IMAGE${NC}"
    
    # Check for common base images
    if [[ "$BASE_IMAGE" == *"golang"* ]]; then
        IMAGE_GO_VERSION=$(echo "$BASE_IMAGE" | grep -oP '(?<=golang:)[0-9.]+' || echo "unknown")
        if [[ "$IMAGE_GO_VERSION" != "$GO_VERSION" ]]; then
            echo -e "${YELLOW}  ⚠ Go version mismatch: image has $IMAGE_GO_VERSION, project uses $GO_VERSION${NC}"
        else
            echo -e "${GREEN}  ✓ Go version matches ($GO_VERSION)${NC}"
        fi
    fi
    
    if [[ "$BASE_IMAGE" == *"alpine"* ]]; then
        ALPINE_VERSION=$(echo "$BASE_IMAGE" | grep -oP '(?<=alpine:)[0-9.]+' || echo "unknown")
        echo -e "${BLUE}  Alpine version: $ALPINE_VERSION${NC}"
    fi
else
    echo -e "${YELLOW}  ? Dockerfile not found${NC}"
fi
echo ""

# Check 4: GitHub Actions Workflows
echo -e "${YELLOW}[4/5] Checking GitHub Actions versions...${NC}"
if [[ -d ".github/workflows" ]]; then
    echo -e "${BLUE}  Checking for outdated actions...${NC}"
    # Check for pinned versions vs @master
    UNPINNED=$(grep -r "uses:.*@master\|uses:.*@main" .github/workflows/ 2>/dev/null || true)
    if [[ -n "$UNPINNED" ]]; then
        echo -e "${YELLOW}  ⚠ Found unpinned actions (recommend using @vX or SHA):${NC}"
        echo "$UNPINNED" | head -5 | while read -r line; do
            echo -e "    ${YELLOW}  - $(echo "$line" | cut -d: -f2- | xargs)${NC}"
        done
    else
        echo -e "${GREEN}  ✓ All actions are pinned to versions${NC}"
    fi
    
    # Count workflows
    WORKFLOW_COUNT=$(find .github/workflows -name "*.yml" -o -name "*.yaml" | wc -l)
    echo -e "${BLUE}  Total workflows: $WORKFLOW_COUNT${NC}"
else
    echo -e "${YELLOW}  ? .github/workflows not found${NC}"
fi
echo ""

# Check 5: Security Advisories
echo -e "${YELLOW}[5/5] Checking for security updates...${NC}"
# Check if dependabot is configured
if [[ -f ".github/dependabot.yml" ]]; then
    echo -e "${GREEN}  ✓ Dependabot is configured${NC}"
else
    echo -e "${RED}  ⚠ Dependabot not configured - recommend adding .github/dependabot.yml${NC}"
fi

# Check for CodeQL
if grep -q "codeql-action" .github/workflows/*.yml 2>/dev/null; then
    echo -e "${GREEN}  ✓ CodeQL scanning is enabled${NC}"
else
    echo -e "${YELLOW}  ⚠ CodeQL not found - recommend adding for security scanning${NC}"
fi
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Current Version: v$CURRENT_VERSION${NC}"
echo -e "${BLUE}  Latest Release:  v$LATEST_RELEASE${NC}"
echo ""

# Generate update commands
if [[ "$VERSION_STATUS" == "update_available" ]]; then
    echo -e "${GREEN}To update to v$LATEST_RELEASE:${NC}"
    echo "  1. Create release notes:"
    echo "     gh release create v$LATEST_RELEASE"
    echo ""
    echo "  2. Or create a PR for the update:"
    echo "     git checkout -b chore/update-to-v$LATEST_RELEASE"
    echo "     # Update CHANGELOG.md"
    echo "     git commit -m 'chore: update to v$LATEST_RELEASE'"
    echo "     git push origin chore/update-to-v$LATEST_RELEASE"
    echo "     gh pr create"
    echo ""
fi

echo -e "${BLUE}To check for dependency updates:${NC}"
echo "  cd src/mcp-policy-proxy && go get -u all"
echo ""
