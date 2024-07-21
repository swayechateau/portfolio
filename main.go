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

// ContactForm represents a contact form submission.
type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// Project represents a project with its details.
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

// Post represents a blog post.
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

// ApiResponse represents the structure of the API response.
type ApiResponse struct {
	// Recent represents a list of recent posts.
	// It is a slice of Post structs and is tagged with `json:"recent"` for JSON serialization.
	Recent []Post `json:"recent"`
	// Featured represents a list of featured posts.
	// It is a field of type []Post and is tagged with `json:"featured"`.
	Featured []Post `json:"featured"`
}

// Database represents the structure of the database.
type Database struct {
	Projects []Project   `json:"projects"`
	Posts    ApiResponse `json:"posts"`
}

// App represents the main application struct.
type App struct {
	CSRFToken     CSRFToken
	Database      Database
	TemplateCache map[string]*template.Template
	Home          Home
	About         About
}

// Home represents the home page of the website.
type Home struct {
	Title            string    // Title is the title of the home page.
	BlogUrl          string    // BlogUrl is the URL of the blog website.
	ProjectsUrl      string    // ProjectsUrl is the URL of the projects website.
	Projects         []Project // Projects is a list of projects.
	Posts            []Post    // Posts is a list of blog posts.
	Submitted        bool      // Submitted indicates whether a form has been submitted.
	SubmittedMessage string    // SubmittedMessage is the message to display after form submission.
	SubmittedClass   string    // SubmittedClass is the CSS class to apply after form submission.
	CSRF             string    // CSRF is the Cross-Site Request Forgery token.
}

// About represents information about a person or organization.
type About struct {
	Title       string // The title of the about page.
	BlogUrl     string // The URL of the blog website.
	ProjectsUrl string // The URL of the projects website.
}

// CSRFToken represents a Cross-Site Request Forgery (CSRF) token.
type CSRFToken struct {
	Token     string    // The CSRF token value.
	ExpiresAt time.Time // The expiration time of the CSRF token.
}

// SaveToCache saves the database to a cache file in JSON format.
// It creates a new file named "cache.json" and writes the JSON representation of the database to it.
// The database is marshaled using JSON indentation for readability.
// If any error occurs during the file creation, marshaling, or writing process, an error is returned.
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

// UpdateCacheIfNewData updates the cache with new data if available.
// It fetches new data from the API using the provided blog URL and token.
// If new data is found, it updates the cache and returns nil.
// If no new data is found, it logs a message and returns nil.
// If there is an error while fetching new data or updating the cache, it returns an error.
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

// LoadFromCache loads data from a cache file into the Database.
// It opens the cache file, decodes the JSON data into the Database object,
// and returns an error if any error occurs during the process.
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

// FetchFromAPI fetches data from an API and saves it to the database.
// It takes a blog URL and an authentication token as input parameters.
// It returns an error if there was an issue fetching the posts or projects.
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

// fetchPosts fetches posts from the API and updates the database with the response.
// It takes a URL and a token as parameters and returns an error if fetching posts fails.
func (db *Database) fetchPosts(url, token string) error {
	apiResponse, err := fetchPostsFromAPI(url, token)
	if err != nil {
		return fmt.Errorf("error fetching posts: %w", err)
	}
	db.Posts = apiResponse
	return nil
}

// fetchProjects fetches projects from the API and updates the database with the fetched projects.
// It returns an error if there was an issue fetching the projects.
func (db *Database) fetchProjects() error {
	projects, err := fetchProjectsFromAPI()
	if err != nil {
		return fmt.Errorf("error fetching projects: %w", err)
	}
	db.Projects = projects
	return nil
}

// CacheTemplates caches the parsed templates for the given filenames.
// It takes a variadic parameter filenames, which represents the paths of the template files.
// If the TemplateCache is nil, it initializes it as an empty map.
// For each filename, it parses the template file using template.ParseFiles function.
// If there is an error parsing the template, it logs a fatal error.
// Finally, it stores the parsed template in the TemplateCache map with the filename as the key.
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

