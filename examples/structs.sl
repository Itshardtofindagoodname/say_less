struct Point
    x = 0
    y = 0

struct Person
    name = ""
    age = 0

p = Point(x: 10, y: 20)
print "Point: (" + to_string(p.x) + ", " + to_string(p.y) + ")"

person = Person(name: "Jeet", age: 30)
print "Person: " + person.name + " (age " + to_string(person.age) + ")"

# Struct in a list
p1 = Point(x: 1, y: 2)
p2 = Point(x: 3, y: 4)
p3 = Point(x: 5, y: 6)
points = [p1, p2, p3]

for pt in points
    print "(" + to_string(pt.x) + ", " + to_string(pt.y) + ")"
