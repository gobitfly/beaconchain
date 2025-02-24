# /configs
Configuration file templates or default configs.

Put your confd or consul-template template files here.

The service configuration can be thought of as the aggregate of the configurations of various components:

* The development `Environment` the service is running in (Development,Staging,Production)\
* The `Chain` it is running against

For each tuple of `Environment` and `Chain`, there should be a configuration file which describes the properties it should run with
For example, if we are running the mainnet chain, we probably dont need to re-define the environment variables for ChainId for each of the different environments. We only need to do it once.
However, there may be a situation in which a developer wants to run against a "fake" mainnet chain, and provide overrides 
