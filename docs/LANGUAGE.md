# Say Less Language Specification

## 1. Philosophy

> Say less. Make the computer understand more.

Every feature must justify its existence. Syntax is minimized. Meaning is maximized.
A child should understand simple programs. An expert should build serious systems.

## 2. Lexical Structure

### 2.1 Encoding
Source files are UTF-8 encoded.

### 2.2 Keywords

```
fn if else else_if while for in
return break continue
true false none
use as
struct
test assert
let mut const
and or not
match when
self

# English-like aliases
when        # alias for if
otherwise   # alias for else
loop        # alias for while
each        # alias for for
give        # alias for return
```

### 2.3 Identifiers

```
name        # valid
_count      # valid
camelCase   # valid
myVar123    # valid

123name     # INVALID - starts with digit
my-var      # INVALID - contains hyphen
```

### 2.4 Literals

#### Integers
```
42
0
1_000_000
0xFF
0b1010
0o77
```

#### Floats
```
3.14
0.5
1.0e10
```

#### Strings
```
"hello world"
'hello world'
"line one\nline two"
"she said \"hello\""
```

#### Multi-line Strings
```
"""
This is a
multi-line string.
"""
```

#### Booleans
```
true
false
```

#### None
```
none
```

### 2.5 Operators

```
+ - * / %          # arithmetic
== != < > <= >=    # comparison
= += -= *= /= %=  # assignment
.                   # member access
->                  # arrow (future: lambdas)
```

### 2.6 Comments

```
# This is a line comment
```

No block comments. Keep comments simple.

### 2.7 Significant Newlines

Newlines separate statements. Statements are not terminated by semicolons.

A line is continued if it ends with:
- An operator
- An open parenthesis `(`, bracket `[`, or brace `{`
- A comma `,`

## 3. Types

### 3.1 Primitive Types

| Type | Description | Example |
|------|-------------|---------|
| `integer` | 64-bit signed integer | `42` |
| `float` | 64-bit IEEE 754 float | `3.14` |
| `boolean` | true or false | `true` |
| `string` | Immutable UTF-8 string | `"hello"` |
| `bytes` | Raw byte sequence | `bytes(10)` |
| `none` | Absence of value | `none` |

### 3.2 Compound Types

| Type | Description | Example |
|------|-------------|---------|
| `list<T>` | Dynamic array | `[1, 2, 3]` |
| `map<K, V>` | Hash map | `{"a": 1}` |
| `struct` | Named record | `struct Point` |
| `function` | First-class function | `fn(a, b) a + b` |

### 3.3 Type Inference

```
age = 20              # inferred: integer
name = "Jeet"         # inferred: string
active = true         # inferred: boolean
items = [1, 2, 3]    # inferred: list<integer>
```

Explicit types are optional:
```
port: integer = 8080
name: string = "Jeet"
```

### 3.4 Type Annotations

Function parameters and return types:
```
fn add(a integer, b integer) integer
    return a + b
```

Short form (types inferred):
```
fn add a b
    return a + b
```

## 4. Variables

### 4.1 Bindings

```
name = "Jeet"         # immutable binding
mut counter = 0        # mutable binding
const MAX_SIZE = 1024  # compile-time constant
```

**Default**: Immutable. Use `mut` when mutation is needed.

### 4.2 Augmented Assignment

```
counter += 1
counter -= 1
counter *= 2
counter /= 3
counter %= 5
```

## 5. Functions

### 5.1 Definition

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

### 5.2 Calling

```
greet "Jeet"
result = add 2 3
```

Parentheses are optional for function calls:
```
greet("Jeet")      # also valid
result = add(2, 3) # also valid
```

### 5.3 Default Arguments

```
fn greet(name, greeting = "Hello")
    print greeting + ", " + name

greet "Jeet"                    # Hello, Jeet
greet "Jeet", greeting = "Hi"   # Hi, Jeet
```

### 5.4 First-Class Functions

```
fn apply(f, value)
    return f(value)

double = fn(x) x * 2
result = apply double 5    # 10
```

### 5.5 Closures

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

## 6. Control Flow

Blocks support three styles. All three are equivalent - use whichever you prefer:

- **Curly braces** - `if cond { ... }` (C/Java style)
- **Colon + indent** - `if cond: ...` (Python style)
- **Bare indent** - `if cond\n    ...` (no colon needed)

The colon after `if`, `while`, `for`, `fn`, `struct`, and `test` is always optional.

### 6.1 If / Else

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

### 6.2 While

```
mut i = 0
while i < 10 {
    print i
    i += 1
}

# Bare indent (no colon)
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

### 6.3 For

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

### 6.4 Break / Continue

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

### 6.5 Match (planned)

```
match value
    0 -> print "zero"
    1 -> print "one"
    _ -> print "other"
```

## 7. Error Handling

### 7.1 Result Values

Functions return `ok(value)` or `err(message)`:

```
fn divide(a, b)
    if b == 0
        return err("division by zero")
    return ok(a / b)

result = divide 10 2
if result.ok
    print result.value

result = divide 10 0
if result.err
    print result.error
```

### 7.2 Error Propagation (planned)

```
fn load_config()
    content = read_file("config.json")?  # propagates error
    return parse_json content
```

## 8. Modules

### 8.1 Importing

```
use fs
use http
use json
use ./utils
use ./models/user
```

### 8.2 Using

```
use fs

content = fs.read "file.txt"
fs.write "output.txt" "hello"
```

### 8.3 Module Files

Each file is a module. Public functions are accessible:

```python
# utils.sl
fn helper()
    return 42
```

```python
# main.sl
use utils
print utils.helper()
```

## 9. Structs

### 9.1 Definition

```
struct Point
    x integer
    y integer
```

### 9.2 Instantiation

```
p = Point(x: 10, y: 20)
print p.x    # 10
print p.y    # 20
```

### 9.3 Methods (planned)

```
struct Point
    x integer
    y integer

    fn distance(self, other)
        dx = self.x - other.x
        dy = self.y - other.y
        return math.sqrt(dx * dx + dy * dy)
```

## 10. Testing

### 10.1 Test Blocks

```
test "addition works"
    assert add(2, 3) == 5

test "division by zero"
    result = divide(10, 0)
    assert result.err
    assert result.error == "division by zero"
```

### 10.2 Running Tests

```
sale test
```

Tests are collected from all `.sl` files in the project.

## 11. Web Development

### 11.1 HTTP Server (planned)

```
use http

server on 8080

get "/"
    return "Hello from Say Less"

get "/users"
    users = ["Alice", "Bob"]
    return json users

post "/data"
    body = request.body
    return json { status: "ok" }
```

### 11.2 HTML Generation (planned)

```
page "/"
    view
        heading "Say Less"
        text "Build more. Say less."
        div class "hero"
            heading "Welcome"
```

### 11.3 CSS (planned)

```
css
    body
        margin 0
        font-family sans-serif

    .hero
        display flex
        justify-content center
```

## 12. Concurrency (Phase 2)

```
task fetch url
    return http.get url

result = await fetch "https://api.example.com"
```

Channels:
```
ch = channel integer

task producer
    ch.send 1
    ch.send 2
    ch.close()

task consumer
    for value in ch
        print value
```

## 13. Operator Precedence (highest to lowest)

1. `.` member access
2. Function call `f(x)`
3. Unary `-`, `+`, `not`
4. `*`, `/`, `%`
5. `+`, `-`
6. `==`, `!=`, `<`, `>`, `<=`, `>=`
7. `and`
8. `or`
9. `=` assignment
