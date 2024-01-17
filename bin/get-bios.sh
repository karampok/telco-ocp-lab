#!/usr/bin/env bash
set -euo pipefail

BMC_HOST=https://${1:-"192.168.18.183"}

cargs=(-sLk -H "OData-Version: 4.0" -H "Content-Type: application/json; charset=utf-8" -u "$AUTH")

function get_bios_attributes {
  bios_attr_uri=$(
    curl "${cargs[@]}" "$BMC_HOST"/redfish/v1/Registries/ | jq -r '.Members[]."@odata.id" | match("(/.*BiosAttribute.*)").string'
  )
  
  bios_attr_jsonschema_uri=$(
    curl "${cargs[@]}" "$BMC_HOST$bios_attr_uri" | jq -r '.Location[] | select(."Language" == "en" or ."Language" == "en-US")."Uri"'
  )

  bios_attr_tmp=$(mktemp -t bios_attr.XXX)
  curl "${cargs[@]}" "$BMC_HOST$bios_attr_jsonschema_uri" > "$bios_attr_tmp"
  c="cat"
  if ! [[ "$(file "$bios_attr_tmp" | grep ':.*gzip compressed data')" == "" ]]; then
    c="zcat"
  fi
  $c "$bios_attr_tmp" | jq -r '.SupportedSystems'

  system_uri=$(
    curl "${cargs[@]}" "$BMC_HOST"/redfish/v1/Systems/ | jq -r '.Members[0]."@odata.id"'
  )

  bios_attr_uri=$(
    curl "${cargs[@]}" "$BMC_HOST$system_uri" | jq -r '.Bios."@odata.id"'
  )

  bios_values_tmp=$(mktemp -t bios_values.XXX)
  curl "${cargs[@]}" "$BMC_HOST$bios_attr_uri"> "$bios_values_tmp"

  for a in "$@"
  do
     $c "$bios_attr_tmp" | jq --arg a "$a" '.RegistryEntries.Attributes[] | select(."AttributeName" == $a)'
     jq --arg a "$a" '.Attributes |  {($a): .[$a]}' "$bios_values_tmp"
  done
  echo "$c $bios_attr_tmp | jq ."
  echo "cat $bios_values_tmp |jq ."
}

attrs=(
# AdvancedMemProtection
# BootMode
# CollabPowerControl
EnergyEfficientTurbo
#EnergyPerfBias
# EnhancedProcPerf
# MemPatrolScrubbing
 MinProcIdlePkgState
 MinProcIdlePower
# NumaGroupSizeOpt
PowerRegulator
ProcTurbo
# ProcHyperthreading
# ProcessorPhysicalAddress
ThermalConfig
# UncoreFreqScaling
# WorkloadProfile
)


get_bios_attributes "${attrs[@]}"
