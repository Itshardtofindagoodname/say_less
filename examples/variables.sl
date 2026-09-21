name = "Say Less"
age = 10
active = true
pi = 3.14

print "Name: " + name
print "Age: " + to_string(age)
print "Active: " + to_string(active)
print "Pi: " + to_string(pi)

mut counter = 0
counter += 1
counter += 1
counter += 1
print "Counter: " + to_string(counter)

if age >= 10
    print "old enough"
else
    print "too young"

for i in range(5)
    print "i = " + to_string(i)

fruits = ["apple", "banana", "cherry"]
for fruit in fruits
    print "Fruit: " + fruit

greeting = fn(name)
    return "Hello, " + name + "!"

print greeting("World")

fn add(a, b)
    return a + b

result = add(10, 20)
print "10 + 20 = " + to_string(result)
