package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Commit struct {
	SHA         string    `json:"sha" db:"sha"`
	Message     string    `json:"message" db:"message"`
	URL         string    `json:"url" db:"url"`
	AuthorName  string    `json:"author_name" db:"author_name"`
	AuthorEmail string    `json:"author_email" db:"author_email"`
	Date        time.Time `json:"date" db:"date"`

	RepositoryID int `json:"repository_id" db:"repository_id"`
}

// Save saves the given repository metadata to the repositories table
func (c *Commit) Save() error {
	// get db connection instance
	db, err := SQLConnect()
	if err != nil {
		return err
	}

	// insert statement
	insert := `insert into commits (
		sha,
		message,
		url,
		author_name,
		author_email,
		date,
		repository_id
	) values ($1,$2,$3,$4,$5,$6,$7)
	 ON CONFLICT (sha) DO NOTHING`

	// execute insert statement
	_, err = db.Exec(insert,
		c.SHA,
		c.Message,
		c.URL,
		c.AuthorName,
		c.AuthorEmail,
		c.Date,
		c.RepositoryID,
	)

	return err
}

// FetchCommits fetchs the repository metadata from github, stores it, and return it
func FetchCommits(repo_url string, start *time.Time) ([]Commit, error) {
	// get repo from repos table
	repo, err := GetRepoByURL(repo_url)
	if err != nil {
		fmt.Println("Error getting repository from database : ", err)
	}

	URL := repo_url + "/commits"
	if start != nil {
		URL += "?since=" + start.Format("2006-01-02T15:04:05Z")
	}

	// create the request
	req, err := NewRequest("GET", URL, nil)
	if err != nil {
		fmt.Println(err)
	}

	// create https client and execute the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response : ", err)
		fmt.Println(err)
	}

	commits := []Commit{}
	response := []struct {
		SHA    string `json:"sha"`
		URL    string `json:"url"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				AuthorName  string    `json:"name"`
				AuthorEmail string    `json:"email"`
				Date        time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error parsing commits : ", err)
		return nil, err
	}

	for _, c := range response {
		commits = append(commits, Commit{
			SHA:          c.SHA,
			Message:      c.Commit.Message,
			URL:          c.URL,
			AuthorName:   c.Commit.Author.AuthorName,
			AuthorEmail:  c.Commit.Author.AuthorEmail,
			Date:         c.Commit.Author.Date,
			RepositoryID: repo.ID,
		})
	}

	// clear existing commits
	err = DeleteCommitByRepoID(repo.ID)
	if err != nil {
		return nil, err
	}

	// save commits
	for _, commit := range commits {
		commit.RepositoryID = repo.ID
		err = commit.Save()
		if err != nil {
			fmt.Println("Error saving commit : ", err)
		}
	}

	return commits, nil
}
func FetchCommitsNoOverride(repo_url string, start *time.Time) ([]Commit, error) {
	// get repo from repos table
	repo, err := GetRepoByURL(repo_url)
	if err != nil {
		fmt.Println("Error getting repository from database : ", err)
	}

	URL := repo_url + "/commits"
	if start != nil {
		URL += "?since=" + start.Format("2006-01-02T15:04:05Z")
	}

	// create the request
	req, err := NewRequest("GET", URL, nil)
	if err != nil {
		fmt.Println(err)
	}

	// create https client and execute the request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError(fmt.Errorf("error reading commits response : %v", err))
	}

	commits := []Commit{}
	response := []struct {
		SHA    string `json:"sha"`
		URL    string `json:"url"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				AuthorName  string    `json:"name"`
				AuthorEmail string    `json:"email"`
				Date        time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}{}

	LogApp("Response : " + string(body))

	err = json.Unmarshal(body, &response)
	if err != nil {
		LogError(fmt.Errorf("error parsing commits : %v", err))
	}

	for _, c := range response {
		commits = append(commits, Commit{
			SHA:          c.SHA,
			Message:      c.Commit.Message,
			URL:          c.URL,
			AuthorName:   c.Commit.Author.AuthorName,
			AuthorEmail:  c.Commit.Author.AuthorEmail,
			Date:         c.Commit.Author.Date,
			RepositoryID: repo.ID,
		})
	}

	// save commits
	for _, commit := range commits {
		commit.RepositoryID = repo.ID
		err = commit.Save()
		if err != nil {
			LogError(fmt.Errorf("error saving commits : %v", err))
		}
	}

	return commits, nil
}

