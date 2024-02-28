# GoServe
General Golang backend server designed to be generally extensible for building web apps.
See content_server directory for bulk of functionality.

## Functionality
User creation, authentication and login
- Highly secure thanks to JWT and PostgreSQL
- Safe, secret password storage with bcrypt password hashing
- OAuth coming soon

Dynamic Table Rendering
- Extremely flexible table customization, suitable for analytics applications.
  - Items per page
  - Number of page buttons
  - Data model
    - Headers
    - Columns & SQL queries
  - Display
    - Custom divs for each cell type
- Very clean, modular codebase for easy maintenance and extension
- Filters and UI resizing/interaction coming soon

Single Page Application
- Fast and responsive pages thanks to partial page loading with HTMX AJAX
- Flexible and easy to maintain with templated html component responses for dynamic rendering and modular codebase
- Responsive page design with Tailwind CSS
- Interactive elements with javascript and alpine.js

Interactive graphs coming soon (with charts.js)

Specs:
- Front end
    - windmill-dashboard-master template
    - htmx
    - alpine.js
- Back end
    - golang
    - echo
    - postgres
    - python (for model serving)


## create your postgres instance
```
chmod +x postgres.sh
# starts postgres instance, builds tables, populates sample data
./postgres.sh
```

## create and start the server
```
go build
./main
```


## Roadmap
- Graph visualizations
- Table filters, resizing, etc (UI, already have back end)
- OAuth
- 