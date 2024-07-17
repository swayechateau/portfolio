package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
)

type Project struct {
	Hero       string   `json:"hero"`
	Title      string   `json:"title"`
	Excerpt    string   `json:"excerpt"`
	Tags       []string `json:"tags"`
	OpenSource bool     `json:"open_source"`
	GitRepo    string   `json:"git_repo"`
	LiveUrl    string   `json:"live_url"`
	CaseStudy  string   `json:"case_study"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

type Post struct {
	Locale    string `json:"locale"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Featured  bool   `json:"featured"`
	Excerpt   string `json:"excerpt"`
	HeroImage string `json:"hero_image"`
	Category  string `json:"category"`
	Author    string `json:"author"`
	ReadTime  string `json:"read_time"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	FullUrl   string `json:"full_url"`
}

type ApiResponse struct {
	Recent   []Post `json:"recent"`
	Featured []Post `json:"featured"`
}

type Database struct {
	Projects []Project   `json:"projects"`
	Posts    ApiResponse `json:"posts"`
}

type App struct {
	Home  Home
	About About
	APIs  struct {
		Blog     string
		Projects string
	}
	Database Database
}

type Home struct {
	Title       string
	BlogUrl     string
	ProjectsUrl string
	Projects    []Project
	Posts       []Post
}

type About struct {
	Title string
}

func (db *Database) SaveToCache() error {
	file, err := os.Create("cache.json")
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal data: %w", err)
	}

	if _, err = file.Write(data); err != nil {
		return fmt.Errorf("could not write to file: %w", err)
	}

	return nil
}

func (db *Database) UpdateCacheIfNewData(blogUrl string) error {
	// Fetch the latest data from the API
	newData := Database{}
	if err := newData.FetchFromAPI(blogUrl); err != nil {
		return fmt.Errorf("could not fetch new data from API: %w", err)
	}

	// Compare the new data with the cached data
	if !reflect.DeepEqual(db, &newData) {
		log.Println("New data found, updating cache")
		*db = newData
		if err := db.SaveToCache(); err != nil {
			return fmt.Errorf("could not update cache: %w", err)
		}
	} else {
		log.Println("No new data found")
	}

	return nil
}

func (db *Database) LoadFromCache() error {
	file, err := os.Open("cache.json")
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(db); err != nil {
		return fmt.Errorf("could not decode JSON: %w", err)
	}

	return nil
}

func (db *Database) FetchFromAPI(postApi string) error {
	log.Println("Fetching data from API")
	if err := db.fetchPosts(postApi); err != nil {
		return err
	}
	if err := db.fetchProjects(); err != nil {
		return err
	}
	return db.SaveToCache()
}

func (db *Database) fetchPosts(url string) error {
	apiResponse, err := fetchPostsFromAPI(url)
	if err != nil {
		return fmt.Errorf("error fetching posts: %w", err)
	}
	db.Posts = apiResponse
	return nil
}

func (db *Database) fetchProjects() error {
	projects, err := fetchProjectsFromAPI()
	if err != nil {
		return fmt.Errorf("error fetching projects: %w", err)
	}
	db.Projects = projects
	return nil
}

func fetchPostsFromAPI(url string) (ApiResponse, error) {
	var response ApiResponse

	resp, err := http.Get(url)
	if err != nil {
		return response, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("received non-200 status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, fmt.Errorf("error reading response body: %w", err)
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return response, fmt.Errorf("error unmarshalling response body: %w", err)
	}

	return response, nil
}

func fetchProjectsFromAPI() ([]Project, error) {
	// Replace with actual API call when created
	return []Project{
		{
			Hero:       "https://file.swayechateau.com/view/swayechateauWLGYnBgsrYxGZSputQx822",
			Title:      "The Coldest Sunset",
			Excerpt:    "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Voluptatibus quia, nulla! Maiores et perferendis eaque, exercitationem praesentium nihil.",
			Tags:       []string{"photography", "travel", "winter"},
			OpenSource: true,
			GitRepo:    "https://github.com/swayechateau/fileserver",
			LiveUrl:    "https://file.swayechateau.com",
			CaseStudy:  "https://nobodycare.dev/en/post/building-a-file-server-api",
		},
		{
			Hero:       "https://file.swayechateau.com/view/globaliyndTnSCK14onpASVq7n5?share_code=s5LUL0lAdDLS",
			Title:      "File Server",
			Excerpt:    "Custom built CDN for my media files.",
			Tags:       []string{"markdown", "lumen", "microservice", "mariadb", "api"},
			OpenSource: true,
			GitRepo:    "https://github.com/swayechateau/fileserver",
			LiveUrl:    "https://file.solemnity.icu",
			CaseStudy:  "https://nobodycare.dev/en/post/building-a-file-server-api",
		},
		{
			Hero:       "https://file.swayechateau.com/view/globalMaJKf2UDzFdqba7hG96U6?share_code=s6LHjQlIsFHc",
			Title:      "Web Meta Grabber",
			Excerpt:    "I Wanted an api I had permissions to use to get the meta data from websites for a chat application I was building.",
			Tags:       []string{"markdown", "go", "docker", "microservice", "api"},
			OpenSource: true,
			GitRepo:    "https://github.com/swayechateau/web-meta-grabber",
			LiveUrl:    "https://meta.solemnity.icu/",
			CaseStudy:  "https://nobodycare.dev/en/posts/web-meta-grabber",
		},
	}, nil
}

func (a *App) FetchData() error {
	if err := a.Database.UpdateCacheIfNewData(a.APIs.Blog); err != nil {
		return fmt.Errorf("could not fetch data: %w", err)
	}
	return nil
}

func (a *App) EnsureData() error {
	// Load from cache or fetch from API if cache load fails
	if err := a.Database.LoadFromCache(); err != nil {
		log.Printf("Error loading from cache: %s, fetching from API", err.Error())
		return a.Database.FetchFromAPI(a.APIs.Blog)
	}
	log.Println("Loaded data from cache")

	return a.Database.UpdateCacheIfNewData(a.APIs.Blog)
}

func (a *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.FetchData(); err != nil {
		// Log the error and continue to render the page
		log.Printf("Error fetching data: %s\n", err)
	}
	a.Home.Projects = a.Database.Projects
	a.Home.Posts = a.Database.Posts.Recent
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, a.Home)
}

func (a *App) AboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/about.html")
	if err != nil {
		log.Printf("Templating Error: %s\n", err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, a.About)
}

func main() {
	var app App
	app.APIs.Blog = "http://localhost:8000/api/posts"
	app.APIs.Projects = "http://localhost:8000/api/projects"

	// Load from cache or fetch from API if cache load fails
	if err := app.EnsureData(); err != nil {
		log.Printf("Error loading from API: %s", err.Error())
	}

	app.Home = Home{
		Title:       "Welcome To My Portfolio | Swaye Chateau",
		BlogUrl:     "http://localhost:8000",
		ProjectsUrl: "http://localhost:8000/projects",
		Projects:    []Project{},
		Posts:       []Post{},
	}

	app.About = About{
		Title: "About Me | Swaye Chateau",
	}

	http.HandleFunc("/", app.HomeHandler)
	http.HandleFunc("/about", app.AboutHandler)
	log.Println("Starting server on :5000")
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
