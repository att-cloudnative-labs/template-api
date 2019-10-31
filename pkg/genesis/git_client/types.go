package git_client

type BitBucketRepoResponse struct {
	Size       int                         `json:"size"`
	Limit      int                         `json:"limit"`
	IsLastPage bool                        `json:"isLastPage"`
	Values     []BitBucketRepoResponseItem `json:"values"`
}

type BitBucketRepoResponseItem struct {
	Slug          string                       `json:"slug"`
	Id            int                          `json:"id"`
	Name          string                       `json:"name"`
	ScmId         string                       `json:"scm_id"`
	State         string                       `json:"state"`
	StatusMessage string                       `json:"status_message"`
	Forkable      bool                         `json:"forkable"`
	Project       BitBucketProjectResponseItem `json:"project"`
}

type BitBucketProjectResponseItem struct {
	Key         string `json:"key"`
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Type        string `json:"type"`
}

func (b *BitBucketRepoResponse) GetRepositoryNames() []string {
	repoNames := make([]string, b.Size)

	for idx, item := range b.Values {
		repoNames[idx] = item.Name
	}

	return repoNames
}
