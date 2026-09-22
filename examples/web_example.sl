# Say Less Web - Example Application
# This demonstrates the core Say Less Web features

page "/"

    state count = 0

    main

        h1 "My First Say Less Website"

        p "I did not write HTML, CSS, or JavaScript."

        button "Clicked: " + count
            on click
                count += 1

        p "The counter updates without page reload."
