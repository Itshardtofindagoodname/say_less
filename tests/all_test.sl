# Say Less Tests

test "integer arithmetic"
    assert 1 + 1 == 2
    assert 10 - 3 == 7
    assert 4 * 5 == 20
    assert 10 / 2 == 5
    assert 10 % 3 == 1

test "float arithmetic"
    assert 1.5 + 2.5 == 4.0
    assert 3.0 * 2.0 == 6.0

test "string operations"
    assert "hello" + " " + "world" == "hello world"
    assert "hello".length == 5
    assert "HELLO".lower() == "hello"
    assert "hello".upper() == "HELLO"
    assert "hello world".contains("world")
    assert not "hello".contains("xyz")

test "boolean logic"
    assert true and true
    assert not false
    assert true or false
    assert not (true and false)

test "comparison operators"
    assert 1 < 2
    assert 2 > 1
    assert 1 <= 1
    assert 1 >= 1
    assert 1 != 2
    assert 1 == 1

test "list operations"
    items = [1, 2, 3]
    assert items.length == 3
    items.push(4)
    assert items.length == 4
    last = items.pop()
    assert last == 4
    assert items.length == 3

test "map operations"
    m = {name: "Jeet", age: 30}
    assert m.name == "Jeet"
    assert m.age == 30
    assert has(m, "name")
    assert not has(m, "email")

test "json encode/decode"
    data = json.decode("{\"x\": 42}")
    assert data.x == 42
    encoded = json.encode({a: 1})
    assert encoded == "{\"a\":1}"

test "math functions"
    assert math.sqrt(9) == 3.0
    assert math.floor(3.7) == 3.0
    assert math.ceil(3.2) == 4.0

test "string methods"
    parts = "a,b,c".split(",")
    assert parts.length == 3
    joined = parts.join("-")
    assert joined == "a-b-c"
    assert "hello".starts_with("he")
    assert "hello".ends_with("lo")
    assert "hello".index_of("ll") == 2

test "range"
    mut sum = 0
    for i in 1..4
        sum += i
    assert sum == 6

test "closures"
    fn make_counter()
        mut count = 0
        return fn()
            count += 1
            return count
    c = make_counter()
    assert c() == 1
    assert c() == 2
    assert c() == 3

test "fibonacci recursive"
    fn fib(n)
        if n <= 1
            return n
        return fib(n - 1) + fib(n - 2)
    assert fib(0) == 0
    assert fib(1) == 1
    assert fib(10) == 55

test "higher order functions"
    fn apply(f, x)
        return f(x)
    fn double(x)
        return x * 2
    assert apply(double, 5) == 10

test "struct"
    struct Point
        x = 0
        y = 0
    p = Point(x: 10, y: 20)
    assert p.x == 10
    assert p.y == 20

test "nested loops"
    mut total = 0
    for i in 1..4
        for j in 1..4
            total += 1
    assert total == 9

test "error handling"
    result = err("something went wrong")
    assert result.err
    assert result.message == "something went wrong"

test "none value"
    x = none
    assert not x

print ""
print "All tests passed!"
