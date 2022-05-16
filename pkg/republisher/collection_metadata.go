package republisher

const (
	ScopeBoth           = "both"
	ScopeContent        = "content"
	ScopeMetadata       = "metadata"
	CmsNotifier         = "cms-notifier"
	CmsMetadataNotifier = "cms-metadata-notifier"
)

type Collections map[string]CollectionMetadata

type CollectionMetadata struct {
	name                  string
	defaultOriginSystemID string
	notifierApp           string
	scope                 string
}

var DefaultCollections = Collections{
	"universal-content": {
		name:                  "universal-content",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/cct",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
	"video": {
		name:                  "video",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/next-video-editor",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
	"pac-metadata": {
		name:                  "pac-metadata",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/pac",
		notifierApp:           CmsMetadataNotifier,
		scope:                 ScopeMetadata,
	},
	"next-video-editor": {
		name:                  "video-metadata",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/next-video-editor",
		notifierApp:           CmsMetadataNotifier,
		scope:                 ScopeMetadata,
	},
	"pages": {
		name:                  "pages",
		defaultOriginSystemID: "http://cmdb.ft.com/systems/spark-lists",
		notifierApp:           CmsNotifier,
		scope:                 ScopeContent,
	},
}
