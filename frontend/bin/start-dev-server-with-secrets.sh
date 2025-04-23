#!/bin/bash

if ! command -v gcloud >/dev/null 2>&1; then
  echo "⚠️  Google-Cloud-SDK (gcloud CLI) is not installed."
  echo "gcloud is needed to access google secret manager (e.g. to access SSR-Secret)"
  echo
  OS="$(uname)"
  hasHomebrew="$(command -v brew)"
  if [[  $OS == "Darwin" && $hasHomebrew  ]]; then
    echo "👉 Running:"
    echo "brew install --cask google-cloud-sdk"
    brew install --cask google-cloud-sdk
  else
    echo "Visit: https://cloud.google.com/sdk/docs/install"
  fi
  exit 1
fi

BEACONCHAIN_DEVELOPER_SSR_SECRET="$(gcloud secrets versions access latest --project etherchain --secret BEACONCHAIN_DEVELOPER_SSR_SECRET)" && \
NUXT_PRIVATE_SSR_SECRET=$BEACONCHAIN_DEVELOPER_SSR_SECRET nuxt dev