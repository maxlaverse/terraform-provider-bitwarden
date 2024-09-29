package bitwarden

type ListObjectsOptionGenerator func(id string) ListObjectsOption
type ListObjectsFilterOptions struct {
	CollectionFilter   string
	FolderFilter       string
	OrganizationFilter string
	SearchFilter       string
	UrlFilter          string
}

func (f *ListObjectsFilterOptions) IsValid() bool {
	return f.SearchFilter != ""
}

type ListObjectsOption func(filters *ListObjectsFilterOptions)

func WithCollectionID(id string) ListObjectsOption {
	return func(f *ListObjectsFilterOptions) {
		f.CollectionFilter = id
	}
}

func WithFolderID(id string) ListObjectsOption {
	return func(f *ListObjectsFilterOptions) {
		f.FolderFilter = id
	}
}

func WithOrganizationID(id string) ListObjectsOption {
	return func(f *ListObjectsFilterOptions) {
		f.OrganizationFilter = id
	}
}

func WithSearch(search string) ListObjectsOption {
	return func(f *ListObjectsFilterOptions) {
		f.SearchFilter = search
	}
}

func WithUrl(url string) ListObjectsOption {
	return func(f *ListObjectsFilterOptions) {
		f.UrlFilter = url
	}
}

func ListObjectsOptionsToFilterOptions(options ...ListObjectsOption) ListObjectsFilterOptions {
	filter := ListObjectsFilterOptions{}

	for _, option := range options {
		option(&filter)
	}

	return filter
}