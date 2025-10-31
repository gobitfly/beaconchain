---
description: Manage your validators via API
---

# ⚒️ Manage Validators \[API]

To modify the validator dashboard via API, make sure to have an active [Orca subscription](https://beaconcha.in/premium).\
\
After creating a [validator dashboard](https://v2-beta-holesky.beaconcha.in/) and the desired [validator groups](validator-groups.md) through the User Interface, you can add and assign validators to groups using our API.

***

* For mainnet, please adjust the base URL to: [`beaconcha.in`](https://beaconcha.in/)
* The Group ID can be found in the `Group Manage` [modal ](validator-groups.md)on the Dashboard
* The Dashboard ID is visible in the Dashboard URL
* During our beta the **API key** will only be visible in the [account settings](https://beaconcha.in/user/settings#api) on [https://beaconcha.in/user/settings#api](https://beaconcha.in/user/settings#api)
* Pass the API key as a parameter `api_key`
  * `https://beaconcha.in/api/v2/validator-dashboards/{dashboard_id}/validators?api_key=KEY`    \


{% hint style="warning" %}
Please adjust the base URL from the examples below based on the network. \
\
Holesky: `holesky.beaconcha.in`\
Mainnet: `beaconcha.in` \
Gnosis: `gnosis.beaconcha.in`
{% endhint %}

{% openapi src="../.gitbook/assets/API_spec.json" path="/validator-dashboards/{dashboard_id}/validators" method="post" %}
[API_spec.json](../.gitbook/assets/API_spec.json)
{% endopenapi %}

{% openapi src="../.gitbook/assets/API_spec.json" path="/validator-dashboards/{dashboard_id}/groups/{group_id}/validators" method="delete" %}
[API_spec.json](../.gitbook/assets/API_spec.json)
{% endopenapi %}

{% openapi src="../.gitbook/assets/API_spec.json" path="/validator-dashboards/{dashboard_id}/validators/bulk-deletions" method="post" %}
[API_spec.json](../.gitbook/assets/API_spec.json)
{% endopenapi %}

