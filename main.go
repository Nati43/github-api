package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/nsf/termbox-go"
)

var repos = []string{}
var repositories = []Repository{}
var repository Repository

var commits = []Commit{}
var commitsShort = []string{}
var err error

type Menu struct {
	title    string
	items    []string
	parent   *Menu
	selected int
}

var mainMenu = &Menu{
	title: "Main Menu",
	items: []string{
		"- List Repositories",
		"- Add Repository",
		"Exit",
	},
}

var reposList = &Menu{
	title:  "Repositories",
	items:  []string{},
	parent: mainMenu,
}

var repoMenu = &Menu{
	title: "Repository Menu",
	items: []string{
		"- Commits",
		"- Pull",
		"- Top Authors",
		"Back",
	},
	parent: reposList,
}

var commitsList = &Menu{
	title:  "Commits",
	items:  []string{},
	parent: repoMenu,
}

var authorsList = &Menu{
	title:  "Commits",
	items:  []string{},
	parent: repoMenu,
}

var currentMenu *Menu

func main() {
	// start refresh cron job
	startCRON()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.Clear(termbox.ColorMagenta, termbox.ColorMagenta)

	currentMenu = mainMenu
	drawMenu(currentMenu)
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				if currentMenu.selected > 0 {
					currentMenu.selected--
				}
			case termbox.KeyArrowDown:
				if currentMenu.selected < len(currentMenu.items)-1 {
					currentMenu.selected++
				}
			case termbox.KeyEnter:
				handleSelect()
			case termbox.KeyEsc:
				return
			}
		}
		drawMenu(currentMenu)
	}
}
func handleSelect() {
	switch currentMenu.title {
	case "Main Menu":
		switch currentMenu.selected {
		case 0:
			// List repos selected
			repositories, err = GetRepos()
			if err != nil {
				fmt.Println("Error getting repositories from db : ", err)
			}
			repos = []string{}
			repos = append(repos, fmt.Sprintf("ID\t\t\tName\t\t\tLanguage\t\t\tForks\t\t\tStars\t\t\tIssues\t\t\tWatchers"))
			for _, r := range repositories {
				repos = append(repos, fmt.Sprintf("%d\t\t\t%s\t\t\t%s\t\t\t%d\t\t\t%d\t\t\t%d\t\t\t%d",
					r.ID, r.Name, r.Language,
					r.ForksCount, r.StarsCount,
					r.OpenIssuesCount, r.WatchersCount))
			}
			repos = append(repos, "Back")
			reposList.items = repos

			currentMenu = reposList
			currentMenu.selected = 1
		case 1:
			// add repo selected
			url := promptForRepoURL()
			// fetch commits
			y := 2
			drawText(0, y, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Fetching rep : %v", url))
			y++
			termbox.Flush()
			time.Sleep(2 * time.Second)

			_, err := FetchRepo(url)
			if err != nil {
				drawText(0, y, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Error fetching repository : %v", err))
				y++
			}

			repositories, err = GetRepos()
			if err != nil {
				fmt.Println("Error getting repositories from db : ", err)
			}
			repos = []string{}
			repos = append(repos, fmt.Sprintf("ID\t\t\tName\t\t\tLanguage\t\t\tForks\t\t\tStars\t\t\tIssues\t\t\tWatchers"))
			for _, r := range repositories {
				repos = append(repos, fmt.Sprintf("%d\t\t\t%s\t\t\t%s\t\t\t%d\t\t\t%d\t\t\t%d\t\t\t%d",
					r.ID, r.Name, r.Language,
					r.ForksCount, r.StarsCount,
					r.OpenIssuesCount, r.WatchersCount))
			}
			repos = append(repos, "Back")
			reposList.items = repos

			currentMenu = reposList
			currentMenu.selected = len(reposList.items) - 2
		case 2:
			// exit
			termbox.Close()
			os.Exit(0)
			return
		}
	case "Repositories":
		// under repos list
		switch currentMenu.selected {
		case 0:
			// don't do anything, first item is header row
		case len(currentMenu.items) - 1:
			// back selected
			currentMenu = currentMenu.parent
			currentMenu.selected = 0
		default:
			// repo selected
			repository = repositories[currentMenu.selected-1]
			currentMenu = repoMenu
			currentMenu.selected = 0
		}
	case "Repository Menu":
		switch currentMenu.selected {
		case 0:
			// commits
			commits, err = GetCommits(repository.ID)
			if err != nil {
				fmt.Println("Error getting repositories from db : ", err)
			}

			commitsShort = []string{}
			commitsShort = append(commitsShort, fmt.Sprintf("Date\t\t\t\tAuthor\t\t\t\tMessage"))
			for _, c := range commits {
				msg := c.Message
				if len(msg) > 50 {
					msg = msg[:50] + "..."
				}

				commitsShort = append(commitsShort, fmt.Sprintf("%s\t\t\t\t%s\t\t\t\t%s",
					c.Date.Format("2006-01-02"), c.AuthorName, msg))
			}
			commitsShort = append(commitsShort, "Back")
			commitsList.items = commitsShort

			currentMenu = commitsList
			currentMenu.selected = 1
		case 1:
			// pull selected
			y := 2
			// take date input from user
			t, err := promptForDate()
			if err != nil {
				drawText(0, y, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Error parsing your date : %v", err))
				y++
			}

			if t != nil {
				drawText(0, y, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Pulling commits since : %v", t.Format("2006-01-02")))
			} else {
				drawText(0, y, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Pulling commits from start"))
			}
			y++
			termbox.Flush()
			time.Sleep(2 * time.Second)

			// fetch commits
			commits, err := FetchCommits(repository.URL, t)
			if err != nil {
				drawText(0, y, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Error fetching commits : %v", err))
				y++
			}

			// update display
			commitsShort = []string{}
			commitsShort = append(commitsShort, fmt.Sprintf("Date\t\t\t\tAuthor\t\t\t\tMessage"))
			for _, c := range commits {
				msg := c.Message
				if len(msg) > 50 {
					msg = msg[:50] + "..."
				}

				commitsShort = append(commitsShort, fmt.Sprintf("%s\t\t\t\t%s\t\t\t\t%s",
					c.Date.Format("2006-01-02"), c.AuthorName, msg))
			}
			commitsShort = append(commitsShort, "Back")
			commitsList.items = commitsShort

			currentMenu = commitsList
			currentMenu.selected = 1
		case 2:
			// top authors
			authors, err := GetTopAuthors(repository.ID, 10)
			if err != nil {
				drawText(0, 0, termbox.ColorWhite, termbox.ColorDefault, fmt.Sprintf("Error getting authors from database : %v", err))
			}

			authorsShort := []string{}
			authorsShort = append(commitsShort, fmt.Sprintf("Commits\t\t\t\tEmail\t\t\t\tName"))
			for _, a := range authors {

				authorsShort = append(authorsShort, fmt.Sprintf("%d\t\t\t\t%s\t\t\t\t%s",
					a.Commits, a.AuthorEmail, a.AuthorName))
			}
			authorsShort = append(authorsShort, "Back")
			authorsList.items = authorsShort

			currentMenu = authorsList
			currentMenu.selected = 1
		case len(currentMenu.items) - 1:
			// back selected
			currentMenu = currentMenu.parent
			currentMenu.selected = 0
		}
	case "Commits":
		switch currentMenu.selected {
		case len(currentMenu.items) - 1:
			// back selected
			currentMenu = currentMenu.parent
			currentMenu.selected = 0
		}
	}
}

func drawMenu(menu *Menu) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawText(0, 0, termbox.ColorWhite, termbox.ColorDefault, menu.title)
	for i, item := range menu.items {
		if i == menu.selected {
			drawText(0, i+1, termbox.ColorBlack, termbox.ColorWhite, item)
		} else {
			drawText(0, i+1, termbox.ColorWhite, termbox.ColorDefault, item)
		}
	}
	termbox.Flush()
}

func drawText(x, y int, fg, bg termbox.Attribute, text string) {
	for i, c := range text {
		termbox.SetCell(x+i, y, c, fg, bg)
	}
}

func promptForDate() (*time.Time, error) {
	x, y := 0, 0
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawText(x, y, termbox.ColorWhite, termbox.ColorDefault, "Please enter a date to pull commits since (YYYY-MM-DD)")
	y = 1
	l := "Leave emtpt to pull from first commit: "
	drawText(x, y, termbox.ColorWhite, termbox.ColorDefault, l)
	x = len(l)
	termbox.Flush()

	var t time.Time

	var input []rune
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			if ev.Key == termbox.KeyEnter {
				break
			} else if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			} else if ev.Ch != 0 {
				input = append(input, ev.Ch)
			}
			drawText(x, 1, termbox.ColorWhite, termbox.ColorDefault, strings.Repeat(" ", len(input)+10))
			drawText(x, 1, termbox.ColorWhite, termbox.ColorDefault, string(input))
			termbox.Flush()
		}
	}

	date := string(input)
	if date != "" {
		t, err = time.Parse("2006-01-02", date)
		if err != nil {
			return nil, err
		}
	}

	termbox.Flush()

	return &t, nil
}

func promptForRepoURL() string {
	x, y := 0, 0
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawText(x, y, termbox.ColorWhite, termbox.ColorDefault, "Please enter the repository URL : ")
	termbox.Flush()

	var input []rune
	for {
		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			if ev.Key == termbox.KeyEnter {
				break
			} else if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if len(input) > 0 {
					input = input[:len(input)-1]
				}
			} else if ev.Ch != 0 {
				input = append(input, ev.Ch)
			}
			drawText(x, 1, termbox.ColorWhite, termbox.ColorDefault, strings.Repeat(" ", len(input)+10))
			drawText(x, 1, termbox.ColorWhite, termbox.ColorDefault, string(input))
			termbox.Flush()
		}
	}

	return string(input)
}

func startCRON() {
	interval := os.Getenv("INTERVAL")
	i := uint64(1)
	if interval != "" {
		val, err := strconv.Atoi(interval)
		if err != nil {
			LogError(fmt.Errorf("error parsing INTERVAL env variable : %v", err))
		}
		i = uint64(val)
	}
	gocron.Every(i).Hours().Do(RefreshRepos)
	gocron.Start()
}