// FetchData fetches data from the blog API and updates the cache in the database.
// It returns an error if the data fetching or cache update fails.
func (a *App) FetchData() error {
	if err := a.Database.UpdateCacheIfNewData(a.GetBlogAPI(), a.GetBlogApiToken()); err != nil {
		return fmt.Errorf("could not fetch data: %w", err)
	}
	return nil
}

// EnsureData ensures that the data is loaded into the App's database.
// If the data is not available in the cache, it fetches it from the API.
// It then updates the cache if new data is available.
func (a *App) EnsureData() error {
	if err := a.Database.LoadFromCache(); err != nil {
		log.Printf("Error loading from cache: %s, fetching from API", err)
		return a.Database.FetchFromAPI(a.GetBlogAPI(), a.GetBlogApiToken())
	}
	log.Println("Loaded data from cache")

	return a.Database.UpdateCacheIfNewData(a.GetBlogAPI(), a.GetBlogApiToken())
}

// GetBlogUrl returns the URL of the blog.
// It first checks the value of the environment variable "BLOG_URL".
// If the environment variable is not set, it falls back to "http://localhost:8000".
func (a *App) GetBlogUrl() string {
	return urlFallback(
		os.Getenv("BLOG_URL"),
		"http://localhost:8000",
	)
}

// GetBlogAPI returns the URL of the blog API. It first checks the value of the "BLOG_API" environment variable.
// If the environment variable is not set, it falls back to the default URL "http://localhost:8000/api/posts".
func (a *App) GetBlogAPI() string {
	return urlFallback(
		os.Getenv("BLOG_API"),
		"http://localhost:8000/api/posts",
	)
}

// GetBlogApiToken returns the API token for the blog.
func (a *App) GetBlogApiToken() string {
	return os.Getenv("BLOG_API_TOKEN")
}

// GetBlogClientId returns the client ID for the blog.
func (a *App) GetBlogClientId() string {
	return os.Getenv("BLOG_CLIENT_ID")
}

// GetBlogClientSecret returns the client secret for the blog.
func (a *App) GetBlogClientSecret() string {
	return os.Getenv("BLOG_CLIENT_SECRET")
}

// GetBlogApiAuthToken retrieves the API authentication token for the blog.
// It sends a POST request to the blog's OAuth token endpoint with the client credentials,
// and returns the access token received in the response.
// If any error occurs during the process, it will be logged and a fatal error will be thrown.
// The access token is also printed to the console for debugging purposes.
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

// GetProjectsUrl returns the URL for retrieving projects.
// It first checks the value of the PROJECTS_URL environment variable.
// If the environment variable is not set, it falls back to "http://localhost:8000/projects".
func (a *App) GetProjectsUrl() string {
	return urlFallback(
		os.Getenv("PROJECTS_URL"),
		"http://localhost:8000/projects",
	)
}

// GetProjectsAPI returns the URL of the projects API.
// It first checks the value of the PROJECTS_API environment variable.
// If the environment variable is not set, it falls back to the default URL "http://localhost:8000/api/projects".
func (a *App) GetProjectsAPI() string {
	return urlFallback(
		os.Getenv("PROJECTS_API"),
		"http://localhost:8000/api/projects",
	)
}

// GetCSRFToken returns the CSRF token for the App.
// If the CSRF token is empty, it generates a new one using the generateCSRFToken function.
func (a *App) GetCSRFToken() string {
	if a.CSRFToken.Token == "" {
		a.CSRFToken, _ = generateCSRFToken()
	}
	return a.CSRFToken.Token
}

// NewCSRFToken generates a new CSRF token for the App.
// It calls the generateCSRFToken function to generate the token
// and assigns it to the App's CSRFToken field.
// It returns the generated CSRF token.
func (a *App) NewCSRFToken() string {
	a.CSRFToken, _ = generateCSRFToken()
	return a.CSRFToken.Token
}

