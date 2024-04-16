package bw

type ListObjectsOption func(args *[]string)
type ListObjectsOptionGenerator func(id string) ListObjectsOption

func WithCollectionID(id string) ListObjectsOption {
	return func(args *[]string) {
		*args = append(*args, "--collectionid", id)
	}
}

func WithFolderID(id string) ListObjectsOption {
	return func(args *[]string) {
		*args = append(*args, "--folderid", id)
	}
}

func WithOrganizationID(id string) ListObjectsOption {
	return func(args *[]string) {
		*args = append(*args, "--organizationid", id)
	}
}

func WithSearch(search string) ListObjectsOption {
	return func(args *[]string) {
		*args = append(*args, "--search", search)
	}
}

func WithUrl(url string) ListObjectsOption {
	return func(args *[]string) {
		*args = append(*args, "--url", url)
	}
}
