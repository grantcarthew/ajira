#!/bin/bash
# Show all ajira help output for review

set -e

AJIRA="${1:-./ajira}"

divider() {
    echo ""
    echo "========================================================================"
    echo "  $1"
    echo "========================================================================"
    echo ""
}

divider "ajira --help"
$AJIRA --help

divider "ajira help agents"
$AJIRA help agents

divider "ajira help schemas"
$AJIRA help schemas

divider "ajira help markdown"
$AJIRA help markdown

divider "ajira completion --help"
$AJIRA completion --help

divider "ajira me --help"
$AJIRA me --help

divider "ajira open --help"
$AJIRA open --help

divider "ajira project --help"
$AJIRA project --help

divider "ajira project list --help"
$AJIRA project list --help

divider "ajira board --help"
$AJIRA board --help

divider "ajira board list --help"
$AJIRA board list --help

divider "ajira release --help"
$AJIRA release --help

divider "ajira release list --help"
$AJIRA release list --help

divider "ajira user --help"
$AJIRA user --help

divider "ajira user search --help"
$AJIRA user search --help

divider "ajira field --help"
$AJIRA field --help

divider "ajira field list --help"
$AJIRA field list --help

divider "ajira issue --help"
$AJIRA issue --help

divider "ajira issue list --help"
$AJIRA issue list --help

divider "ajira issue view --help"
$AJIRA issue view --help

divider "ajira issue create --help"
$AJIRA issue create --help

divider "ajira issue edit --help"
$AJIRA issue edit --help

divider "ajira issue delete --help"
$AJIRA issue delete --help

divider "ajira issue clone --help"
$AJIRA issue clone --help

divider "ajira issue assign --help"
$AJIRA issue assign --help

divider "ajira issue move --help"
$AJIRA issue move --help

divider "ajira issue watch --help"
$AJIRA issue watch --help

divider "ajira issue unwatch --help"
$AJIRA issue unwatch --help

divider "ajira issue open --help"
$AJIRA issue open --help

divider "ajira issue comment --help"
$AJIRA issue comment --help

divider "ajira issue comment add --help"
$AJIRA issue comment add --help

divider "ajira issue comment edit --help"
$AJIRA issue comment edit --help

divider "ajira issue link --help"
$AJIRA issue link --help

divider "ajira issue link add --help"
$AJIRA issue link add --help

divider "ajira issue link remove --help"
$AJIRA issue link remove --help

divider "ajira issue link types --help"
$AJIRA issue link types --help

divider "ajira issue link url --help"
$AJIRA issue link url --help

divider "ajira issue type --help"
$AJIRA issue type --help

divider "ajira issue status --help"
$AJIRA issue status --help

divider "ajira issue priority --help"
$AJIRA issue priority --help

divider "ajira epic --help"
$AJIRA epic --help

divider "ajira epic list --help"
$AJIRA epic list --help

divider "ajira epic create --help"
$AJIRA epic create --help

divider "ajira epic add --help"
$AJIRA epic add --help

divider "ajira epic remove --help"
$AJIRA epic remove --help

divider "ajira sprint --help"
$AJIRA sprint --help

divider "ajira sprint list --help"
$AJIRA sprint list --help

divider "ajira sprint add --help"
$AJIRA sprint add --help

echo ""
echo "========================================================================"
echo "  Done - all help output displayed"
echo "========================================================================"