// ValidateCSRFToken checks if the provided CSRF token is valid.
// It returns true if the token matches the stored token and has not expired; otherwise, it returns false.
func (a *App) ValidateCSRFToken(token string) bool {
	if a.CSRFToken.Token == "" {
		return false
	}
	return a.CSRFToken.Token == token && time.Now().Before(a.CSRFToken.ExpiresAt)
}

// HomeHandler handles the HTTP request for the home page.
// It sets the CSRF token, fetches data, and populates the home page with projects and recent posts.
// It also checks for query parameters related to form submission status and updates the home page accordingly.
// If the template is not found or there is an error rendering the template, it returns an HTTP error.
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

// AboutHandler handles the HTTP request for the about page.
// It loads the "templates/about.html" template from the App's TemplateCache
// and renders it with the data stored in the App's About field.
// If the template or rendering fails, it returns an HTTP 500 Internal Server Error.
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

// ContactFormHandler handles the HTTP request for the contact form.
// It checks the Accept header of the request and calls the appropriate handler based on the content type.
// If the Accept header is "application/json", it calls the ContactFormJSONHandler.
// Otherwise, it calls the ContactFormRedirectHandler.
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

// ContactFormJSONHandler handles the JSON request for the contact form.
// It validates the request, processes the form data, and returns a JSON response.
// If the request method is not POST, it returns an error response with status code 405 (Method Not Allowed).
// If the CSRF token is invalid, it returns an error response with status code 403 (Forbidden).
// Otherwise, it processes the form data, logs the submission, and returns a success response with status code 200 (OK).
//
// Parameters:
// - w: The http.ResponseWriter used to write the response.
// - r: The *http.Request representing the incoming request.
//
// Example usage:
//
//	http.HandleFunc("/contact", app.ContactFormJSONHandler)
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

	if !a.ValidateCSRFToken(r.FormValue("csrf")) {
		response.Status = "error"
		response.Message = "Invalid CSRF token"
		w.WriteHeader(http.StatusForbidden)
		jsonResponse, _ := json.Marshal(response)
		w.Write(jsonResponse)
		log.Println("Invalid CSRF token")
		return
	}

	form := ContactForm{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Message: r.FormValue("message"),
	}

	// Process the form data
	log.Printf("Received contact form submission: %+v\n", form)

	jsonResponse, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// ContactFormRedirectHandler handles the redirection after submitting a contact form.
// It takes in the HTTP response writer and request as parameters.
// If the request method is not POST, it redirects to the contact form page with an error status.
// If the CSRF token is invalid, it redirects to the contact form page with an error status.
// If the form data is valid, it processes the form data and redirects to the contact form page with a success status.
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

// main is the entry point of the application.
// It initializes the `app` variable, caches the templates,
// sets the port, ensures data is loaded from the API,
// initializes the `Home` and `About` structs,
// sets up the HTTP request handlers, and starts the server.
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", app.HomeHandler)
	http.HandleFunc("/about", app.AboutHandler)
	http.HandleFunc("/contact", app.ContactFormHandler)
	log.Println("Starting server on :" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}

// fetchPostsFromAPI fetches posts from the specified API endpoint.
// It sends a GET request to the provided URL with the given token as authorization.
// The function returns an ApiResponse and an error if any occurred.
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

// fetchProjectsFromAPI fetches projects from an API.
// It returns a slice of Project structs and an error, if any.
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

// generateCSRFToken generates a CSRF token.
// It uses the crypto/rand package to generate a random byte slice,
// encodes it using base64.URLEncoding, and returns a CSRFToken struct
// containing the token and its expiration time.
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

// urlFallback returns the given URL if it is not empty, otherwise it returns the fallback URL.
func urlFallback(url, fallback string) string {
	if url == "" {
		return fallback
	}
	return url
}
