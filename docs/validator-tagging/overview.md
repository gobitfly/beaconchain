---
title: Validator Tagging
slug: validator-tagging
published: true
description: >-
  An in-depth explainer of the validator tagging module powering entity
  attribution on beaconcha.in, including data sources, on-chain lookups, whale
  clustering logic, schedules, and FAQs.
icon: tag
---

# Overview

#### Summary

This article explains how [beaconcha.in](broken-reference) assigns human‑readable tags to Ethereum validators. Tags make it easier to understand which validators belong to well‑known entities (e.g., Lido, Rocket Pool), as well as large independent operators ("whales"). It covers data sources, the step‑by‑step tagging pipeline, how often tagging runs, and how conflicts are resolved.

Shout‑out: A big thank‑you to [Hildobby](https://x.com/hildobby) for maintaining a community dataset of validator/entity mappings used by many in the ecosystem.

***

#### What is a validator tag?

A validator tag is a short label associated with one or more validators. Examples:

* `Coinbase` or `Binance`
* `Lido` or `Lido (SimpleDVT)` (Lido modules)
* `Rocket Pool` (Rocket Pool validators)
* `Whale_0x1234` (an independent cluster labeled by address prefix)

Tags help you:

* Attribute validator performance to entities via their BeaconScore benchmark value
* Understand decentralization by operator/entity
* Identify large clusters of validators controlled by the same address

***

#### High‑level pipeline

At a high level, the tagging module:

1. Imports community tags from the [Hildobby](https://x.com/hildobby) dataset
2. Queries on‑chain contracts to detect Lido validators (Curated, Simple DVT, Community Staking)
3. Queries on‑chain contracts to detect Rocket Pool validators
4. For any remaining untagged validators, clusters by withdrawal address and assigns whale tags if the cluster balance exceeds 320 ETH
5. For any remaining untagged validators, clusters by deposit (funding) address and assigns whale tags if the cluster balance exceeds 320 ETH

The pipeline runs automatically once a day.

***

#### Step‑by‑step details

**1) Community tags via Hildobby**

* Source: [Hildobby’s](https://x.com/hildobby) public validator/entity dataset (community maintained).
* Matching: The dataset contains mappings of validator indices or pubkeys to known entities. We import those tags directly.
* Why first? Community‑maintained mappings often include entities that are not easily inferred from on‑chain structures.
* Note: We retain Hildobby’s naming semantics and apply them as the baseline when no higher‑certainty on‑chain inference is available.

**2) Lido validators from on‑chain contracts**

We identify validators belonging to Lido by reading their on‑chain registry data. This includes:

* `Lido Curated Module`
* `Lido SimpleDVT Module`
* `Lido Community Staking Module`

Implementation outline:

* Enumerate the validator set exposed by each module’s contracts
* Map each validator pubkey to the appropriate tag, typically `Lido`, `Lido (SimpleDVT)`, or `Lido (CSM)`
* Prefer these determinations over community tags when a contract says a validator belongs to Lido

Rationale: On‑chain state is authoritative for Lido’s modules and provides the highest‑certainty attribution.

**3) Rocket Pool validators from on‑chain contracts**

We detect Rocket Pool validators by reading the relevant Rocket Pool contracts that register or imply validator ownership (e.g., minipools and their validator keys/addresses).

Implementation outline:

* Enumerate active and historical minipools
* Retrieve associated validator pubkeys
* Tag those validators as `RocketPool`

Rationale: Like Lido, Rocket Pool’s on‑chain structures allow for accurate, direct inference of validator membership.

**4) Whale tagging by withdrawal address (> 320 ETH)**

For remaining untagged validators, we infer likely common control by clustering validators sharing the same withdrawal credentials/address.

* Cluster key: withdrawal address
* Cluster size metric: total effective balance of validators in the cluster
* Threshold: `> 320 ETH` (equivalent to more than 10 full validators at 32 ETH each)
* Tag format: `Whale_0x1234` where `0x1234` is a short prefix of the withdrawal address for readability

Operators often consolidate withdrawals to a single address, making it a strong signal of common ownership/control.

**5) Whale tagging by deposit (funding) address (> 320 ETH)**

For any still‑untagged validators, we cluster by the funding address that sent the deposit transaction(s).

* Cluster key: deposit/funding (from) address
* Cluster size metric: total effective balance of validators funded by that address
* Threshold: `> 320 ETH`
* Tag format: `Whale_0x1234`

Large operators often fund many deposits from a single address, even if they diversify withdrawal addresses.

***

#### Tag precedence and conflict resolution

When a validator qualifies for multiple tags, we follow this precedence (from highest to lowest certainty):

1. On‑chain entity detection (Lido modules, Rocket Pool)
2. Community dataset (Hildobby)
3. Inferred whale clustering by withdrawal address
4. Inferred whale clustering by deposit address

***

#### Schedule and freshness

The tagging system runs on a fixed UTC schedule:

* Daily at 10:00 UTC: Full tagging pipeline (all steps) runs in order, followed by a precompute job that updates the entity data dashboards.
* Hourly (other than 10:00): Precompute only. This ensures UI remains fast and consistent even between daily runs.

***

#### Examples

* A validator included in Lido’s Curated module contracts will be tagged `Lido`.
* An independent operator running 15 validators with the same withdrawal address will be tagged `Whale_0xABCD` where `0xABCD` is a short prefix of their withdrawal address.
* A cluster of 8 validators funded from address `0xF...` will not get a whale tag because the balance threshold (`> 320 ETH`) is not met.

***

#### How to request a correction

If you believe a tag is incorrect:

* Open an issue in the beaconcha.in [support channel](https://dsc.gg/beaconchain) and include validator indices or pubkeys, plus the correct attribution if known.
* For on‑chain attributions (Lido/Rocket Pool), please provide the relevant contract references or transaction links.

Corrections propagate after the next daily run.

***

#### Frequently asked questions

* Why does a whale tag use a short address prefix? For readability in the UI. The underlying full address is retained for accuracy.
* Will small operators get whale tags? No. The threshold is strict: only clusters with total balance `> 320 ETH` qualify.
* When are tags assigned for new validators? Tags refresh daily with new data and may evolve if on‑chain ownership changes or the community dataset is updated.
* Does tagging affect rewards or protocol behavior? No. Tagging is a UI attribution feature only; it does not interact with consensus or execution layer incentives.

***

#### Contact

If you have questions or suggestions about validator tagging, please reach out via our [support channel](https://dsc.gg/beaconchain).
