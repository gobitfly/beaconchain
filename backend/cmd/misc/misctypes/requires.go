package misctypes

type Requires struct {
	Bigtable      bool
	RawBigtable   bool
	Redis         bool
	ClNode        bool
	ElNode        bool
	NetworkDBs    bool
	UserDBs       bool
	ClickhouseDBs bool
}
