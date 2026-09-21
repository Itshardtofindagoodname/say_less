# Web Application
use http

server on 8080

get "/"
    return html
        """
        <!DOCTYPE html>
        <html>
        <head><title>Say Less Web App</title></head>
        <body>
            <h1>Hello from Say Less!</h1>
            <p>Build more. Say less.</p>
        </body>
        </html>
        """
