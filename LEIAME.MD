# Server settings
SERVER_ADDRESS=:8080
GIN_MODE=debug

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=scheduling_db
DB_SSLMODE=disable

# JWT settings
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRE_HOURS=24

# CORS settings
CORS_ALLOWED_ORIGINS=*



.env.example:


# Server settings
SERVER_ADDRESS=:8080
GIN_MODE=debug  # debug, release, test

# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=scheduling_db
DB_SSLMODE=disable  # disable, require, verify-ca, verify-full

# JWT settings
JWT_SECRET=your-secret-key-change-this-in-production
JWT_EXPIRE_HOURS=24

# CORS settings
CORS_ALLOWED_ORIGINS=*  # comma-separated list of allowed origins, * for all

# Logging
LOG_LEVEL=debug  # debug, info, warn, error

# Rate limiting
RATE_LIMIT_REQUESTS=60
RATE_LIMIT_DURATION=1m  # time.Duration format (e.g., 1m, 1h)





TOML:

# .air.toml
# Hot reload configuration for development
# Install air: go install github.com/cosmtrek/air@latest

# Working directory
root = "."
tmp_dir = "tmp"

[build]
# Binary file yielded from go build
bin = "tmp/main"
# Command to build the app
cmd = "go build -o ./tmp/main ./cmd/api"
# Binary file with command line arguments
full_bin = "./tmp/main"
# Watch these filename extensions
include_ext = ["go", "tpl", "tmpl", "html"]
# Ignore these directories or files
exclude_dir = ["assets", "tmp", "vendor", ".git", "testdata"]
# Ignore files that match these patterns
exclude_file = []
# Exclude specific patterns
exclude_regex = ["_test.go"]
# Exclude unchanged files
exclude_unchanged = true
# Follow symlink
follow_symlink = true
# This log file places in your tmp_dir
log = "air.log"
# Poll files for changes instead of using fsnotify
poll = false
# Poll interval (defaults to the minimum interval of 500ms)
poll_interval = 500 # ms
# Stop running old binary when build errors occur
stop_on_error = true
# Send interrupt signal before killing process
send_interrupt = false
# Delay after each execution
delay = 1000 # ms
# Kill process if it runs for this duration in seconds
kill_delay = 500 # ms

[log]
# Show log time
time = true

[color]
# Customize each section's color
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"

[misc]
# Delete tmp directory on exit
clean_on_exit = true

