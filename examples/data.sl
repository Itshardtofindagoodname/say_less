# Data structures
point = {x: 10, y: 20}
print "Point: " + json.encode(point)

nums = [1, 2, 3, 4, 5]
print "List: " + json.encode(nums)

# String operations
name = "hello world"
print "Upper: " + name.upper()
print "Lower: " + name.lower()
print "Length: " + to_string(name.length)
print "Contains 'world': " + to_string(name.contains("world"))
print "Split: " + json.encode(name.split(","))
print "Join: " + strings.join(["x", "y", "z"], "-")

# Math
print "sqrt(144) = " + to_string(math.sqrt(144))
print "pi = " + to_string(math.pi)
print "floor(3.7) = " + to_string(math.floor(3.7))
print "ceil(3.2) = " + to_string(math.ceil(3.2))

# JSON
data = json.decode("{\"name\": \"Jeet\", \"age\": 30}")
print "Name: " + data.name
print "Age: " + to_string(data.age)

# Map operations
m = {a: 1, b: 2, c: 3}
print "Keys: " + json.encode(keys(m))
print "Has 'a': " + to_string(has(m, "a"))
print "Has 'z': " + to_string(has(m, "z"))

# List operations
items = [10, 20, 30]
print "Length: " + to_string(items.length)
items.push(40)
print "After push: " + json.encode(items)
last = items.pop()
print "Popped: " + to_string(last)
print "After pop: " + json.encode(items)

# Assertions
assert 1 + 1 == 2
assert "hello" + " " + "world" == "hello world"
assert true and true
assert not false
print "All assertions passed!"
