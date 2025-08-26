#!/bin/bash

echo "=== MetalLB Operator Debug Script ==="
echo

# Check current CSV status
echo "1. CSV Status:"
omc get csv -n openshift-metallb-operator
echo

# Check CSV details in YAML if status is not ok
CSV_STATUS=$(omc get csv -n openshift-metallb-operator -o jsonpath='{.items[0].status.phase}' 2>/dev/null)
if [ "$CSV_STATUS" != "Succeeded" ] && [ ! -z "$CSV_STATUS" ]; then
    echo "CSV YAML (status not ok):"
    omc get csv -n openshift-metallb-operator -o yaml
    echo
fi

# Check InstallPlan
echo "2. InstallPlan Status:"
omc get installplan -n openshift-metallb-operator
echo

# Get detailed InstallPlan info
INSTALLPLAN=$(omc get installplan -n openshift-metallb-operator -o jsonpath='{.items[0].metadata.name}')
if [ ! -z "$INSTALLPLAN" ]; then
    echo "InstallPlan Details:"
    omc describe installplan $INSTALLPLAN -n openshift-metallb-operator
    echo
fi

# Check Subscription
echo "3. Subscription Status:"
omc get subscription -n openshift-metallb-operator
echo

SUB=$(omc get subscription -n openshift-metallb-operator -o jsonpath='{.items[0].metadata.name}')
if [ ! -z "$SUB" ]; then
    echo "Subscription Details:"
    omc describe subscription $SUB -n openshift-metallb-operator
    echo
fi

# Check CatalogSource
echo "4. CatalogSource Status:"
omc get catalogsource -n openshift-metallb-operator
echo

# Check OLM logs
echo "5. OLM Operator Logs (last 20 lines):"
omc logs -n openshift-cluster-olm-operator deployment/olm-operator
echo

echo "6. Catalog Operator Logs (last 20 lines):"
omc logs -n openshift-cluster-olm-operator deployment/catalog-operator
echo

# Check events
echo "7. Recent Events:"
omc get events -n openshift-metallb-operator --sort-by='.lastTimestamp' | tail -10
echo

echo "=== Debug Complete ==="
