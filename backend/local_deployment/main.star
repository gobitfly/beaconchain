input_parser = import_module("github.com/ethpandaops/ethereum-package/src/package_io/input_parser.star")
eth_network_module = import_module("github.com/ethpandaops/ethereum-package/main.star")
transaction_spammer = import_module("github.com/ethpandaops/ethereum-package/src/transaction_spammer/transaction_spammer.star")
blob_spammer = import_module("github.com/ethpandaops/ethereum-package/src/blob_spammer/blob_spammer.star")
genesis_constants = import_module("github.com/ethpandaops/ethereum-package/src/prelaunch_data_generator/genesis_constants/genesis_constants.star")
shared_utils = import_module("github.com/ethpandaops/ethereum-package/src/shared_utils/shared_utils.star")

POSTGRES_DB = "db"
ALLOY_PORT_ID = "alloy"
ALLOY_DB = "alloy"
CLICKHOUSE_PORT_ID = "clickhouse"
CLICKHOUSE_DB = "clickhouse"
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "pass"

REDIS_PORT_ID = "redis"

LITTLE_BIGTABLE_PORT_ID = "littlebigtable"

EXPLORER_CONFIG_FILENAME = "config.yml"

def run(plan, args):
	args_with_right_defaults = input_parser.input_parser(plan, args)
	network_params = args_with_right_defaults.network_params

	db_services = plan.add_services(
		configs={
			# Add Postgres server to simulate alloy; consider switching to omni (https://cloud.google.com/alloydb/docs/omni)
			"alloy": ServiceConfig(
				image = "postgres:15.2-alpine",
				ports = {
					ALLOY_PORT_ID: PortSpec(5432, application_protocol = "postgresql"),
				},
				env_vars = {
					"POSTGRES_DB": ALLOY_DB,
					"POSTGRES_USER": POSTGRES_USER,
					"POSTGRES_PASSWORD": POSTGRES_PASSWORD,
				},
			),
			# Add a Clickhouse server
			"clickhouse": ServiceConfig(
				image = "clickhouse/clickhouse-server:24.5",
				ports = {
					CLICKHOUSE_PORT_ID: PortSpec(9000, application_protocol = "clickhouse"),
					"http": PortSpec(8123),
				},
				env_vars = {
					"CLICKHOUSE_DB": CLICKHOUSE_DB,
					"CLICKHOUSE_USER": POSTGRES_USER,
					"CLICKHOUSE_PASSWORD": POSTGRES_PASSWORD,
				},
			),
			# Add a Redis server
			"redis": ServiceConfig(
				image = "redis:7",
				ports = {
					REDIS_PORT_ID: PortSpec(6379, application_protocol = "tcp"),
				},
			),
			# Add a Bigtable Emulator server
			"littlebigtable": ServiceConfig(
				image = "gobitfly/little_bigtable:latest",
				ports = {
					LITTLE_BIGTABLE_PORT_ID: PortSpec(9000, application_protocol = "tcp"),
				},
			),
		}
	)

	# Spin up a local ethereum testnet
	eth_network_module.run(plan, args)


def new_config_template_data(cl_node_info, el_uri, lbt_host, lbt_port, db_host, db_port, alloy_port, redis_uri):
	return {
		"CLNodeHost": cl_node_info.ip_addr,
		"CLNodePort": cl_node_info.http_port_num,
		"ELNodeEndpoint": el_uri,
		"LBTHost": lbt_host,
		"LBTPort": lbt_port,
		"DBHost": db_host,
		"DBPort": db_port,
		"AlloyPort": alloy_port,
		"RedisEndpoint": redis_uri,
	}
