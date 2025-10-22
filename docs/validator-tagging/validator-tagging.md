---
title: Validator Tagging
description: An in-depth explainer of the validator tagging module powering entity attribution on beaconcha.in, including data sources, on-chain lookups, whale clustering logic, schedules, and FAQs.
slug: validator-tagging
published: true
---

### Summary
This article explains how beaconcha.in assigns human‑readable tags to Ethereum validators. Tags make it easier to understand which validators belong to well‑known entities (e.g., Lido, Rocket Pool), as well as large independent operators ("whales"). It covers data sources, the step‑by‑step tagging pipeline, how often tagging runs, and how conflicts are resolved.

Shout‑out: A big thank‑you to Hildobby for maintaining a community dataset of validator/entity mappings used by many in the ecosystem. Learn more at https://x.com/hildobby.

---

### Who is this for?
- Users who want to understand attribution on beaconcha.in
- Node operators looking to see how their validators get tagged
- Researchers and analysts reviewing entity-level metrics

---

### What is a validator tag?
A validator tag is a short label associated with one or more validators. Examples:
- `Coinbase` or `Binance`
- `Lido_Curated` or `Lido_SimpleDV` (Lido modules)
- `RocketPool` (Rocket Pool validators)
- `Whale_0x1234` (an independent cluster labeled by address prefix)

Tags help you:
- Attribute validator performance to entities via their BeaconScore benachmark value
- Understand decentralization by operator/entity
- Identify large clusters of validators controlled by the same address

---

### High‑level pipeline
At a high level, the tagging module:
1) Imports community tags from the Hildobby dataset
2) Queries on‑chain contracts to detect Lido validators (Curated, Simple DVT, Community Staking)
3) Queries on‑chain contracts to detect Rocket Pool validators
4) For any remaining untagged validators, clusters by withdrawal address and assigns whale tags if the cluster balance exceeds 320 ETH
5) For any remaining untagged validators, clusters by deposit (funding) address and assigns whale tags if the cluster balance exceeds 320 ETH

The pipeline runs automatically once a day.

---

### Visual overview
```mermaid
flowchart TD
  A[Start] --> B[Import Hildobby dataset\n(community tags)]
  B --> C[Query on-chain: Lido modules\n- Curated Module\n- Simple DVT Module\n- Community Staking]
  C --> D[Query on-chain: Rocket Pool validators]
  D --> E[Cluster remaining by withdrawal address\nAssign Whale_0x1234 if balance > 320 ETH]
  E --> F[Cluster remaining by deposit address\nAssign Whale_0x1234 if balance > 320 ETH]
  F --> G[Finalize + Precompute\n(caches/indexes)]
  style B fill:#e9f5ff,stroke:#4da3ff
  style C fill:#efffec,stroke:#43a047
  style D fill:#efffec,stroke:#43a047
  style E fill:#fff7e6,stroke:#f0a500
  style F fill:#fff7e6,stroke:#f0a500
  style G fill:#f2f2f2,stroke:#9e9e9e
```

---

### Step‑by‑step details

#### 1) Community tags via Hildobby
- Source: Hildobby’s public validator/entity dataset (community maintained).
- Matching: The dataset contains mappings of validator indices or pubkeys to known entities. We import those tags directly.
- Why first? Community‑maintained mappings often include entities that are not easily inferred from on‑chain structures.
- Note: We retain Hildobby’s naming semantics and apply them as the baseline when no higher‑certainty on‑chain inference is available.

#### 2) Lido validators from on‑chain contracts
We identify validators belonging to Lido by reading their on‑chain registry data. This includes:
- `Lido Curated Module`
- `Lido SimpleDVT Module`
- `Lido Community Staking Module`

Implementation outline:
- Enumerate the validator set exposed by each module’s contracts
- Map each validator pubkey to the appropriate tag, typically `Lido_Curated`, `Lido_SimpleDV`, or `Lido_Community`
- Prefer these determinations over community tags when a contract says a validator belongs to Lido

Rationale: On‑chain state is authoritative for Lido’s modules and provides the highest‑certainty attribution.

