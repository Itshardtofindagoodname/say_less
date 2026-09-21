# FizzBuzz
for i in 1..16
    if i % 15 == 0
        print "FizzBuzz"
    else_if i % 3 == 0
        print "Fizz"
    else_if i % 5 == 0
        print "Buzz"
    else
        print to_string(i)

print ""

# Counting
mut count = 0
while count < 5
    print "count = " + to_string(count)
    count += 1
    
# Nested loops
print ""
print "Multiplication table:"
for i in 1..4
    mut row = ""
    for j in 1..4
        row = row + to_string(i * j) + "\t"
    print row
