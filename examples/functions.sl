fn factorial(n)
    if n <= 1
        return 1
    return n * factorial(n - 1)

print "5! = " + to_string(factorial(5))
print "10! = " + to_string(factorial(10))

fn fibonacci(n)
    if n <= 1
        return n
    return fibonacci(n - 1) + fibonacci(n - 2)

for i in range(10)
    print "fib(" + to_string(i) + ") = " + to_string(fibonacci(i))

fn fibonacci_loop(n)
    mut a = 0
    mut b = 1
    mut i = 0
    while i < n
        mut temp = b
        b = a + b
        a = temp
        i += 1
    return a

print "fib_loop(10) = " + to_string(fibonacci_loop(10))
