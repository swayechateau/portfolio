# Portfolio

Welcome to the Portfolio project. This project is a personal portfolio website for showcasing projects, blog posts, and providing a contact form for visitors to reach out.

Website URL: [https://swaye.dev](https://swaye.dev)

## Purpose of the Project

The purpose of this project is to create an online presence for myself, displaying a collection of projects, blog posts, and offering an easy way for visitors to get in touch. It is built using Go and utilizes TailwindCSS for styling.

## Features

- **Home Page**: Introduction and a brief overview of the portfolio.
- **About Page**: Information about Swaye Chateau.
- **Projects Website**: Link to where I keep my projects (GitHub Profile).
- **Blog Website**: Link to my blog website.
- **Contact Form**: A form for visitors to send messages.
- **Custom 404 Page**: A user-friendly page for handling 404 errors.
- **Responsive Design**: Ensures the website is fully functional on all devices.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) (version 1.16 or later)
- [Node.js and npm](https://nodejs.org/) (for TailwindCSS)
- [Docker](https://www.docker.com/) (optional, for containerized deployment)
- [Make](https://www.gnu.org/software/make/) (optional, for using the Makefile)
- [Task](https://taskfile.dev/) (optional, for using the Taskfile)

### Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/swayechateau/portfolio.git
   cd portfolio
   ```

2. Install dependencies:

   ```sh
   npm install
   ```

3. Build the TailwindCSS styles:

   ```sh
   npm run build:css
   ```

4. Run the Go server:

   ```sh
   go run main.go
   ```

5. Open your browser and navigate to `http://localhost:5050`.

### Using Makefile

A `Makefile` is included to simplify running common commands. Here are some available targets:

- **build**: Builds the Go application
- **build-linux**: Builds the Go application for Linux
- **css-build**: Builds CSS files
- **css-watch**: Watches CSS files
- **docker-build**: Builds Docker image
- **docker-run**: Runs Docker container
- **docker-stop**: Stops Docker container
- **run**: Runs the Go application
- **dev**: Runs Go application in development mode
- **prod**: Runs Go application in production mode
- **stop**: Stops Docker container

Example usage:

```sh
make build
make build-linux
make css-build
make css-watch
make docker-build
make docker-run
make docker-stop
make run
make dev
make prod
make stop
```

### Using Taskfile

A `Taskfile.yml` is also included for those who prefer using Task.

Example usage:

```sh
task build
task build-linux
task css-build
task css-watch
task docker-build
task docker-run
task docker-stop
task run
task dev
task prod
task stop
```

### Docker Setup

1. Build the Docker image:

   ```sh
   docker build -t portfolio .
   ```

2. Run the Docker container:

   ```sh
   docker run -p 5050:5050 portfolio
   ```

## Project Structure

```
/portfolio-root
    /static
        /css
            style.css
        /js
            contact-form.js
            matrix.js
            navigation.js
        /img
            hero-deep-blue.jpg
            logo-swaye.png
            project-hulu-clone.png
    /storage
        app.log
        cache.json
    /templates
        index.html
        about.html
        404.html
    .env.example
    docker-compose.dev.yml
    docker-compose.yml
    Dockerfile
    main.go
    package.json
    style.css
    tailwind.config.js
    Makefile
    Taskfile.yml
    LICENSE
```

## Outstanding Tasks

- **SEO Optimization**: Improve SEO to increase visibility on search engines.
- **Accessibility**: Enhance accessibility to meet web standards.
- **Contact Form**: Fix the issue with the contact form not sending emails.
- **Typography**: Enhance the sites readability by choosing a better fontface.
- **Font Padding**: Enhance the about page font padding for better readability.
- **Projects API**: Add a live projects API and remove the hardcoded projects.
- **404 Error Routing**: Fix the 404 error routing.

## Contributing

Contributions are welcome! Please fork the repository and create a pull request with your changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
