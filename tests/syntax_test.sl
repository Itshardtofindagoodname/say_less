# Say Less - All Syntax Styles
# This language supports 3 ways to write blocks:

print "=== Curly Braces (C/Java style) ==="

if true {
    print "  if with braces works!"
}

for i in 1..3 {
    print "  for loop: " + to_string(i)
}

fn multiply(a, b) {
    return a * b
}
assert multiply(3, 4) == 12

print ""
print "=== Colons + Indent (Python style) ==="

if true:
    print "  if with colon works!"

for i in 1..3:
    print "  for loop: " + to_string(i)

fn divide(a, b) {
    return a / b
}
assert divide(10, 2) == 5

print ""
print "=== Just Indent (no colon needed) ==="

x = 42
if x == 42
    print "  bare indent works!"

print ""
print "=== English Aliases ==="

each i in 1..3 {
    print "  each (=for): " + to_string(i)
}

fn greet(name) {
    give "Hello, " + name + "!"
}
assert greet("World") == "Hello, World!"

print ""
print "=== Nested Braces ==="

for i in 1..3 {
    for j in 1..2 {
        print "  " + to_string(i) + "x" + to_string(j) + "=" + to_string(i * j)
    }
}

print ""
print "=== Mixed Styles ==="

if true {
    result = ""
    for word in ["hello", "world"]
        result += word + " "
    assert result == "hello world "
}

print ""
print "=== Route Handlers (for web) ==="

# server on 8080
# get "/" {
#     return html """..."""
# }

print "All syntax tests passed!"
