# Say Less - Architecture

## Overview

Say Less is a general-purpose programming language designed around one principle:
**Say less. Make the computer understand more.**

The system consists of:

1. **Lexer** - Tokenizes source code
2. **Parser** - Produces AST from tokens
3. **Evaluator** - Tree-walking interpreter (Phase 1)
4. **Compiler** - Bytecode compiler + VM (Phase 2, separate project)
5. **Standard Library** - Core modules implemented in Go, callable from Say Less
6. **Package Manager** - Dependency resolution, registry, caching
7. **CLI** - The `sale` command
8. **Tooling** - Formatter, linter, REPL, test runner

## Compilation Pipeline (Phase 1)

```
Source (.sl)
    ↓
Lexer → Tokens
    ↓
Parser → AST
    ↓
Evaluator (tree-walk) → Result
```

## Compilation Pipeline (Phase 2 - future)

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

## Lexical Rules

- **Indentation**: Uses 4-space indentation for blocks. Tabs are not supported. Curly braces `{ }` can be used as an alternative to indentation.
- **Comments**: `# this is a comment` (line comments only)
- **Strings**: Double-quoted `"hello"`, single-quoted `'hello'`, multi-line `"""`
- **Numbers**: Integers `42`, floats `3.14`, hex `0xFF`, binary `0b1010`
- **Booleans**: `true`, `false`
- **None**: `none`
- **Identifiers**: `lowercase_with_underscores` or `camelCase`. Must start with letter or underscore.
- **Newlines**: Significant. Each newline terminates a statement unless the line is incomplete (ends with operator, open paren, etc.).
- **English aliases**: `when`=if, `otherwise`=else, `loop`=while, `each`=for, `give`=return

## Grammar (EBNF)

```
program         = statement* EOF

statement       = assignment
                | augmented_assignment
                | function_def
                | struct_def
                | if_stmt
                | while_stmt
                | for_stmt
                | return_stmt
                | break_stmt
                | continue_stmt
                | use_stmt
                | test_block
                | expression_stmt

assignment      = IDENT (':' type)? '=' expression
augmented_assignment = IDENT ('+=' | '-=' | '*=' | '/=' | '%=') expression

function_def    = 'fn' IDENT '(' param_list? ')' block
                | 'fn' IDENT param_list block

struct_def      = 'struct' IDENT block

if_stmt         = 'if' expression block ('else_if' expression block)* ('else' block)?
while_stmt      = 'while' expression block
for_stmt        = 'for' IDENT 'in' expression block
return_stmt     = 'return' expression?
break_stmt      = 'break'
continue_stmt   = 'continue'

use_stmt        = 'use' IDENT ('as' IDENT)?

test_block      = 'test' STRING block

block           = ':' NEWLINE INDENT statement* DEDENT
                | '{' statement* '}'
                | NEWLINE INDENT statement* DEDENT

expression      = or_expr
or_expr         = and_expr ('or' and_expr)*
and_expr        = not_expr ('and' not_expr)*
not_expr        = 'not' not_expr | comparison
comparison      = addition (('==' | '!=' | '<' | '>' | '<=' | '>=') addition)*
addition        = multiplication (('+' | '-') multiplication)*
multiplication  = unary (('*' | '/' | '%') unary)*
unary           = ('-' | '+') unary | postfix
postfix         = primary ('.' IDENT | '(' arg_list? ')')*

primary         = INT | FLOAT | STRING | BOOL | NONE
                | IDENT
                | '(' expression ')'
                | list_literal
                | map_literal

list_literal    = '[' (expression (',' expression)*)? ']'
map_literal     = '{' (IDENT ':' expression (',' IDENT ':' expression)*)? '}'

param_list      = param (',' param)*
param           = IDENT (':' type)? ('=' expression)?

type            = IDENT | IDENT '<' type '>' | type '|'
                | list_type | map_type
list_type       = 'list' '<' type '>'
map_type        = 'map' '<' type ',' type '>'

arg_list        = expression (',' expression)*
```

## Type System

### Primitive Types
- `integer` - 64-bit signed integer
- `float` - 64-bit IEEE 754 float
- `boolean` - `true` or `false`
- `string` - Immutable UTF-8 string
- `bytes` - Raw byte sequence
- `none` - Absence of value

