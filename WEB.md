# Say Less - WEB.md

> Say less. Make the computer understand more.

Say Less is a general-purpose programming language with a built-in web compilation pipeline. This document covers everything about the application: the language, the web framework, the HTTP server, the CLI, architecture, deployment, and all features.

---

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [Language Reference](#language-reference)
- [Web Development](#web-development)
- [HTTP Server](#http-server)
- [Built-in Modules](#built-in-modules)
- [Built-in Functions](#built-in-functions)
- [Project Structure](#project-structure)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Compilation Pipeline](#compilation-pipeline)
- [Testing](#testing)
- [Deployment](#deployment)
- [Examples](#examples)

---

## Overview

Say Less is both a **programming language** and a **web framework**. It operates in two modes:

1. **Interpreter mode** -- Lex, parse, and tree-walk evaluate `.sl` source files. This is the current execution model.
2. **Web compiler mode** -- Lex, parse, compile to a Web IR, and generate HTML/CSS/JS output. This allows building websites without writing HTML, CSS, or JavaScript.

The language is implemented in Go with **zero external dependencies**. It compiles to a single standalone binary (`sale`) with no runtime requirements.

**Key design principles:**

- Minimal syntax, maximal meaning
- Three block syntax styles (curly braces, colon+indent, bare indent)
- English-like aliases (`when`, `otherwise`, `loop`, `each`, `give`)
- Immutable-by-default variables
- First-class functions with closures
- Errors as values, not exceptions
- npm/pip package integration from Say Less code

---

## Installation

### Pre-built Binaries

Download from the [Releases](https://github.com/Itshardtofindagoodname/say_less/releases) page — or just take the pre-built `.exe` right out of this repository's root. No Go compiler required.

| Platform | File |
|----------|------|
| Windows (x64) — one installer exe | `say_less.exe` |
| macOS (Intel) | `sale-macos-amd64.tar.gz` |
| macOS (Apple Silicon) | `sale-macos-arm64.tar.gz` |
| Linux (x64) | `sale-linux-amd64.tar.gz` |
| Linux (ARM64) | `sale-linux-arm64.tar.gz` |

#### Windows — GUI installer (recommended)

On Windows just grab **`say_less.exe`** from the repository root or the [Releases](https://github.com/Itshardtofindagoodname/say_less/releases) page.

1. Download **`say_less.exe`** (in this repo's root, or from [Releases](https://github.com/Itshardtofindagoodname/say_less/releases)).
2. Double-click **`say_less.exe`**.
3. A setup window opens (same style as the Python installer). Keep **Add `sale` to your PATH** checked and click **Install**.
4. The installer copies `sale.exe` to your chosen location. If adding to the system PATH, Windows asks for administrator permission once (UAC) — accept it, and `sale` is added to your PATH automatically.
5. Open a new terminal and verify:

```powershell
sale --version
# sale 0.1.0
```

#### Windows — zip + install.bat (optional)

1. Extract the zip file.
2. Right-click `install.bat` and select **Run as administrator**.
3. The installer copies `sale.exe` to `C:\Program Files\SayLess\bin\` and adds it to your system PATH.

#### macOS / Linux

1. Extract the archive.
2. Run the installer:

```bash
chmod +x install.sh
./install.sh
```

The installer copies the binary to `~/.sayless/bin/` and adds it to your PATH.

3. Start a new terminal, or run:

```bash
source ~/.bashrc
```

#### Verify

```bash
sale --version
# sale 0.1.0
```

### Build from Source

Requires Go 1.21+.

```bash
git clone https://github.com/Itshardtofindagoodname/say_less.git
cd say_less
go build -o sale ./cmd/sale/        # macOS / Linux
go build -o sale.exe ./cmd/sale/    # Windows
```

Move the binary to a directory in your PATH:

```bash
sudo mv sale /usr/local/bin/        # macOS / Linux
```

Or on Windows:

```powershell
$env:PATH += ";C:\path\to\your\binary"
```

### Cross-Platform Build

Run `build.bat` to build for all platforms. Binaries go to `dist/`:

- `say_less.exe` (the one Windows installer — embeds `sale.exe`, requests admin via UAC only when updating PATH)
- `sale-windows-amd64.zip`
- `sale-macos-amd64.tar.gz`
- `sale-macos-arm64.tar.gz`
- `sale-linux-amd64.tar.gz`
- `sale-linux-arm64.tar.gz`

`release.bat` packages everything (including `say_less.exe` for Windows) into the final archives.

---

## Quick Start

```bash
# Create a new project
sale new hello
cd hello

# Run it
sale run src/main.sl
```

### Hello World

Create a file `hello.sl`:

```
print "Hello, World!"
```

Run it:

```bash
sale run hello.sl
```

### REPL

```bash
sale repl
> 1 + 1
2
> name = "Jeet"
> print "Hello, " + name
Hello, Jeet
> exit
```

---

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `sale run <file>` | Run a Say Less source file |
| `sale build [--release]` | Build the project. `--release` for optimized builds |
| `sale dev` | Start the development server with file watching |
| `sale test` | Run all test files in the project |

#### `sale run <file>`

Runs a single `.sl` file from start to finish.

```bash
sale run src/main.sl
```

#### `sale build [--release]`

Compiles the project. Without `--release`, builds a debug version. With `--release`, produces an optimized binary.

```bash
sale build
sale build --release
```

#### `sale dev`

Starts a development server that watches for file changes and re-runs automatically. Looks for `src/main.sl` or `main.sl`.

```bash
sale dev
```

#### `sale test`

Discovers and runs all test files. Files are found by looking for `.sl` files containing "test" in the path. Falls back to all `.sl` files if no test-specific files are found.

```bash
sale test
```

### Project Commands

| Command | Description |
|---------|-------------|
| `sale new <name>` | Create a new empty project |
| `sale create --web <name>` | Create a web project with HTTP server scaffold |
| `sale create --backend <name>` | Create a backend/API project scaffold |
| `sale create --system <name>` | Create a systems-level project scaffold |
| `sale init [name]` | Initialize a project in the current directory |

#### `sale new <name>`

Creates a new project directory with `src/main.sl`, `sale.toml`, and `README.md`.

```bash
sale new myapp
cd myapp
sale run src/main.sl
```

#### `sale create --web <name>`

Creates a web project with HTTP server boilerplate, Dockerfile, and compose.yml.

```bash
sale create --web mysite
cd mysite
sale dev
```

#### `sale create --backend <name>`

Creates a backend API project with HTTP and JSON scaffolding, Dockerfile, and compose.yml.

```bash
sale create --backend myapi
cd myapi
sale dev
```

#### `sale create --system <name>`

Creates a systems-level project with a minimal setup.

```bash
sale create --system mytool
cd mytool
sale build --release
```

#### `sale init [name]`

Initializes a `sale.toml` and `src/main.sl` in the current directory. Uses the directory name if no name is given.

```bash
mkdir myproject
cd myproject
sale init myproject
```

### Package Commands

| Command | Description |
|---------|-------------|
| `sale add <package>` | Add a dependency to `sale.toml` |
| `sale remove <package>` | Remove a dependency |
| `sale install` | Install dependencies listed in `sale.toml` |

#### `sale add <package>`

Adds a package to the `[dependencies]` section of `sale.toml`.

```bash
sale add http
```

#### `sale remove <package>`

Removes a package from the project.

```bash
sale remove http
```

#### `sale install`

Reads `sale.toml` and installs listed dependencies.

```bash
sale install
```

### Tooling Commands

| Command | Description |
|---------|-------------|
| `sale fmt [file]` | Format source files. Without arguments, formats all `.sl` files |
| `sale check [file]` | Check code for parse/lexer errors without running it |
| `sale repl` | Start an interactive Read-Eval-Print Loop |
| `sale clean` | Remove `build/` and `.sale/` directories |
| `sale doctor` | Print installation health info (version, platform, project status) |

#### `sale fmt [file]`

Formats `.sl` source files. With no arguments, walks the current directory and formats all `.sl` files.

```bash
sale fmt
sale fmt src/main.sl
```

#### `sale check [file]`

Lexes and parses a file to check for errors without executing it. Prints "No errors found." on success.

```bash
sale check
sale check src/main.sl
```

#### `sale repl`

Starts an interactive REPL with a persistent environment. Variables and functions defined in one line carry over to the next.

```bash
sale repl
> 1 + 1
2
> name = "Jeet"
> print "Hello, " + name
Hello, Jeet
> exit
```

#### `sale clean`

Removes build output directories (`build/` and `.sale/`).

```bash
sale clean
```

#### `sale doctor`

Prints diagnostic info: version, platform, whether a `sale.toml` is found, and installation status.

```bash
sale doctor
# Say Less Doctor
# ===============
# Version: 0.1.0
# Platform: linux/amd64
# Project: found sale.toml
# Installation: OK
```

### Options

| Option | Description |
|--------|-------------|
| `sale --version`, `-v` | Print the version number |
| `sale --help`, `-h` | Print the help message |

---

## Language Reference

### Variables

```
name = "Jeet"
age = 30
active = true
pi = 3.14
mut counter = 0
counter += 1
```

Variables are **immutable by default**. Use `mut` to declare a mutable variable. Use `const` for compile-time constants.

```
const MAX_SIZE = 1024
```

### Augmented Assignment

```
counter += 1
counter -= 1
counter *= 2
counter /= 3
counter %= 5
```

### Types

#### Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `integer` | 64-bit signed integer | `42` |
| `float` | 64-bit IEEE 754 float | `3.14` |
| `boolean` | true or false | `true` |
| `string` | Immutable UTF-8 string | `"hello"` |
| `bytes` | Raw byte sequence | `bytes(10)` |
| `none` | Absence of value | `none` |

#### Compound Types

| Type | Description | Example |
|------|-------------|---------|
| `list<T>` | Dynamic array | `[1, 2, 3]` |
| `map<K, V>` | Hash map | `{"a": 1}` |
| `struct` | Named record | `struct Point` |
| `function` | First-class function | `fn(a, b) a + b` |

#### Literals

```
# Integers
42
1_000_000
0xFF
0b1010
0o77

# Floats
3.14
0.5
1.0e10

# Strings
"hello world"
'hello world'
"line one\nline two"
"she said \"hello\""

# Multi-line strings
"""
This is a
multi-line string.
"""

# Booleans
true
false

# None
none
```

#### Type Inference

Types are inferred from context. Explicit types are optional:

```
age = 20              # inferred as integer
name = "Jeet"         # inferred as string
active = true         # inferred as boolean
items = [1, 2, 3]    # inferred as list<integer>
port: integer = 8080  # explicit type annotation
```

### Functions

#### Definition

```
fn greet(name) {
    print "Hello, " + name
}

fn add(a, b) {
    return a + b
}

# Bare indent (no colon, no braces)
fn multiply(a, b)
    return a * b
```

#### Calling

Parentheses are optional:

```
greet "Jeet"
result = add 2 3

# Also valid with parentheses:
greet("Jeet")
result = add(2, 3)
```

#### Default Arguments

```
fn greet(name, greeting = "Hello")
    print greeting + ", " + name

greet "Jeet"                    # Hello, Jeet
greet "Jeet", greeting = "Hi"   # Hi, Jeet
```

#### First-Class Functions

```
fn apply(f, value)
    return f(value)

double = fn(x) x * 2
result = apply double 5    # 10
```

#### Closures

```
fn counter()
    mut count = 0
    return fn()
        count += 1
        return count

c = counter()
c()  # 1
c()  # 2
```

### Control Flow

Blocks support **three styles**. All three are equivalent -- use whichever you prefer:

- **Curly braces** -- `if cond { ... }` (C/Java style)
- **Colon + indent** -- `if cond: ...` (Python style)
- **Bare indent** -- `if cond\n    ...` (no colon needed)

The colon after `if`, `while`, `for`, `fn`, `struct`, and `test` is always optional.

#### If / Else

```
# Curly braces
if age >= 18 {
    print "adult"
} else {
    print "child"
}

# Colon + indent
if age >= 18:
    print "adult"
else:
    print "child"

# Bare indent
if age >= 18
    print "adult"
else
    print "child"

# English aliases: when = if, otherwise = else
when age >= 18 {
    print "adult"
} otherwise {
    print "child"
}
```

#### While

```
mut i = 0
while i < 10 {
    print i
    i += 1
}

# Bare indent
mut i = 0
while i < 10
    print i
    i += 1

# Alias: loop = while
mut i = 0
loop i < 10 {
    print i
    i += 1
}
```

#### For

```
for item in items {
    print item
}

# Bare indent
for item in items
    print item

# Alias: each = for
each item in items {
    print item
}
```

#### Break / Continue

```
for item in items {
    if item == 0 {
        continue
    }
    if item > 100 {
        break
    }
    print item
}
```

### Data Structures

#### Lists

```
items = [1, 2, 3]
items.push(4)
print items.length

# List methods
items.push(4)              # [1, 2, 3, 4]
items.pop()                # returns 4, list is [1, 2, 3]
items.length               # 3
items.join(", ")           # "1, 2, 3"
items.map(fn(x) x * 2)    # [2, 4, 6]
items.filter(fn(x) x > 1) # [2, 3]
items.find(2)              # 2
```

#### Maps

```
person = {name: "Jeet", age: 30}
print person.name

# Map methods
keys(person)    # ["name", "age"]
values(person)  # ["Jeet", 30]
has(person, "name")  # true
```

#### Structs

```
struct Point
    x integer
    y integer

p = Point(x: 10, y: 20)
print p.x    # 10
print p.y    # 20
```

### String Methods

```
name = "hello world"
print name.upper()              # HELLO WORLD
print name.lower()              # hello world
print name.length               # 11
print name.trim()               # "hello world"
print name.contains("world")    # true
print name.split(" ")           # ["hello", "world"]
print name.replace("hello", "hey")
print name.starts_with("hello") # true
print name.ends_with("world")   # true
print name.index_of("world")    # 6
print name.substring(0, 5)      # "hello"
print name.chars()              # ["h", "e", "l", "l", "o", ...]
print name.to_int()             # (if numeric string)
```

### Operators

| Operator | Description |
|----------|-------------|
| `+` `-` `*` `/` `%` | Arithmetic |
| `==` `!=` `<` `>` `<=` `>=` | Comparison |
| `=` `+=` `-=` `*=` `/=` `%=` | Assignment |
| `.` | Member access |
| `and` `or` `not` | Logical |

**Operator precedence** (highest to lowest):

1. `.` member access
2. Function call `f(x)`
3. Unary `-`, `+`, `not`
4. `*`, `/`, `%`
5. `+`, `-`
6. `==`, `!=`, `<`, `>`, `<=`, `>=`
7. `and`
8. `or`
9. `=` assignment

### Comments

```
# This is a line comment
```

No block comments. Keep comments simple.

### Significant Newlines

Newlines separate statements. Statements are not terminated by semicolons.

A line is continued if it ends with:
- An operator
- An open parenthesis `(`, bracket `[`, or brace `{`
- A comma `,`

---

## Web Development

Say Less includes a first-class web compilation pipeline that generates HTML, CSS, and JavaScript from `.sl` source files. You write Say Less; the compiler produces a complete website.

### Page Declarations

Pages define routes and their content:

```
page "/"
    main
        h1 "My Website"
        p "Welcome to Say Less Web."
```

Each `page` block becomes a separate HTML file. The route string defines the URL path.

### HTML Elements

HTML tags are used directly in Say Less code. Attributes use the `attribute value` syntax:

```
page "/"
    main
        h1 class "title" "Hello"
        p "This is a paragraph."
        a href "/about" "Go to About"
        img src "/logo.png" alt "Logo"
        input type "text" placeholder "Enter name"
```

### Components

Components are reusable UI building blocks with parameters:

```
component Card(title, description)
    article class "card"
        h2 title
        p description
```

Components are called like functions with named arguments:

```
page "/"
    Card(
        title: "First Post",
        description: "This is my first post."
    )
    Card(
        title: "Second Post",
        description: "This is my second post."
    )
```

### State Management

Reactive state variables automatically update the DOM when changed:

```
page "/"
    state count = 0

    main
        h1 "Counter: " + count

        button "Click me"
            on click
                count += 1
```

State is declared with the `state` keyword inside a page or component. The compiler generates JavaScript runtime code that tracks state changes and updates the DOM reactively.

### Event Handlers

Events are attached to elements using `on <event>`:

```
button "Click me"
    on click
        count += 1

input type "text"
    on input
        value = input.value

form
    on submit
        print "Form submitted!"
```

Supported events: `click`, `input`, `submit`, `keydown`, `keyup`, `mouseover`, `mouseout`, and any standard DOM event.

### Conditional Rendering

Use `if`/`else` inside pages to conditionally show content:

```
page "/"
    state loggedIn = false

    main
        if loggedIn
            p "Welcome back!"
        else
            p "Please log in."
```

### Loop Rendering

Use `for` to render lists of items:

```
page "/"
    state items = ["Apple", "Banana", "Cherry"]

    main
        ul
            for item in items
                li item
```

### Inline Styles

Style blocks can be placed inside elements:

```
div
    style
        background-color: blue
        padding: 20px
        color: white

    p "Styled content"
```

### Component Composition

Components can be nested and composed:

```
component Header(title)
    header
        h1 title

component Footer
    footer
        p "2026 Say Less"

page "/"
    Header(title: "My Site")
    main
        p "Content goes here."
    Footer()
```

### Text Interpolation

String concatenation with the `+` operator works for dynamic content:

```
state name = "World"
h1 "Hello, " + name + "!"
```

### Web Compiler Output

The web compiler produces three outputs:

1. **HTML** -- The complete HTML document with elements, attributes, and structure
2. **CSS** -- Scoped styles from component style blocks
3. **JavaScript** -- The Say Less Web Runtime (`SL` object) that handles:
   - Reactive state management
   - Event handler binding
   - DOM updates
   - Template rendering

The generated HTML includes:
- `data-reactive="true"` on elements that reference state variables
- `data-on-<event>` attributes for event handlers
- `data-component` and `data-instance` for component identification
- HTML comments for conditionals, loops, and awaits (server-side markers)

### npm Package Integration

Pull in JavaScript packages with `use npm:<package>`:

```
use npm:gsap

page "/"
    script
        gsap.fromTo("#title", {y: 60, opacity: 0}, {y: 0, opacity: 1})
```

The compiler auto-injects `<script>` tags for npm packages.

### pip Package Integration

Python packages can be referenced:

```
use pip:flask
```

---

## HTTP Server

Say Less includes a built-in HTTP server for backend development.

### Starting the Server

```
server on 8080
```

### Defining Routes

```
server on 8080

get "/"
    return "Hello from Say Less!"

get "/users"
    users = [{name: "Alice", age: 30}, {name: "Bob", age: 25}]
    return json users

post "/echo"
    body = request.body
    return json body

put "/users/:id"
    return json {status: "updated"}

delete "/users/:id"
    return json {status: "deleted"}
```

Supported HTTP methods: `get`, `post`, `put`, `delete`, `patch`.

### Request Object

Inside route handlers, the `request` object provides:

```
request.method   # "GET", "POST", etc.
request.path     # "/users"
request.body     # Request body (for POST/PUT)
request.headers  # Request headers
```

### JSON Responses

Use `json` to return JSON responses:

```
get "/api/data"
    data = {name: "Jeet", items: [1, 2, 3]}
    return json data
```

### Static File Serving

The server automatically serves static files from:

- `src/styles/` -- CSS files accessible at `/styles/*`
- `node_modules/` -- npm packages accessible at `/node_modules/*`
- `public/` -- Public directory files accessible at `/public/*`

### Full Example

```
use http

server on 8080

get "/"
    return """
        <!DOCTYPE html>
        <html>
        <head><title>My App</title></head>
        <body>
            <h1>Hello!</h1>
            <script src="/node_modules/gsap/dist/gsap.min.js"></script>
        </body>
        </html>
        """

get "/api/status"
    return json {status: "ok", version: "0.1.0"}

post "/api/echo"
    return json request.body
```

---

## Built-in Modules

### `math`

| Function | Description |
|----------|-------------|
| `math.sqrt(x)` | Square root |
| `math.pow(base, exp)` | Power |
| `math.floor(x)` | Floor |
| `math.ceil(x)` | Ceiling |
| `math.sin(x)` | Sine |
| `math.cos(x)` | Cosine |
| `math.tan(x)` | Tangent |
| `math.log(x)` | Natural logarithm |
| `math.pi` | Pi constant |
| `math.e` | Euler's number |

```
print math.sqrt(144)    # 12
print math.pi           # 3.141592653589793
print math.floor(3.7)   # 3
print math.ceil(3.2)    # 4
```

### `json`

| Function | Description |
|----------|-------------|
| `json.encode(value)` | Encode a value to JSON string |
| `json.decode(string)` | Decode a JSON string to a value |

```
data = json.decode("{\"x\": 42}")
print data.x

encoded = json.encode({a: 1, b: 2})
print encoded
```

### `strings`

| Function | Description |
|----------|-------------|
| `strings.contains(haystack, needle)` | Check if string contains substring |
| `strings.split(str, delimiter)` | Split string by delimiter |
| `strings.join(list, delimiter)` | Join list with delimiter |
| `strings.replace(str, old, new)` | Replace occurrences |
| `strings.has_prefix(str, prefix)` | Check prefix |
| `strings.has_suffix(str, suffix)` | Check suffix |

```
print strings.contains("hello world", "world")  # true
parts = strings.split("a,b,c", ",")             # ["a", "b", "c"]
joined = strings.join(["x", "y"], "-")          # "x-y"
```

### `io`

| Function | Description |
|----------|-------------|
| `io.read(path)` | Read file contents |
| `io.write(path, content)` | Write content to file |
| `io.exists(path)` | Check if file exists |

```
content = io.read("file.txt")
io.write("output.txt", "hello")
exists = io.exists("file.txt")  # true or false
```

### `os`

| Function | Description |
|----------|-------------|
| `os.args` | Command-line arguments |
| `os.env(name)` | Get environment variable |
| `os.clock()` | Current time in seconds |

```
args = os.args
val = os.env("HOME")
t = os.clock()
```

---

## Built-in Functions

| Function | Description |
|----------|-------------|
| `print(...)` | Print values to stdout (adds newline) |
| `len(x)` | Get length of string, list, or map |
| `to_string(x)` | Convert to string |
| `to_int(x)` | Convert to integer |
| `to_float(x)` | Convert to float |
| `type_of(x)` | Get type name as string |
| `exit(code)` | Exit with status code |
| `range(n)` / `a..b` | Create a range for iteration |
| `push(list, item)` | Add item to list |
| `pop(list)` | Remove and return last item |
| `keys(map)` | Get list of map keys |
| `values(map)` | Get list of map values |
| `has(map, key)` | Check if key exists |
| `slice(x, start, end)` | Slice a string or list |
| `char(n)` | Convert integer to character |
| `ord(s)` | Convert character to integer |
| `abs(x)` | Absolute value |
| `min(a, b)` | Minimum of two values |
| `max(a, b)` | Maximum of two values |
| `random()` | Random float between 0 and 1 |
| `sleep(seconds)` | Sleep for N seconds |
| `time()` | Unix timestamp |
| `clock()` | High-precision time in seconds |
| `fail(msg)` | Create an error value |
| `ok(val)` | Wrap a value as success |
| `err(msg)` | Create an error value |
| `assert(cond)` | Assert a condition is truthy |

---

## Project Structure

### Standard Layout

```
myproject/
    sale.toml          # Project configuration
    src/
        main.sl        # Entry point
    tests/
        all_test.sl    # Test files
    public/            # Static files
    build/             # Build output
```

### Web Project Layout

```
mywebsite/
    sale.toml
    src/
        main.sl        # HTTP server + routes
        pages/
            index.sl   # Page components
            about.sl
        styles/
            base.css   # Base styles
            main.css   # App styles
    public/            # Static assets
    build/
        web/           # Compiled HTML/CSS/JS output
    Dockerfile
    compose.yml
```

### Full Example Project Layout (test-web)

```
test-web/
    src/
        main.sl                    # Entry point with HTTP server
        pages/
            index.sl               # Home page
        styles/
            base.css               # Base styles
            main.css               # Main styles
            page.css               # Page-specific styles
    sale.toml                      # Project config
    Dockerfile                     # Docker build file
    compose.yml                    # Docker Compose config
    README.md
```

---

## Configuration

### sale.toml

The project configuration file:

```toml
name = "myproject"
version = "0.1.0"

[dependencies]
http = "0.1"
json = "0.1"
npm:gsap = "latest"
npm:lodash = "^4.17"
pip:flask = "latest"
```

#### Dependency Types

- **Built-in modules**: `http`, `json`, `math`, `strings`, `io`, `os`
- **npm packages**: `npm:<package-name>` -- auto-injects script tags in web mode
- **pip packages**: `pip:<package-name>` -- exposes Python packages

### Environment Variables

The application uses Go's `os.Getenv()` for environment variables. No `.env` file is used.

---

## Architecture

### System Components

```
Source (.sl)
    |
    v
Lexer -> Tokens
    |
    v
Parser -> AST
    |
    v
Evaluator -> Result          (Interpreter mode)
    |
    v
Web IR -> HTML/CSS/JS        (Web compiler mode)
```

### Lexer

Tokenizes source code into tokens. Located in `internal/lexer/lexer.go` (713 lines).

**Token types:**
- Keywords: `fn`, `if`, `else`, `while`, `for`, `return`, `struct`, `test`, `assert`, `use`, `as`, `mut`, `const`, `and`, `or`, `not`
- English aliases: `when`, `otherwise`, `loop`, `each`, `give`
- Literals: `INT`, `FLOAT`, `STRING`, `BOOL`, `NONE`
- Operators: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `=`, `+=`, `-=`, `*=`, `/=`, `%=`
- Delimiters: `(`, `)`, `{`, `}`, `[`, `]`, `,`, `:`, `.`
- Special: `NEWLINE`, `INDENT`, `DEDENT`, `EOF`

### Parser

Recursive descent parser that produces an AST. Located in `internal/parser/parser.go` (1169 lines) and `internal/parser/ast.go` (464 lines).

**AST node types:**
- Statements: `Let`, `Assign`, `AugAssign`, `FnDef`, `StructDef`, `If`, `While`, `For`, `Return`, `Break`, `Continue`, `Use`, `TestBlock`, `ExprStmt`
- Expressions: `Ident`, `IntLit`, `FloatLit`, `StringLit`, `BoolLit`, `NoneLit`, `Binary`, `Unary`, `Call`, `Member`, `ListLit`, `MapLit`
- Web nodes: `Page`, `Component`, `State`, `HtmlElement`, `EventHandler`, `StyleBlock`, `TextInterp`, `AwaitExpr`
- Other: `Program`, `Block`, `Param`

### Evaluator

Tree-walking interpreter that evaluates AST nodes directly. Located in `internal/eval/eval.go` (699 lines), `internal/eval/expressions.go` (600 lines), and `internal/eval/builtins.go` (753 lines).

**Key components:**
- `Environment` -- Nested scope chain with `map[string]Value`
- `Value` -- Runtime value representation (integer, float, string, boolean, none, list, map, struct, function, range, error)
- Built-in functions and modules registered at initialization

### Web Compiler

Three-stage pipeline that compiles Say Less AST to web output:

1. **Compiler** (`internal/web/compiler.go`, 455 lines) -- Transforms AST into Web IR
2. **IR** (`internal/web/ir.go`, 159 lines) -- Intermediate representation node types
3. **Generator** (`internal/web/generator.go`, 383 lines) -- Produces HTML, CSS, JS from IR

**IR node types:**
- `IRText` -- Static text content
- `IRInterp` -- Interpolated text
- `IRElement` -- HTML element with attributes, children, events
- `IRComponent` -- Component instance with props
- `IRState` -- Reactive state variable
- `IRConditional` -- If/else rendering
- `IRLoop` -- For-loop rendering
- `IRAwait` -- Async data loading
- `IRFragment` -- Group of child nodes
- `StyleBlock` -- Inline CSS properties

### Browser Runtime

The `SL` JavaScript runtime (`internal/web/runtime.go`, 193 lines) is injected into generated HTML. It provides:

- **State management**: `SL.setState(name, value)`, `SL.getState(name)`
- **Reactive bindings**: Automatic DOM updates when state changes
- **Event delegation**: Handles click, input, submit events
- **DOM helpers**: `$(selector)`, `$$(selector)`, `createElement(tag, attrs, children)`
- **Template rendering**: `${expression}` syntax in templates
- **Fetch helper**: `SL.fetchData(url, options)` with JSON auto-parsing
- **Storage helpers**: `SL.saveLocal(key, value)`, `SL.loadLocal(key)`

---

## Compilation Pipeline

### Interpreter Pipeline (Current)

```
Source (.sl)
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Evaluator (tree-walk) → Result
```

### Web Compiler Pipeline

```
Source (.sl)
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Web Compiler → Web IR
    ↓
Generator → HTML + CSS + JS
```

### Planned Compiler Pipeline (Phase 2)

```
Source (.sl)
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Semantic Analysis → Typed AST
    ↓
IR Generator → SSA-based IR
    ↓
Optimizer → Optimized IR
    ↓
Code Generator → Bytecode / Native
    ↓
VM Execution / Native Execution
```

---

## Testing

### Test Blocks

```
test "addition works"
    assert add(2, 3) == 5

test "string operations"
    assert "hello".length == 5
    assert "HELLO".lower() == "hello"

test "list mutation"
    items = [1, 2, 3]
    items.push(4)
    assert items.length == 4

test "division by zero"
    result = divide(10, 0)
    assert result.err
    assert result.error == "division by zero"
```

### Running Tests

```bash
sale test
```

Tests are collected from all `.sl` files in the project. Files with "test" in the path are prioritized.

### Error Handling in Tests

```
fn divide(a, b)
    if b == 0
        return err("division by zero")
    return ok(a / b)

test "error handling"
    result = divide 10 0
    assert result.err
    print result.error  # "division by zero"
```

---

## Deployment

### Docker

Each web project includes a `Dockerfile` and `compose.yml`.

**Dockerfile:**

```dockerfile
FROM alpine:latest
COPY sale-bin /app/sale-bin
EXPOSE 8080
CMD ["/app/sale-bin"]
```

**compose.yml:**

```yaml
version: "3"
services:
  app:
    build: .
    ports:
      - "8080:8080"
```

### Running with Docker

```bash
docker compose up --build
```

### Release Process

`release.bat` builds for all platforms and packages into archives:

- `say_less.exe` (Windows GUI installer — recommended for Windows users)
- `sale-windows-amd64.zip`
- `sale-macos-amd64.tar.gz`
- `sale-macos-arm64.tar.gz`
- `sale-linux-amd64.tar.gz`
- `sale-linux-arm64.tar.gz`

### Build Output

```
build/
    web/
        index.html     # Generated HTML
```

The web compiler outputs to `build/web/`. Each `page` declaration generates a separate HTML file.

---

## Examples

### Minimal Website

```
page "/"
    h1 "Hello, World!"
    p "My first Say Less website."
```

### Counter App

```
page "/"

    state count = 0

    main
        h1 "Counter: " + count

        button "Increment"
            on click
                count += 1

        button "Decrement"
            on click
                count -= 1

        button "Reset"
            on click
                count = 0
```

### Todo List

```
page "/"

    state todos = []
    state newTodo = ""

    main
        h1 "Todo List"

        input type "text" placeholder "Add a todo..."
            on input
                newTodo = input.value

        button "Add"
            on click
                if newTodo != ""
                    push(todos, newTodo)
                    newTodo = ""

        ul
            for todo in todos
                li todo
```

### Component Library

```
component Button(text, color)
    button style ("background-color: " + color + "; color: white; padding: 10px 20px; border: none; cursor: pointer;")
        text

component Card(title, description, color)
    div style ("border: 2px solid " + color + "; border-radius: 8px; padding: 20px; margin: 10px;")
        h2 title
        p description

page "/"
    main
        h1 "Component Library"

        Button(text: "Primary", color: "#007bff")
        Button(text: "Danger", color: "#dc3545")

        Card(
            title: "Card 1",
            description: "This is a card component.",
            color: "#007bff"
        )
```

### HTTP API

```
use http
use json

server on 8080

get "/"
    return "API is running"

get "/api/users"
    users = [
        {id: 1, name: "Alice"},
        {id: 2, name: "Bob"}
    ]
    return json users

post "/api/users"
    newUser = request.body
    return json {status: "created", user: newUser}
```

### Full Web Application with npm

```
use http
use npm:gsap

server on 8080

get "/"
    return """
        <!DOCTYPE html>
        <html>
        <head>
            <title>Say Less App</title>
            <script src="/node_modules/gsap/dist/gsap.min.js"></script>
        </head>
        <body>
            <h1 id="title">Welcome</h1>
            <script>
                gsap.fromTo("#title",
                    {y: 60, opacity: 0},
                    {y: 0, opacity: 1, duration: 1}
                );
            </script>
        </body>
        </html>
        """
```

---

## Roadmap

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 0 | Tree-walking interpreter | Current |
| Phase 1 | Bytecode compiler + VM | Planned |
| Phase 2 | Self-hosting compiler core | Planned |
| Phase 3 | Full self-hosting | Planned |

### Planned Features

- `match` expression for pattern matching
- `await`/`task` for concurrency (green threads)
- Error propagation operator `?`
- Methods on structs
- WebSocket support
- CSS/Tailwind integration
- Standard library expansion (crypto, regex, bytes, terminal)
- Package registry and publishing
- Linter and formatter improvements
- Full self-hosting (compiler rewritten in Say Less)

---

## License

Say Less is an open-source project. See the repository for license details.
