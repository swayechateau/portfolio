package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"
	"time"
)

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

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
	CSRFToken     CSRFToken
	Database      Database
	TemplateCache map[string]*template.Template
	Home          Home
	About         About
}

type Home struct {
	Title            string
	BlogUrl          string
	ProjectsUrl      string
	Projects         []Project
	Posts            []Post
	Submitted        bool
	SubmittedMessage string
	SubmittedClass   string
	CSRF             string
}

type About struct {
	Title       string
	BlogUrl     string
	ProjectsUrl string
}

type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
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

func (db *Database) UpdateCacheIfNewData(blogUrl, token string) error {
	newData := Database{}
	if err := newData.FetchFromAPI(blogUrl, token); err != nil {
		return fmt.Errorf("could not fetch new data from API: %w", err)
	}

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

func (db *Database) FetchFromAPI(blogUrl, token string) error {
	var wg sync.WaitGroup
	var errPosts, errProjects error

	wg.Add(2)

	go func() {
		defer wg.Done()
		errPosts = db.fetchPosts(blogUrl, token)
	}()

	go func() {
		defer wg.Done()
		errProjects = db.fetchProjects()
	}()

	wg.Wait()

	if errPosts != nil {
		return fmt.Errorf("error fetching posts: %w", errPosts)
	}
	if errProjects != nil {
		return fmt.Errorf("error fetching projects: %w", errProjects)
	}

	return db.SaveToCache()
}

func (db *Database) fetchPosts(url, token string) error {
	apiResponse, err := fetchPostsFromAPI(url, token)
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

func (a *App) CacheTemplates(filenames ...string) {
	if a.TemplateCache == nil {
		a.TemplateCache = make(map[string]*template.Template)
	}

	for _, filename := range filenames {
		tmpl, err := template.ParseFiles(filename)
		if err != nil {
			log.Fatalf("Error parsing template %s: %s\n", filename, err)
		}
		a.TemplateCache[filename] = tmpl
	}
}

func (a *App) FetchData() error {
	if err := a.Database.UpdateCacheIfNewData(a.GetBlogAPI(), a.GetBlogApiToken()); err != nil {
		return fmt.Errorf("could not fetch data: %w", err)
	}
	return nil
}

func (a *App) EnsureData() error {
	if err := a.Database.LoadFromCache(); err != nil {
		log.Printf("Error loading from cache: %s, fetching from API", err)
		return a.Database.FetchFromAPI(a.GetBlogAPI(), a.GetBlogApiToken())
	}
	log.Println("Loaded data from cache")

	return a.Database.UpdateCacheIfNewData(a.GetBlogAPI(), a.GetBlogApiToken())
}

func (a *App) GetBlogUrl() string {
	return urlFallback(
		os.Getenv("BLOG_URL"),
		"http://localhost:8000",
	)
}
func (a *App) GetBlogAPI() string {
	return urlFallback(
		os.Getenv("BLOG_API"),
		"http://localhost:8000/api/posts",
	)
}
func (a *App) GetBlogApiToken() string {
	return os.Getenv("BLOG_API_TOKEN")
}

func (a *App) GetBlogClientId() string {
	return os.Getenv("BLOG_CLIENT_ID")
}
func (a *App) GetBlogClientSecret() string {
	return os.Getenv("BLOG_CLIENT_SECRET")
}

func (a *App) GetBlogApiAuthToken() string {
	client := &http.Client{}

	formData := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     a.GetBlogClientId(),
		"client_secret": a.GetBlogClientSecret(),
		"scope":         "",
	}
	formDataBytes, err := json.Marshal(formData)
	if err != nil {
		log.Fatalf("Error marshalling form data: %v", err)
	}

	req, err := http.NewRequest("POST", a.GetBlogUrl()+"/oauth/token", bytes.NewBuffer(formDataBytes))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Received non-200 status code: %d, body: %s", resp.StatusCode, body)
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	accessToken := tokenResponse.AccessToken
	fmt.Printf("Access Token: %s\n", accessToken)
	return accessToken
}

func (a *App) GetProjectsUrl() string {
	return urlFallback(
		os.Getenv("PROJECTS_URL"),
		"http://localhost:8000/projects",
	)
}

func (a *App) GetProjectsAPI() string {
	return urlFallback(
		os.Getenv("PROJECTS_API"),
		"http://localhost:8000/api/projects",
	)
}

func (a *App) GetCSRFToken() string {
	if a.CSRFToken.Token == "" {
		a.CSRFToken, _ = generateCSRFToken()
	}
	return a.CSRFToken.Token
}

func (a *App) NewCSRFToken() string {
	a.CSRFToken, _ = generateCSRFToken()
	return a.CSRFToken.Token
}

func (a *App) ValidateCSRFToken(token string) bool {
	if a.CSRFToken.Token == "" {
		return false
	}
	return a.CSRFToken.Token == token && time.Now().Before(a.CSRFToken.ExpiresAt)
}

func (a *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	a.Home.CSRF = a.NewCSRFToken()
	if err := a.FetchData(); err != nil {
		log.Printf("Error fetching data: %s\n", err)
	}
	a.Home.Projects = a.Database.Projects
	a.Home.Posts = a.Database.Posts.Recent
	a.Home.SubmittedClass = "hidden"
	// Check for query parameters
	query := r.URL.Query()
	if query.Get("status") == "success" {
		a.Home.Submitted = true
		a.Home.SubmittedClass = "border-green-500"
		a.Home.SubmittedMessage = "Contact form submitted successfully"
	}

	if query.Get("status") == "error" {
		a.Home.Submitted = true
		a.Home.SubmittedClass = "border-red-500"
		a.Home.SubmittedMessage = "An error occurred while submitting the contact form"
	}

	if a.Home.Submitted {
		log.Printf("Contact form submitted: %s\n", a.Home.SubmittedMessage)
	}

	tmpl, ok := a.TemplateCache["templates/index.html"]
	if !ok {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, a.Home); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

func (a *App) AboutHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, ok := a.TemplateCache["templates/about.html"]
	if !ok {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, a.About); err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

func (a *App) ContactFormHandler(w http.ResponseWriter, r *http.Request) {
	acceptHeader := r.Header.Get("Accept")
	log.Printf("Accept header: %s\n", acceptHeader)
	r.ParseForm()
	if acceptHeader == "application/json" {
		log.Println("Received JSON request")
		a.ContactFormJSONHandler(w, r)
		return
	}
	log.Println("Received form request")
	a.ContactFormRedirectHandler(w, r)
}

func (a *App) ContactFormJSONHandler(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	response := Response{
		Status:  "success",
		Message: "Contact form submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		response.Status = "error"
		response.Message = "Method not allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		jsonResponse, _ := json.Marshal(response)
		w.Write(jsonResponse)
		log.Printf("Method not allowed: %s\n", r.Method)
		return
	}

	form := ContactForm{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Message: r.FormValue("message"),
	}

	// if err != nil {
	// 	response.Status = "error"
	// 	response.Message = "Bad request"
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	jsonResponse, _ := json.Marshal(response)
	// 	w.Write(jsonResponse)
	// 	log.Printf("Error decoding JSON: %v\n", err)
	// 	return
	// }
	// defer r.Body.Close()

	// Process the form data
	log.Printf("Received contact form submission: %+v\n", form)

	jsonResponse, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (a *App) ContactFormRedirectHandler(w http.ResponseWriter, r *http.Request) {
	redirectURL, _ := url.Parse("/#contactForm")
	query := redirectURL.Query()
	query.Set("status", "error")

	if r.Method != http.MethodPost {
		log.Printf("Method not allowed: %s\n", r.Method)
		redirectURL.RawQuery = query.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	if !a.ValidateCSRFToken(r.FormValue("csrf")) {
		log.Println("Invalid CSRF token")
		redirectURL.RawQuery = query.Encode()
		http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
		return
	}

	form := ContactForm{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Message: r.FormValue("message"),
	}

	// Process the form data
	log.Printf("Received contact form submission: %+v\n", form)

	// Set query parameters
	query.Set("status", "success")
	redirectURL.RawQuery = query.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusSeeOther)
}

func main() {
	var app App

	app.CacheTemplates("templates/index.html", "templates/about.html")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	if err := app.EnsureData(); err != nil {
		log.Printf("Error loading from API: %s", err.Error())
	}

	app.Home = Home{
		Title:            "Welcome To My Portfolio | Swaye Chateau",
		BlogUrl:          app.GetBlogUrl(),
		ProjectsUrl:      app.GetProjectsUrl(),
		Projects:         []Project{},
		Posts:            []Post{},
		Submitted:        false,
		SubmittedMessage: "",
		SubmittedClass:   "hidden",
	}

	app.About = About{
		Title:       "About Me | Swaye Chateau",
		BlogUrl:     app.GetBlogUrl(),
		ProjectsUrl: app.GetProjectsUrl(),
	}

	http.HandleFunc("/", app.HomeHandler)
	http.HandleFunc("/about", app.AboutHandler)
	http.HandleFunc("/contact", app.ContactFormHandler)
	log.Println("Starting server on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

func fetchPostsFromAPI(url, token string) (ApiResponse, error) {
	var response ApiResponse
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoClient/1.1)")

	resp, err := client.Do(req)
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

func generateCSRFToken() (CSRFToken, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Error generating CSRF token: %s\n", err)
		return CSRFToken{}, err
	}
	token := base64.URLEncoding.EncodeToString(b)
	return CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil
}

func urlFallback(url, fallback string) string {
	if url == "" {
		return fallback
	}
	return url
}
