package types

type ApiWeightItem struct {
	Bucket   string `db:"bucket"`
	Endpoint string `db:"endpoint"`
	Method   string `db:"method"`
	Weight   int    `db:"weight"`
}

type InternalGetRatelimitWeightsResponse ApiDataResponse[[]ApiWeightItem]
