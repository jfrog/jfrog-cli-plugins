package utils

func CreateSearchBySha1AndRepoAqlQuery(repo string, sha1s []string) string {
	return "{" + getRepoQuery(repo) + getSha1sQueryPart(sha1s) + "}"
}

func getSha1sQueryPart(sha1s []string) string {
	if len(sha1s) > 1 {
		result := `"$or":[{`
		for _, sha1 := range sha1s {
			result += `"actual_sha1": "` + sha1 + `",`
		}
		return result[:len(result)-1] + `}]`
	}
	return `"actual_sha1": "` + sha1s[0] + `"`
}

func getRepoQuery(repositories string) string {
	if len(repositories) == 0 {
		return ""
	}
	return `"repo": "` + repositories + `",`
}

// AQL requests have a size limit, therefore, we split the requests into small groups.
// Group large slice into small groups, for example:
// sliceToGroup = []string{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}; groupSize = 3
// returns : [['0' '1' '2'] ['3' '4' '5'] ['6' '7' '8'] ['9]]
func GroupItems(sliceToGroup []string, groupSize int) [][]string {
	var groups [][]string
	if groupSize > len(sliceToGroup) {
		return append(groups, sliceToGroup)
	}
	for groupSize < len(sliceToGroup) {
		sliceToGroup, groups = sliceToGroup[groupSize:], append(groups, sliceToGroup[0:groupSize:groupSize])
	}
	return append(groups, sliceToGroup)
}
