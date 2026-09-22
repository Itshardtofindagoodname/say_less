# Say Less Web - Full Example
# Demonstrates components, pages, state, and events

component Card(title, description)

    article class "card"

        h2 title
        p description


page "/"

    state count = 0
    state showDetails = false

    main

        h1 "Say Less Web Demo"

        p "Welcome to the future of web development."

        Card(
            title: "Components",
            description: "Build reusable UI components with simple syntax."
        )

        Card(
            title: "State Management",
            description: "No complex state libraries needed."
        )

        Card(
            title: "No JavaScript",
            description: "Write Say Less, get a working website."
        )

        section

            h2 "Interactive Counter"

            button "Count: " + count
                on click
                    count += 1

            button "Reset"
                on click
                    count = 0

            if count > 10

                p "Wow, you clicked a lot!"

        section

            h2 "Conditional Content"

            button "Toggle Details"
                on click
                    showDetails = not showDetails

            if showDetails

                p "This content is shown when the toggle is on."

                p "Say Less handles all the DOM updates for you."

            else

                p "Click the button to show more content."


page "/about"

    main

        h1 "About Say Less Web"

        p "Say Less Web is a new way to build websites."

        p "No HTML. No CSS. No JavaScript."

        p "Just Say Less."

        a href "/" "Go back home"
