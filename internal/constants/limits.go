package constants

const (
	// DefaultPlaceLimit is the default limit for fetching all places when no specific limit is set
	DefaultPlaceLimit = 10000

	// DefaultExportLimit is the reasonable upper bound for export operations
	DefaultExportLimit = 50000

	// MaxPhotosPerPlace is the maximum number of photos to download per place
	MaxPhotosPerPlace = 5

	// PhotoDownloadDelay is the delay between photo downloads to avoid rate limiting
	PhotoDownloadDelayMs = 50
)
