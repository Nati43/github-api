package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Repository struct {
	ID              int       `json:"id" db:"id"`
	Name            string    `json:"name" db:"name"`
	Description     string    `json:"description" db:"description"`
	URL             string    `json:"url" db:"url"`
	Language        string    `json:"language" db:"language"`
	ForksCount      int       `json:"forks_count" db:"forks_count"`
	StarsCount      int       `json:"stars_count" db:"stars_count"`
	OpenIssuesCount int       `json:"open_issues_count" db:"open_issues_count"`
	WatchersCount   int       `json:"watchers_count" db:"watchers_count"`
	Created         time.Time `json:"created_at" db:"created_at"`
	Pushed          time.Time `json:"pushed_at" db:"pushed_at"`
	Updated         time.Time `json:"updated_at" db:"updated_at"`
}

// Save saves the given repository metadata to the repositories table
func (r *Repository) Save() error {
	// get db connection instance
	db, err := SQLConnect()
	if err != nil {
		fmt.Println("Error connecting to the database : ", err)
		return err
	}

	// insert statement
	insert := `insert into repositories (
		name,
		description,
		url,
		language,
		forks_count,
		stars_count,
		open_issues_count,
		watchers_count,
		created_at,
		pushed_at,
		updated_at
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	ON CONFLICT (url) DO UPDATE SET 
		language=$4,
		forks_count=$5,
		stars_count=$6,
		open_issues_count=$7,
		watchers_count=$8,
		updated_at=$11
	`

	// execute insert statement
	_, err = db.Exec(insert,
		r.Name,
		r.Description,
		r.URL,
		r.Language,
		r.ForksCount,
		r.StarsCount,
		r.OpenIssuesCount,
		r.WatchersCount,
		r.Created,
		r.Pushed,
		r.Updated)

	return err
}

// FetchRepo fetchs the repository metadata from github, stores it, and return it
func FetchRepo(repo_url string) (*Repository, error) {
	req, err := NewRequest("GET", repo_url, nil)
	if err != nil {
		LogError(fmt.Errorf("error creating repository request : %v", err))
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LogError(fmt.Errorf("error with repository request : %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			LogError(fmt.Errorf("error reading repository response  : %v", err))
			return nil, err
		}

		repo := new(Repository)
		err = json.Unmarshal(body, &repo)
		if err != nil {
			LogError(fmt.Errorf("error parsing repository metadata : %v", err))
			return nil, err
		}

		err = repo.Save()
		if err != nil {
			LogError(fmt.Errorf("error saving repository metadata : %v", err))
			return repo, err
		}

		return repo, nil
	} else {
		err = fmt.Errorf("error fetching repo : %v, %v", resp.StatusCode, resp.Status)
		LogError(err)
		return nil, err
	}
}

func GetRepos() ([]Repository, error) {
	db, err := SQLConnect()
	if err != nil {
		LogError(fmt.Errorf("error connecting to the database : %v", err))
		return nil, err
	}

	repos := []Repository{}
	rows, err := db.Query("SELECT * FROM repositories")
	for rows.Next() {
		r := Repository{}
		err = rows.Scan(&r.ID,
			&r.Name, &r.Description, &r.URL, &r.Language,
			&r.ForksCount, &r.StarsCount, &r.OpenIssuesCount,
			&r.WatchersCount, &r.Created, &r.Pushed, &r.Updated)
		if err != nil {
			LogError(fmt.Errorf("error scanning repository result  : %v", err))
			return nil, err
		}

		repos = append(repos, r)
	}

	return repos, nil
}

func GetRepoByID(id int) (*Repository, error) {
	db, err := SQLConnect()
	if err != nil {
		LogError(fmt.Errorf("error connecting to the database : %v", err))
		return nil, err
	}

	r := new(Repository)
	row := db.QueryRow("SELECT * FROM repositories WHERE id=$1", id)
	err = row.Scan(&r.ID,
		&r.Name, &r.Description, &r.URL, &r.Language,
		&r.ForksCount, &r.StarsCount, &r.OpenIssuesCount,
		&r.WatchersCount, &r.Created, &r.Pushed, &r.Updated)
	if err != nil {
		LogError(fmt.Errorf("error scanning repository response : %v", err))
		return nil, err
	}

	return r, nil
}

func GetRepoByURL(repo_url string) (*Repository, error) {
	db, err := SQLConnect()
	if err != nil {
		fmt.Println("Error connecting to the database : ", err)
		return nil, err
	}

	r := new(Repository)
	row := db.QueryRow("SELECT * FROM repositories WHERE url=$1", repo_url)
	err = row.Scan(&r.ID,
		&r.Name, &r.Description, &r.URL, &r.Language,
		&r.ForksCount, &r.StarsCount, &r.OpenIssuesCount,
		&r.WatchersCount, &r.Created, &r.Pushed, &r.Updated)
	if err != nil {
		LogError(fmt.Errorf("error scanning repository response : %v", err))
	}

	return r, nil
}

// init function to create the repositories table if it doesn't exist already
func init() {
	// make sure repositories table exists
	db, err := SQLConnect()
	if err != nil {
		LogError(fmt.Errorf("error connecting to the database : %v", err))
		return
	}

	create := `CREATE TABLE IF NOT EXISTS repositories (
		id SERIAL PRIMARY KEY,
		name varchar(255) NOT NULL,
		description varchar(255),
		url varchar(255) UNIQUE,
		language varchar(255),
		forks_count int,
		stars_count int,
		open_issues_count int,
		watchers_count int,
		created_at timestamp,
		pushed_at timestamp,
		updated_at timestamp
	)`

	_, err = db.Exec(create)
	if err != nil {
		LogError(fmt.Errorf("error creating repositories table : %v", err))
	}
}
