package modules

import (
	"fmt"
	"testing"
)

func TestGapGroup(t *testing.T) {
	groups := getEpochParallelGroups([]uint64{1, 2, 3, 6, 7, 9, 22, 33, 44, 45, 46, 47, 48, 49, 50}, epochFetchParallelism)

	for i, group := range groups {
		fmt.Printf("Group %d: %v\n", i, group)
		//t.Logf("Group %d: %v", i, group)
	}

	groups2 := getEpochParallelGroups([]uint64{1, 2, 3, 5, 7, 9, 10, 11, 12, 13, 20, 30}, 4)
	for i, group := range groups2 {
		fmt.Printf("Group2 %d: %v\n", i, group)
		//t.Logf("Group %d: %v", i, group)
	}

	groups3 := getEpochParallelGroups([]uint64{1238, 1239, 1240, 1241, 1242, 1243, 1244, 1245, 1246, 1247, 1248, 1249, 1250, 1251, 1252, 1253, 1254, 1255, 1256, 1257, 1258, 1259, 1260, 1261, 1262, 1263, 1264, 1265, 1266, 1267, 1268, 1269, 1270, 1271, 1272, 1273, 1274, 1275, 1276, 1277, 1278, 1279, 1280, 1281, 1282, 1283, 1284, 1285, 1286, 1287, 1288, 1289, 1290, 1291, 1292, 1293, 1294, 1295, 1296, 1297, 1298, 1299, 1300, 1301, 1302, 1303, 1304, 1305, 1306, 1307, 1308, 1309, 1310, 1311, 1312, 1313, 1314, 1315, 1316, 1317, 1318, 1319, 1320, 1321, 1322, 1323, 1324, 1325, 1326, 1327, 1328, 1329, 1330, 1331, 1332, 1333, 1334, 1335, 1336, 1337, 1338, 1339, 1340, 1341, 1342, 1343, 1344, 1345, 1346, 1347, 1348, 1349, 1350, 1351, 1352, 1353, 1354, 1355, 1356, 1357, 1358, 1359, 1360, 1361, 1362, 1363, 1364, 1365, 1366, 1367, 1368, 1369, 1370, 1371, 2949}, 6)
	for i, group := range groups3 {
		fmt.Printf("Group3 %d: %v\n", i, group)
		//t.Logf("Group %d: %v", i, group)
	}
}

func TestCompareData(t *testing.T) {
	/*
	   CREATE TABLE IF NOT EXISTS validator_dashboard_data_rolling_total (
	       validator_index int NOT NULL,
	       epoch_start int NOT NULL, -- incl
	       epoch_end int NOT NULL, -- excl
	       attestations_source_reward BIGINT,
	       attestations_target_reward BIGINT,
	       attestations_head_reward BIGINT,
	       attestations_inactivity_reward BIGINT,
	       attestations_inclusion_reward BIGINT,
	       attestations_reward BIGINT,
	       attestations_ideal_source_reward BIGINT,
	       attestations_ideal_target_reward BIGINT,
	       attestations_ideal_head_reward BIGINT,
	       attestations_ideal_inactivity_reward BIGINT,
	       attestations_ideal_inclusion_reward BIGINT,
	       attestations_ideal_reward BIGINT,
	       blocks_scheduled BIGINT,
	       blocks_proposed BIGINT,
	       blocks_cl_reward BIGINT, -- gwei
	       blocks_el_reward NUMERIC, -- wei
	       sync_scheduled BIGINT,
	       sync_executed BIGINT,
	       sync_rewards BIGINT,
	       slashed BOOLEAN,
	       balance_start BIGINT,
	       balance_end BIGINT,
	       deposits_count BIGINT,
	       deposits_amount BIGINT,
	       withdrawals_count BIGINT,
	       withdrawals_amount BIGINT,
	       inclusion_delay_sum BIGINT,
	       sync_chance double precision,
	       block_chance double precision,
	       attestations_scheduled int,
	       attestations_executed int,
	       attestation_head_executed int,
	       attestation_source_executed int,
	       attestation_target_executed int,
	       optimal_inclusion_delay_sum int,
	       primary key (validator_index)
	   );

	*/
	// sqlQuery :=
	// `
	// SELECT
	// 	validator_index,
	// 	sum(attestations_source_reward) as attestations_source_reward,
	// 	sum(attestations_target_reward) as attestations_target_reward,
	// 	sum(attestations_head_reward) as attestations_head_reward,
	// 	sum(attestations_inactivity_reward) as attestations_inactivity_reward,
	// 	sum(attestations_inclusion_reward) as attestations_inclusion_reward,
	// 	sum(attestations_reward) as attestations_reward,
	// 	sum(attestations_ideal_source_reward) as attestations_ideal_source_reward,
	// 	sum(attestations_ideal_target_reward) as attestations_ideal_target_reward,
	// 	sum(attestations_ideal_head_reward) as attestations_ideal_head_reward,
	// 	sum(attestations_ideal_inactivity_reward) as attestations_ideal_inactivity_reward,
	// 	sum(attestations_ideal_inclusion_reward) as attestations_ideal_inclusion_reward,
	// 	sum(attestations_ideal_reward) as attestations_ideal_reward,
	// 	sum(blocks_scheduled) as blocks_scheduled,
	// 	sum(blocks_proposed) as blocks_proposed,
	// 	sum(blocks_cl_reward) as blocks_cl_reward,
	// 	sum(blocks_el_reward) as blocks_el_reward,
	// 	sum(sync_scheduled) as sync_scheduled,
	// 	sum(sync_executed) as sync_executed,
	// 	sum(sync_rewards) as sync_rewards,
	// 	bool_and(slashed) as slashed,
	// 	sum(balance_start) as balance_start,
	// 	sum(balance_end) as balance_end,
	// 	sum(deposits_count) as deposits_count,
	// 	sum(deposits_amount) as deposits_amount,
	// 	sum(withdrawals_count) as withdrawals_count,
	// 	sum(withdrawals_amount) as withdrawals_amount,
	// 	sum(inclusion_delay_sum) as inclusion_delay_sum,

	// `
}
