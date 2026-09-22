# Home Page
fn render()
    return """
        <link rel="preconnect" href="https://fonts.googleapis.com">
        <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
        <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
        <link rel="stylesheet" href="/styles/base.css">
        <link rel="stylesheet" href="/styles/page.css">

        <div class="hero">
            <h1 id="page-title" class="float-anim">Welcome</h1>
            <p>The <span class="accent underline-reveal">Say Less</span> way to build.</p>
        </div>

        <script>
            gsap.fromTo("#page-title",
                { y: 40, opacity: 0, scale: 0.9 },
                { y: 0, opacity: 1, scale: 1, duration: 1, ease: "back.out(1.4)" }
            );

            gsap.fromTo(".hero p",
                { y: 20, opacity: 0 },
                { y: 0, opacity: 1, duration: 0.8, delay: 0.3, ease: "power2.out" }
            );
        </script>
        """
