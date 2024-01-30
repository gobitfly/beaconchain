package clnode

type GetSyncCommitteeAssignmentsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		Validators          []string   `json:"validators"`
		ValidatorAggregates [][]string `json:"validator_aggregates"`
	} `json:"data"`
}

type GetAttestationRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		IdealRewards []struct {
			EffectiveBalance string `json:"effective_balance"`
			Head             string `json:"head"`
			Target           string `json:"target"`
			Source           string `json:"source"`
			InclusionDelay   string `json:"inclusion_delay"`
			Inactivity       string `json:"inactivity"`
		} `json:"ideal_rewards"`
		TotalRewards []struct {
			ValidatorIndex string `json:"validator_index"`
			Head           string `json:"head"`
			Target         string `json:"target"`
			Source         string `json:"source"`
			InclusionDelay string `json:"inclusion_delay"`
			Inactivity     string `json:"inactivity"`
		} `json:"total_rewards"`
	} `json:"data"`
}

type GetSyncCommitteeRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		ValidatorIndex string `json:"validator_index"`
		Reward         string `json:"reward"`
	} `json:"data"`
}

type GetBlockRewardsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                struct {
		ProposerIndex     string `json:"proposer_index"`
		Total             string `json:"total"`
		Attestations      string `json:"attestations"`
		SyncAggregate     string `json:"sync_aggregate"`
		ProposerSlashings string `json:"proposer_slashings"`
		AttesterSlashings string `json:"attester_slashings"`
	} `json:"data"`
}

type GetValidatorsResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		Index     string `json:"index"`
		Balance   string `json:"balance"`
		Status    string `json:"status"`
		Validator struct {
			Pubkey                     string `json:"pubkey"`
			WithdrawalCredentials      string `json:"withdrawal_credentials"`
			EffectiveBalance           string `json:"effective_balance"`
			Slashed                    bool   `json:"slashed"`
			ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
			ActivationEpoch            string `json:"activation_epoch"`
			ExitEpoch                  string `json:"exit_epoch"`
			WithdrawableEpoch          string `json:"withdrawable_epoch"`
		} `json:"validator"`
	} `json:"data"`
}

