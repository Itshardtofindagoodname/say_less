# Say Less

A tiny, fast, general-purpose programming language.

> Say less. Make the computer understand more.

## Installation

### Option 1: Download a pre-built binary (recommended)

No Go compiler is required — every release ships pre-built binaries on the [Releases](https://github.com/Itshardtofindagoodname/say_less/releases) page.

Download the latest release for your platform from the [Releases](https://github.com/Itshardtofindagoodname/say_less/releases) page.

| Platform | File |
|----------|------|
| Windows (x64) | `sale-windows-amd64.zip` |
| macOS (Intel) | `sale-macos-amd64.tar.gz` |
| macOS (Apple Silicon) | `sale-macos-arm64.tar.gz` |
| Linux (x64) | `sale-linux-amd64.tar.gz` |
| Linux (ARM64) | `sale-linux-arm64.tar.gz` |

Each release archive includes the platform's installation script.

#### Windows — install.bat (recommended)

1. Download and extract `sale-windows-amd64.zip`.
2. Double-click `install.bat`.
3. It installs `sale.exe` to `%LOCALAPPDATA%\Programs\SayLess\bin` and adds that folder to your **user** PATH. Administrator permission is not required.
4. Open a new terminal and verify:

```powershell
sale --version
# sale 0.1.0
```

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

#### Verify installation

```bash
sale --version
# sale 0.1.0

sale --help
```

#### Uninstall (Windows)

To remove Say Less — its files, the `sale` command, and its PATH entries — run:

```powershell
sale -g uninstall
```

The command:

- Deletes the install directory (`%LOCALAPPDATA%\Programs\SayLess`)
- Removes the Say Less `bin` folder from the user PATH
- Deletes the running copy once the command finishes

Then close and reopen the terminal and confirm the command is gone:

```powershell
Get-Command sale -ErrorAction SilentlyContinue   # should return nothing
```

### Option 2: Build from source

Requires [Go](https://go.dev/dl/) 1.21 or later.

```bash
git clone https://github.com/yourname/say_less.git
cd say_less
go build -o sale ./cmd/sale/        # macOS / Linux
go build -o sale.exe ./cmd/sale/    # Windows
```

Move the binary to a directory in your PATH:

```bash
sudo mv sale /usr/local/bin/        # macOS / Linux
```

Or on Windows, move `sale.exe` to a folder already in your PATH (e.g. `C:\Go\bin\`), or:

```powershell
$env:PATH += ";C:\path\to\your\binary"
```

## Quick Start

```bash
# Create a new project
sale new hello
cd hello

# Run it
sale run src/main.sl
```

## Hello World

Create a file `hello.sl`:

```
print "Hello, World!"
```

Run it:

```bash
sale run hello.sl
```

## CLI Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `sale run <file>` | Run a Say Less source file |
| `sale build [--release]` | Build the project. Use `--release` for optimized builds |
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
| `sale -g uninstall` | Uninstall Say Less from your system (removes files and PATH entries) |

## Language Overview

### Variables

```
name = "Jeet"
age = 30
active = true
pi = 3.14
mut counter = 0
counter += 1
```

Variables are immutable by default. Use `mut` to declare a mutable variable.

### Functions

```
fn greet(name)
    return "Hello, " + name + "!"

print greet("World")
```

Parentheses are optional:

```
print "hello"
print 2 + 3
```

### Control Flow

Blocks support three styles - pick the one you like:

```
# Style 1: Curly braces (C/Java)
if age >= 18 {
    print "adult"
} else {
    print "child"
}

# Style 2: Colon + indent (Python)
if age >= 18:
    print "adult"
else:
    print "child"

# Style 3: Bare indent (no colon needed)
if age >= 18
    print "adult"
else
    print "child"
```

All three styles work for `if`, `while`, `for`, `fn`, `struct`, and `test`.

English-like aliases are also available:

```
# aliases: when=if, otherwise=else, loop=while, each=for, give=return
when age >= 18 {
    give "adult"
} otherwise {
    give "child"
}

each item in items {
    print item
}
```

### Data Structures

```
# Lists
items = [1, 2, 3]
items.push(4)
print items.length

# Maps
person = {name: "Jeet", age: 30}
print person.name

# Structs
struct Point
    x = 0
    y = 0

p = Point(x: 10, y: 20)
print p.x
```

### String Methods

```
name = "hello world"
print name.upper()              # HELLO WORLD
print name.length               # 11
print name.contains("world")    # true
print name.split(" ")           # ["hello", "world"]
print name.replace("hello", "hey")
print name.starts_with("hello") # true
print name.ends_with("world")   # true
print name.index_of("world")    # 6
print name.trim()
print name.to_int()
```

### List Methods

```
items = [1, 2, 3]
items.push(4)           # [1, 2, 3, 4]
items.pop()             # returns 4, list is [1, 2, 3]
items.length            # 3
items.join(", ")        # "1, 2, 3"
items.map(fn(x) x * 2)  # [2, 4, 6]
items.filter(fn(x) x > 1) # [2, 3]
items.find(2)           # 2
```

### Modules

```
# Math
print math.sqrt(144)    # 12
print math.pi           # 3.141592653589793
print math.floor(3.7)   # 3
print math.ceil(3.2)    # 4

# JSON
data = json.decode("{\"x\": 42}")
print data.x
encoded = json.encode({a: 1, b: 2})
print encoded

# Strings
print strings.contains("hello world", "world")
parts = strings.split("a,b,c", ",")
joined = strings.join(["x", "y"], "-")

# I/O
content = io.read("file.txt")
io.write("output.txt", "hello")
exists = io.exists("file.txt")

# OS
args = os.args
val = os.env("HOME")
t = os.clock()
```

### Built-in Functions

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
| `push(list, item)` | Add item to list (returns new list) |
| `pop(list)` | Remove and return last item (returns new list) |
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

### Testing

```
test "addition works"
    assert 1 + 1 == 2

test "string operations"
    assert "hello".length == 5
    assert "HELLO".lower() == "hello"

test "list mutation"
    items = [1, 2, 3]
    items.push(4)
    assert items.length == 4
```

Run with `sale test`.

## Project Structure

```
myproject/
  sale.toml          # Project configuration
  src/
    main.sl          # Entry point
  tests/
    all_test.sl      # Test files
```

### sale.toml

```toml
name = "myproject"
version = "0.1.0"

[dependencies]
http = "0.1"
json = "0.1"
```

## Architecture

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
Evaluator -> Result
```

**Bootstrap language:** Go
**Runtime:** Modular, minimal tree-walking interpreter

## Built-in Modules

| Module | Description |
|--------|-------------|
| `math` | sqrt, pow, floor, ceil, sin, cos, tan, log, pi, e |
| `json` | encode, decode |
| `strings` | contains, split, join, replace, has_prefix, has_suffix |
| `io` | read, write, exists |
| `os` | args, env, clock |
