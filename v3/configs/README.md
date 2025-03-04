# /configs

Configurations are hierarchical. So for example, given 2 configuration sets `chain/default.yaml`and `chain/mainnet.yaml`, we take the set of all configurations, and if a Config property is defined in both `default` and `mainnet`, we take the one that is defined higher in the hierarchy (i.e. `mainnet`).

We have 2 kinds of configurations:

* Chain Config - Defines the properties for the chain itself. Uses `chain/default.yaml` as its base, with `{chain_name}.yaml` (i.e. chain_name=[`mainnet`, `holesky`]) overwriting.
* Service Config - Defines the properties of the service and the dependencies it should run with. Uses `service/default.yamlz` as its base, with `{env_name}.yaml` (i.e. env_name=[`development`, `staging`, `production`]) overwriting

Generally the configurations follow in order this overwrite hierarchy:
1. `default.yaml`
2. `chain/env.yaml`
3. Environmental Variables

So environment variables have the final say in the value of the config properties. Therefore, to avoid saving sensitive information in plain text (which might get accidentally committed), NEVER save sensitive properties here
and instead specify them as Environment Variables. Sensitive variables include Usernames, Passwords, Internal Endpoints/Hosts, API Keys, etc.