#### 3) Rocket Pool validators from on‑chain contracts
We detect Rocket Pool validators by reading the relevant Rocket Pool contracts that register or imply validator ownership (e.g., minipools and their validator keys/addresses).

Implementation outline:
- Enumerate active and historical minipools
- Retrieve associated validator pubkeys
- Tag those validators as `RocketPool`

Rationale: Like Lido, Rocket Pool’s on‑chain structures allow for accurate, direct inference of validator membership.

#### 4) Whale tagging by withdrawal address (> 320 ETH)
For remaining untagged validators, we infer likely common control by clustering validators sharing the same withdrawal credentials/address.

- Cluster key: withdrawal address
- Cluster size metric: total effective balance of validators in the cluster
- Threshold: `> 320 ETH` (equivalent to more than 10 full validators at 32 ETH each)
- Tag format: `Whale_0x1234` where `0x1234` is a short prefix of the withdrawal address for readability

Why this works: Operators often consolidate withdrawals to a single address, making it a strong signal of common ownership/control.

#### 5) Whale tagging by deposit (funding) address (> 320 ETH)
For any still‑untagged validators, we cluster by the funding address that sent the deposit transaction(s).

- Cluster key: deposit/funding (from) address
- Cluster size metric: total effective balance of validators funded by that address
- Threshold: `> 320 ETH`
- Tag format: `Whale_0x1234`

Why this works: Large operators often fund many deposits from a single address, even if they diversify withdrawal addresses.

---

### Tag precedence and conflict resolution
When a validator qualifies for multiple tags, we follow this precedence (from highest to lowest certainty):
1) On‑chain entity detection (Lido modules, Rocket Pool)
2) Community dataset (Hildobby)
3) Inferred whale clustering by withdrawal address
4) Inferred whale clustering by deposit address

---

### Schedule and freshness
The tagging system runs on a fixed UTC schedule:
- Daily at 10:00 UTC: Full tagging pipeline (all steps) runs in order, followed by a precompute job that updates caches and indexes.
- Hourly (other than 10:00): Precompute only. This ensures UI remains fast and consistent even between daily runs.

What to expect:
- New tags from Hildobby, new on‑chain Lido/Rocket Pool registrations, and new whale clusters will appear after the next daily run.
- Derived metrics and dashboards refresh hourly via the precompute task.

---

### Data quality, safety, and limitations
- Authoritative sources: On‑chain contract reads for Lido/Rocket Pool are authoritative and take precedence.
- Community data caveat: Community mappings can contain occasional errors; corrections will be reflected after the next daily run.
- Heuristic whale tags: Clustering by withdrawal/deposit addresses is a heuristic. It is designed to surface large independent operators but is not a legal identity claim.
- Threshold choice: `> 320 ETH` reflects approximately 10 validators. This helps highlight truly large clusters while minimizing noise.

---

### Examples
- A validator included in Lido’s Curated module contracts will be tagged `Lido`.
- An independent operator running 15 validators with the same withdrawal address will be tagged `Whale_0xABCD` where `0xABCD` is a short prefix of their withdrawal address.
- A cluster of 8 validators funded from address `0xF...` will not get a whale tag because the balance threshold (`> 320 ETH`) is not met.

---

### How to request a correction
If you believe a tag is incorrect:
- Open an issue in the beaconcha.in [support channel](https://dsc.gg/beaconchain) and include validator indices or pubkeys, plus the correct attribution if known.
- For on‑chain attributions (Lido/Rocket Pool), please provide the relevant contract references or transaction links.

Corrections propagate after the next daily run.

---

### Frequently asked questions
- Why does a whale tag use a short address prefix? For readability in the UI. The underlying full address is retained for accuracy.
- Will small operators get whale tags? No. The threshold is strict: only clusters with total balance `> 320 ETH` qualify.
- When are tags assigned for new validators? Tags refresh daily with new data and may evolve if on‑chain ownership changes or the community dataset is updated.
- Does tagging affect rewards or protocol behavior? No. Tagging is a UI attribution feature only; it does not interact with consensus or execution layer incentives.

---

### Contact
If you have questions or suggestions about validator tagging, please reach out via our [support channel](https://dsc.gg/beaconchain).