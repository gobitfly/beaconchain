package types

type MobileBundleData struct {
	BundleUrl                string `json:"bundle_url,omitempty"`
	HasNativeUpdateAvailable bool   `json:"has_native_update_available"`
}

type GetMobileLatestBundleResponse ApiDataResponse[MobileBundleData]
