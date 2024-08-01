package main

import (
	"fmt"
)

func RefreshRepos() {
	repos, err := GetRepos()
	if err != nil {
		// error loading repos to pull changes
		LogError(fmt.Errorf("error fetching repos from the db : %s", err))
	}

	for _, r := range repos {
		// refresh repo meta data first
		_, err := FetchRepo(r.URL)
		if err != nil {
			LogError(fmt.Errorf("error fetching repo metadata : %v", err))
		}

		lastCommit, err := GetLastCommit(r.ID)
		if err != nil {
			LogError(fmt.Errorf("error fetching last commit from db : %v", err))
		}

		// pull commits
		_, err = FetchCommitsNoOverride(r.URL, &lastCommit.Date)
		if err != nil {
			LogError(fmt.Errorf("error fetching new commits : %v", err))
		}
	}

}