type GetProposerAssignmentsResponse struct {
	DependentRoot       string `json:"dependent_root"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
	Data                []struct {
		Pubkey         string `json:"pubkey"`
		ValidatorIndex string `json:"validator_index"`
		Slot           string `json:"slot"`
	} `json:"data"`
}

type GetBeaconSlotResponse struct {
	Version             string `json:"version"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
	Finalized           bool   `json:"finalized"`
	Data                struct {
		Message struct {
			Slot          string `json:"slot"`
			ProposerIndex string `json:"proposer_index"`
			ParentRoot    string `json:"parent_root"`
			StateRoot     string `json:"state_root"`
			Body          struct {
				RandaoReveal string `json:"randao_reveal"`
				Eth1Data     struct {
					DepositRoot  string `json:"deposit_root"`
					DepositCount string `json:"deposit_count"`
					BlockHash    string `json:"block_hash"`
				} `json:"eth1_data"`
				Graffiti          string `json:"graffiti"`
				ProposerSlashings []struct {
					SignedHeader1 struct {
						Message struct {
							Slot          string `json:"slot"`
							ProposerIndex string `json:"proposer_index"`
							ParentRoot    string `json:"parent_root"`
							StateRoot     string `json:"state_root"`
							BodyRoot      string `json:"body_root"`
						} `json:"message"`
						Signature string `json:"signature"`
					} `json:"signed_header_1"`
					SignedHeader2 struct {
						Message struct {
							Slot          string `json:"slot"`
							ProposerIndex string `json:"proposer_index"`
							ParentRoot    string `json:"parent_root"`
							StateRoot     string `json:"state_root"`
							BodyRoot      string `json:"body_root"`
						} `json:"message"`
						Signature string `json:"signature"`
					} `json:"signed_header_2"`
				} `json:"proposer_slashings"`
				AttesterSlashings []struct {
					Attestation1 struct {
						AttestingIndices []string `json:"attesting_indices"`
						Signature        string   `json:"signature"`
						Data             struct {
							Slot            string `json:"slot"`
							Index           string `json:"index"`
							BeaconBlockRoot string `json:"beacon_block_root"`
							Source          struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"source"`
							Target struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"target"`
						} `json:"data"`
					} `json:"attestation_1"`
					Attestation2 struct {
						AttestingIndices []string `json:"attesting_indices"`
						Signature        string   `json:"signature"`
						Data             struct {
							Slot            string `json:"slot"`
							Index           string `json:"index"`
							BeaconBlockRoot string `json:"beacon_block_root"`
							Source          struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"source"`
							Target struct {
								Epoch string `json:"epoch"`
								Root  string `json:"root"`
							} `json:"target"`
						} `json:"data"`
					} `json:"attestation_2"`
				} `json:"attester_slashings"`
				Attestations []struct {
					AggregationBits string `json:"aggregation_bits"`
					Signature       string `json:"signature"`
					Data            struct {
						Slot            string `json:"slot"`
						Index           string `json:"index"`
						BeaconBlockRoot string `json:"beacon_block_root"`
						Source          struct {
							Epoch string `json:"epoch"`
							Root  string `json:"root"`
						} `json:"source"`
						Target struct {
							Epoch string `json:"epoch"`
							Root  string `json:"root"`
						} `json:"target"`
					} `json:"data"`
				} `json:"attestations"`
				Deposits []struct {
					Proof []string `json:"proof"`
					Data  struct {
						Pubkey                string `json:"pubkey"`
						WithdrawalCredentials string `json:"withdrawal_credentials"`
						Amount                string `json:"amount"`
						Signature             string `json:"signature"`
					} `json:"data"`
				} `json:"deposits"`
				VoluntaryExits []struct {
					Message struct {
						Epoch          string `json:"epoch"`
						ValidatorIndex string `json:"validator_index"`
					} `json:"message"`
					Signature string `json:"signature"`
				} `json:"voluntary_exits"`
				SyncAggregate struct {
					SyncCommitteeBits      string `json:"sync_committee_bits"`
					SyncCommitteeSignature string `json:"sync_committee_signature"`
				} `json:"sync_aggregate"`
				ExecutionPayload struct {
					ParentHash    string   `json:"parent_hash"`
					FeeRecipient  string   `json:"fee_recipient"`
					StateRoot     string   `json:"state_root"`
					ReceiptsRoot  string   `json:"receipts_root"`
					LogsBloom     string   `json:"logs_bloom"`
					PrevRandao    string   `json:"prev_randao"`
					BlockNumber   string   `json:"block_number"`
					GasLimit      string   `json:"gas_limit"`
					GasUsed       string   `json:"gas_used"`
					Timestamp     string   `json:"timestamp"`
					ExtraData     string   `json:"extra_data"`
					BaseFeePerGas string   `json:"base_fee_per_gas"`
					BlockHash     string   `json:"block_hash"`
					Transactions  []string `json:"transactions"`
					Withdrawals   []struct {
						Index          string `json:"index"`
						ValidatorIndex string `json:"validator_index"`
						Address        string `json:"address"`
						Amount         string `json:"amount"`
					} `json:"withdrawals"`
				} `json:"execution_payload"`
				BlsToExecutionChanges []any `json:"bls_to_execution_changes"`
			} `json:"body"`
		} `json:"message"`
		Signature string `json:"signature"`
	} `json:"data"`
}

type GetSpecResponse struct {
	Data struct {
		ConfigName                              string `json:"CONFIG_NAME"`
		PresetBase                              string `json:"PRESET_BASE"`
		TerminalTotalDifficulty                 string `json:"TERMINAL_TOTAL_DIFFICULTY"`
		TerminalBlockHash                       string `json:"TERMINAL_BLOCK_HASH"`
		TerminalBlockHashActivationEpoch        string `json:"TERMINAL_BLOCK_HASH_ACTIVATION_EPOCH"`
		SafeSlotsToImportOptimistically         int64  `json:"SAFE_SLOTS_TO_IMPORT_OPTIMISTICALLY,string"`
		MinGenesisActiveValidatorCount          int64  `json:"MIN_GENESIS_ACTIVE_VALIDATOR_COUNT,string"`
		MinGenesisTime                          int64  `json:"MIN_GENESIS_TIME,string"`
		GenesisForkVersion                      string `json:"GENESIS_FORK_VERSION"`
		GenesisDelay                            int64  `json:"GENESIS_DELAY,string"`
		AltairForkVersion                       string `json:"ALTAIR_FORK_VERSION"`
		AltairForkEpoch                         int64  `json:"ALTAIR_FORK_EPOCH,string"`
		BellatrixForkVersion                    string `json:"BELLATRIX_FORK_VERSION"`
		BellatrixForkEpoch                      int64  `json:"BELLATRIX_FORK_EPOCH,string"`
		CapellaForkVersion                      string `json:"CAPELLA_FORK_VERSION"`
		CapellaForkEpoch                        int64  `json:"CAPELLA_FORK_EPOCH,string"`
		SecondsPerSlot                          int64  `json:"SECONDS_PER_SLOT,string"`
		SecondsPerEth1Block                     int64  `json:"SECONDS_PER_ETH1_BLOCK,string"`
		MinValidatorWithdrawabilityDelay        int64  `json:"MIN_VALIDATOR_WITHDRAWABILITY_DELAY,string"`
		ShardCommitteePeriod                    int64  `json:"SHARD_COMMITTEE_PERIOD,string"`
		Eth1FollowDistance                      int64  `json:"ETH1_FOLLOW_DISTANCE,string"`
		SubnetsPerNode                          int64  `json:"SUBNETS_PER_NODE,string"`
		InactivityScoreBias                     int64  `json:"INACTIVITY_SCORE_BIAS,string"`
		InactivityScoreRecoveryRate             int64  `json:"INACTIVITY_SCORE_RECOVERY_RATE,string"`
		EjectionBalance                         int64  `json:"EJECTION_BALANCE,string"`
		MinPerEpochChurnLimit                   int64  `json:"MIN_PER_EPOCH_CHURN_LIMIT,string"`
		ChurnLimitQuotient                      int64  `json:"CHURN_LIMIT_QUOTIENT,string"`
		ProposerScoreBoost                      int64  `json:"PROPOSER_SCORE_BOOST,string"`
		DepositChainID                          int64  `json:"DEPOSIT_CHAIN_ID,string"`
		DepositNetworkID                        int64  `json:"DEPOSIT_NETWORK_ID,string"`
		DepositContractAddress                  string `json:"DEPOSIT_CONTRACT_ADDRESS"`
		GossipMaxSize                           int64  `json:"GOSSIP_MAX_SIZE,string"`
		MinEpochsForBlockRequests               int64  `json:"MIN_EPOCHS_FOR_BLOCK_REQUESTS,string"`
		MaxChunkSize                            int64  `json:"MAX_CHUNK_SIZE,string"`
		TTFBTimeout                             int64  `json:"TTFB_TIMEOUT,string"`
		RespTimeout                             int64  `json:"RESP_TIMEOUT,string"`
		MessageDomainInvalidSnappy              string `json:"MESSAGE_DOMAIN_INVALID_SNAPPY"`
		MessageDomainValidSnappy                string `json:"MESSAGE_DOMAIN_VALID_SNAPPY"`
		AttestationSubnetExtraBits              int64  `json:"ATTESTATION_SUBNET_EXTRA_BITS,string"`
		AttestationSubnetPrefixBits             int64  `json:"ATTESTATION_SUBNET_PREFIX_BITS,string"`
		MaxCommitteesPerSlot                    int64  `json:"MAX_COMMITTEES_PER_SLOT,string"`
		TargetCommitteeSize                     int64  `json:"TARGET_COMMITTEE_SIZE,string"`
		MaxValidatorsPerCommittee               int64  `json:"MAX_VALIDATORS_PER_COMMITTEE,string"`
		ShuffleRoundCount                       int64  `json:"SHUFFLE_ROUND_COUNT,string"`
		HysteresisQuotient                      int64  `json:"HYSTERESIS_QUOTIENT,string"`
		HysteresisDownwardMultiplier            int64  `json:"HYSTERESIS_DOWNWARD_MULTIPLIER,string"`
		HysteresisUpwardMultiplier              int64  `json:"HYSTERESIS_UPWARD_MULTIPLIER,string"`
		SafeSlotsToUpdateJustified              int64  `json:"SAFE_SLOTS_TO_UPDATE_JUSTIFIED,string"`
		MinDepositAmount                        int64  `json:"MIN_DEPOSIT_AMOUNT,string"`
		MaxEffectiveBalance                     int64  `json:"MAX_EFFECTIVE_BALANCE,string"`
		EffectiveBalanceIncrement               int64  `json:"EFFECTIVE_BALANCE_INCREMENT,string"`
		MinAttestationInclusionDelay            int64  `json:"MIN_ATTESTATION_INCLUSION_DELAY,string"`
		SlotsPerEpoch                           int64  `json:"SLOTS_PER_EPOCH,string"`
		MinSeedLookahead                        int64  `json:"MIN_SEED_LOOKAHEAD,string"`
		MaxSeedLookahead                        int64  `json:"MAX_SEED_LOOKAHEAD,string"`
		EpochsPerEth1VotingPeriod               int64  `json:"EPOCHS_PER_ETH1_VOTING_PERIOD,string"`
		SlotsPerHistoricalRoot                  int64  `json:"SLOTS_PER_HISTORICAL_ROOT,string"`
		MinEpochsToInactivityPenalty            int64  `json:"MIN_EPOCHS_TO_INACTIVITY_PENALTY,string"`
		EpochsPerHistoricalVector               int64  `json:"EPOCHS_PER_HISTORICAL_VECTOR,string"`
		EpochsPerSlashingsVector                int64  `json:"EPOCHS_PER_SLASHINGS_VECTOR,string"`
		HistoricalRootsLimit                    int64  `json:"HISTORICAL_ROOTS_LIMIT,string"`
		ValidatorRegistryLimit                  int64  `json:"VALIDATOR_REGISTRY_LIMIT,string"`
		BaseRewardFactor                        int64  `json:"BASE_REWARD_FACTOR,string"`
		WhistleblowerRewardQuotient             int64  `json:"WHISTLEBLOWER_REWARD_QUOTIENT,string"`
		ProposerRewardQuotient                  int64  `json:"PROPOSER_REWARD_QUOTIENT,string"`
		InactivityPenaltyQuotient               int64  `json:"INACTIVITY_PENALTY_QUOTIENT,string"`
		MinSlashingPenaltyQuotient              int64  `json:"MIN_SLASHING_PENALTY_QUOTIENT,string"`
		ProportionalSlashingMultiplier          int64  `json:"PROPORTIONAL_SLASHING_MULTIPLIER,string"`
		MaxProposerSlashings                    int64  `json:"MAX_PROPOSER_SLASHINGS,string"`
		MaxAttesterSlashings                    int64  `json:"MAX_ATTESTER_SLASHINGS,string"`
		MaxAttestations                         int64  `json:"MAX_ATTESTATIONS,string"`
		MaxDeposits                             int64  `json:"MAX_DEPOSITS,string"`
		MaxVoluntaryExits                       int64  `json:"MAX_VOLUNTARY_EXITS,string"`
		InactivityPenaltyQuotientAltair         int64  `json:"INACTIVITY_PENALTY_QUOTIENT_ALTAIR,string"`
		MinSlashingPenaltyQuotientAltair        int64  `json:"MIN_SLASHING_PENALTY_QUOTIENT_ALTAIR,string"`
		ProportionalSlashingMultiplierAltair    int64  `json:"PROPORTIONAL_SLASHING_MULTIPLIER_ALTAIR,string"`
		SyncCommitteeSize                       int64  `json:"SYNC_COMMITTEE_SIZE,string"`
		EpochsPerSyncCommitteePeriod            int64  `json:"EPOCHS_PER_SYNC_COMMITTEE_PERIOD,string"`
		MinSyncCommitteeParticipants            int64  `json:"MIN_SYNC_COMMITTEE_PARTICIPANTS,string"`
		InactivityPenaltyQuotientBellatrix      int64  `json:"INACTIVITY_PENALTY_QUOTIENT_BELLATRIX,string"`
		MinSlashingPenaltyQuotientBellatrix     int64  `json:"MIN_SLASHING_PENALTY_QUOTIENT_BELLATRIX,string"`
		ProportionalSlashingMultiplierBellatrix int64  `json:"PROPORTIONAL_SLASHING_MULTIPLIER_BELLATRIX,string"`
		MaxBytesPerTransaction                  int64  `json:"MAX_BYTES_PER_TRANSACTION,string"`
		MaxTransactionsPerPayload               int64  `json:"MAX_TRANSACTIONS_PER_PAYLOAD,string"`
		BytesPerLogsBloom                       int64  `json:"BYTES_PER_LOGS_BLOOM,string"`
		MaxExtraDataBytes                       int64  `json:"MAX_EXTRA_DATA_BYTES,string"`
		MaxBlsToExecutionChanges                int64  `json:"MAX_BLS_TO_EXECUTION_CHANGES,string"`
		MaxWithdrawalsPerPayload                int64  `json:"MAX_WITHDRAWALS_PER_PAYLOAD,string"`
		MaxValidatorsPerWithdrawalsSweep        int64  `json:"MAX_VALIDATORS_PER_WITHDRAWALS_SWEEP,string"`
		DomainSelectionProof                    string `json:"DOMAIN_SELECTION_PROOF"`
		DomainVoluntaryExit                     string `json:"DOMAIN_VOLUNTARY_EXIT"`
		TargetAggregatorsPerCommittee           int64  `json:"TARGET_AGGREGATORS_PER_COMMITTEE,string"`
		TargetAggregatorsPerSyncSubcommittee    int64  `json:"TARGET_AGGREGATORS_PER_SYNC_SUBCOMMITTEE,string"`
		DomainRandao                            string `json:"DOMAIN_RANDAO"`
		DomainApplicationMask                   string `json:"DOMAIN_APPLICATION_MASK"`
		SyncCommitteeSubnetCount                int64  `json:"SYNC_COMMITTEE_SUBNET_COUNT,string"`
		DomainContributionAndProof              string `json:"DOMAIN_CONTRIBUTION_AND_PROOF"`
		DomainBeaconProposer                    string `json:"DOMAIN_BEACON_PROPOSER"`
		DomainAggregateAndProof                 string `json:"DOMAIN_AGGREGATE_AND_PROOF"`
		DomainDeposit                           string `json:"DOMAIN_DEPOSIT"`
		DomainBeaconAttester                    string `json:"DOMAIN_BEACON_ATTESTER"`
		DomainSyncCommitteeSelectionProof       string `json:"DOMAIN_SYNC_COMMITTEE_SELECTION_PROOF"`
		DomainSyncCommittee                     string `json:"DOMAIN_SYNC_COMMITTEE"`
		BlsWithdrawalPrefix                     string `json:"BLS_WITHDRAWAL_PREFIX"`
	} `json:"data"`
}
