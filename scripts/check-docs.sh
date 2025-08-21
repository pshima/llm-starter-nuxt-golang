#!/bin/bash

# Documentation Validation Script
# This script checks if documentation files exist and have been updated recently

echo "📚 Documentation Validation Check"
echo "================================="

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Required documentation files
REQUIRED_DOCS=(
    "README.md"
    "ARCHITECTURE.md"
    "API_GUIDE.md"
    "DEPLOYMENT.md"
    "SECURITY.md"
    "TEST.md"
    "CLAUDE.md"
    "TASK.md"
    "OVERVIEW.md"
    "backend/api/openapi.yaml"
)

# Check if each required file exists
echo -e "\n📋 Checking required documentation files..."
MISSING_FILES=0

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${GREEN}✓${NC} $doc exists"
    else
        echo -e "${RED}✗${NC} $doc is missing!"
        MISSING_FILES=$((MISSING_FILES + 1))
    fi
done

# Check if documentation is stale (not updated in last 30 days)
echo -e "\n📅 Checking documentation freshness..."
STALE_FILES=0

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        # Get last modified time in days
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            DAYS_OLD=$(( ($(date +%s) - $(stat -f %m "$doc")) / 86400 ))
        else
            # Linux
            DAYS_OLD=$(( ($(date +%s) - $(stat -c %Y "$doc")) / 86400 ))
        fi
        
        if [ $DAYS_OLD -gt 30 ]; then
            echo -e "${YELLOW}⚠${NC}  $doc hasn't been updated in $DAYS_OLD days"
            STALE_FILES=$((STALE_FILES + 1))
        fi
    fi
done

# Check for TODO items in documentation
echo -e "\n📝 Checking for pending TODOs in documentation..."
TODO_COUNT=0

for doc in "${REQUIRED_DOCS[@]}"; do
    if [ -f "$doc" ]; then
        TODOS=$(grep -i "TODO\|FIXME\|XXX\|\[.*to be added.*\]\|\*.*will be added.*\*" "$doc" 2>/dev/null | wc -l)
        if [ $TODOS -gt 0 ]; then
            echo -e "${YELLOW}⚠${NC}  $doc has $TODOS TODO items"
            TODO_COUNT=$((TODO_COUNT + TODOS))
        fi
    fi
done

# Check if API documentation matches OpenAPI spec
echo -e "\n🔍 Checking API documentation consistency..."
if [ -f "backend/api/openapi.yaml" ] && [ -f "API_GUIDE.md" ]; then
    # Count endpoints in OpenAPI spec
    OPENAPI_ENDPOINTS=$(grep -E "^ {2,4}/.+:$" backend/api/openapi.yaml | wc -l)
    # Count endpoint documentation in API guide (rough estimate)
    API_GUIDE_ENDPOINTS=$(grep -E "^(GET|POST|PUT|DELETE|PATCH) /" API_GUIDE.md | wc -l)
    
    if [ $OPENAPI_ENDPOINTS -ne $API_GUIDE_ENDPOINTS ]; then
        echo -e "${YELLOW}⚠${NC}  OpenAPI spec has $OPENAPI_ENDPOINTS endpoints, API_GUIDE has ~$API_GUIDE_ENDPOINTS documented"
    else
        echo -e "${GREEN}✓${NC} API documentation appears consistent"
    fi
fi

# Summary
echo -e "\n📊 Summary"
echo "=========="

if [ $MISSING_FILES -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All required documentation files exist"
else
    echo -e "${RED}✗${NC} $MISSING_FILES documentation files are missing"
fi

if [ $STALE_FILES -eq 0 ]; then
    echo -e "${GREEN}✓${NC} All documentation is reasonably fresh"
else
    echo -e "${YELLOW}⚠${NC}  $STALE_FILES files may need updating"
fi

if [ $TODO_COUNT -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No TODO items found in documentation"
else
    echo -e "${YELLOW}⚠${NC}  $TODO_COUNT TODO items need attention"
fi

# Exit with error if critical issues found
if [ $MISSING_FILES -gt 0 ]; then
    echo -e "\n${RED}❌ Documentation validation failed!${NC}"
    exit 1
else
    echo -e "\n${GREEN}✅ Documentation validation passed!${NC}"
    exit 0
fi