### Compound Types
- `list<T>` - Ordered, dynamic array
- `map<K, V>` - Hash map
- `struct` - Named record type
- `function` - First-class function

### Result Types
- `ok<T>` - Success value
- `err<E>` - Error value

### Type Inference
Types are inferred from context. Explicit types are optional:

```
age = 20              # inferred as integer
name: string = "Jeet" # explicit
```

## Memory Model (Phase 1)

Phase 1 uses Go's garbage collector via the Go runtime.

### Future Memory Model
- Reference counting for deterministic cleanup
- Escape analysis for stack allocation
- Optional manual memory management for systems code
- Arena allocation for batch processing

## Error Handling

Errors are values, not exceptions.

```
result = read_file "config.json"
if result.err
    print result.err
    exit 1
```

Propagation operator `?` (planned):
```
content = read_file "config.json"?  # propagates error to caller
```

## Module System

```
use fs
use http
use json
use ./local_module
use package_name
```

Modules map to files:
- `use fs` → `stdlib/fs.sl` or built-in
- `use ./utils` → `./utils.sl`

## Concurrency Model (Phase 2)

```
task fetch_data url
    return http.get url

result = await fetch_data "https://api.example.com"
```

- Green threads / goroutines under the hood
- Channels for communication
- Structured concurrency

## Standard Library Modules

Core modules (built-in, implemented in Go):
- `io` - stdin/stdout/stderr
- `fs` - file system operations
- `path` - path manipulation
- `http` - HTTP client/server
- `json` - JSON encode/decode
- `time` - time operations
- `math` - math functions
- `strings` - string operations
- `process` - process management
- `env` - environment variables
- `args` - command-line arguments
- `terminal` - terminal colors and formatting
- `crypto` - hashing, encryption
- `bytes` - byte manipulation
- `regex` - regular expressions
- `testing` - test utilities

## Package Manager

### Project Configuration
`sale.toml`:
```toml
name = "myapp"
version = "0.1.0"

[dependencies]
http = "1.0"
json = "1.0"

[dev-dependencies]
testing = "1.0"
```

### Commands
- `sale init` - Initialize project
- `sale add <pkg>` - Add dependency
- `sale remove <pkg>` - Remove dependency
- `sale install` - Install dependencies
- `sale publish` - Publish package

## CLI Architecture

The `sale` CLI is a single binary with subcommands:

```
sale new <name>          # Create new project
sale create --web <name> # Create web project
sale create --backend <name> # Create backend project
sale create --system <name>  # Create system project
sale run <file>          # Run a Say Less file
sale build [--release]   # Build project
sale dev                 # Start development server
sale test                # Run tests
sale fmt                 # Format code
sale check               # Type/check code
sale repl                # Start REPL
sale add <pkg>           # Add dependency
sale remove <pkg>        # Remove dependency
sale install             # Install dependencies
sale update              # Update dependencies
sale publish             # Publish package
sale search <query>      # Search packages
sale info <pkg>          # Package info
sale clean               # Clean build artifacts
sale doctor              # Diagnose issues
```

## Web Architecture

### Backend
- Say Less HTTP server compiles to native binary
- Built-in HTTP router
- JSON serialization/deserialization
- WebSocket support (planned)

### Frontend
- Say Less compiles to HTML/CSS/JS/WASM
- Component-based HTML generation
- CSS support (including Tailwind integration)
- JavaScript interop for browser APIs

## Self-Hosting Strategy

1. **Phase 0**: Interpreter in Go (current)
2. **Phase 1**: Add bytecode compiler in Go
3. **Phase 2**: Rewrite compiler core in Say Less
4. **Phase 3**: Full self-hosting

The compiler is designed so each component can be rewritten independently.

## Bootstrap Language: Go

Go was chosen for the bootstrap compiler because:
- Produces standalone binaries with no runtime dependencies
- Fast compilation
- Excellent standard library
- Cross-platform (Linux, macOS, Windows)
- Simple, readable code
- Easy to produce correct code
- Garbage-collected (simplifies initial implementation)