func GetCommits(repo_id int) ([]Commit, error) {
	db, err := SQLConnect()
	if err != nil {
		fmt.Println("Error connecting to the database : ", err)
		return nil, err
	}

	commits := []Commit{}
	rows, err := db.Query("SELECT sha, message, url, author_name, author_email, date, repository_id FROM commits WHERE repository_id=$1", repo_id)
	for rows.Next() {
		c := Commit{}
		err = rows.Scan(&c.SHA, &c.Message, &c.URL, &c.AuthorName,
			&c.AuthorEmail, &c.Date, &c.RepositoryID)
		if err != nil {
			return nil, err
		}

		commits = append(commits, c)
	}

	return commits, nil
}

func GetLastCommit(repo_id int) (*Commit, error) {
	db, err := SQLConnect()
	if err != nil {
		fmt.Println("Error connecting to the database : ", err)
		return nil, err
	}

	c := Commit{}
	row := db.QueryRow("SELECT sha, message, url, author_name, author_email, date, repository_id FROM commits WHERE repository_id=$1 ORDER BY date DESC LIMIT 1", repo_id)
	err = row.Scan(&c.SHA, &c.Message, &c.URL, &c.AuthorName,
		&c.AuthorEmail, &c.Date, &c.RepositoryID)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

type Author struct {
	AuthorName  string `db:"name"`
	AuthorEmail string `db:"email"`
	Commits     int    `db:"commits"`
}

func GetTopAuthors(repo_id, n int) ([]Author, error) {
	db, err := SQLConnect()
	if err != nil {
		fmt.Println("Error connecting to the database : ", err)
		return nil, err
	}

	authors := []Author{}
	qry := `SELECT author_name, author_email, count(sha) AS commits 
		FROM commits 
		WHERE repository_id=$1 
		GROUP BY author_name, author_email
		ORDER BY commits DESC
	`
	if n > 0 {
		// top n authors
		qry = fmt.Sprintf("%s LIMIT %d", qry, n)
	}

	rows, err := db.Query(qry, repo_id)
	if err != nil {
		fmt.Println("Error getting top authors from db : ", err)
		return nil, err
	}
	for rows.Next() {
		a := Author{}
		err = rows.Scan(&a.AuthorName, &a.AuthorEmail, &a.Commits)
		if err != nil {

			return nil, err
		}

		authors = append(authors, a)
	}

	return authors, nil
}

func DeleteCommitByRepoID(repo_id int) error {
	db, err := SQLConnect()
	if err != nil {
		fmt.Println("Error connecting to the database : ", err)
		return err
	}

	_, err = db.Exec("DELETE FROM commits WHERE repository_id=$1", repo_id)
	if err != nil {
		return err
	}

	return nil
}

// init function to create the commits table if it doesn't exist already
func init() {
	// make sure commits table exists
	db, err := SQLConnect()
	if err != nil {
		return
	}

	create := `CREATE TABLE IF NOT EXISTS commits (
		sha varchar(255) PRIMARY KEY,
		message varchar(255),
		url varchar(255) UNIQUE,
		author_name varchar(255),
		author_email varchar(255),
		repository_id INTEGER,
		date timestamp,
		FOREIGN KEY (repository_id) REFERENCES repositories(id)
	)`

	_, err = db.Exec(create)
	if err != nil {
		fmt.Println("Error creating commits table : ", err)
	}
}
