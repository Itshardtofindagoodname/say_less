# Web Application
use http
use env
use npm:gsap

server on 8080

get "/"
    return """
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Say Less Web App</title>
            <link rel="preconnect" href="https://fonts.googleapis.com">
            <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
            <link href="https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
            <link rel="stylesheet" href="/styles/base.css">
            <link rel="stylesheet" href="/styles/main.css">
        </head>
        <body>
            <div class="glow-line" id="glowLine"></div>
            <div class="cursor-glow" id="cursorGlow"></div>
            <div class="particles" id="particles"></div>

            <div class="hero">
                <p class="tagline" id="tagline">// shipping since 2026</p>
                <h1 id="title">Say Less.</h1>
                <p class="subtitle" id="subtitle">Build more. Write <span>less code</span>.</p>
                <a href="#" class="cta-btn" id="cta">Get Started</a>
            </div>

            <script>
                const tl = gsap.timeline({ defaults: { ease: "power3.out" } });

                tl.fromTo("#glowLine", { scaleX: 0 }, { scaleX: 1, opacity: 1, duration: 1.2 })
                  .fromTo("#title", { y: 60, opacity: 0, skewY: 3 }, { y: 0, opacity: 1, skewY: 0, duration: 1 }, "-=0.6")
                  .fromTo("#tagline", { y: 20, opacity: 0 }, { y: 0, opacity: 1, duration: 0.7 }, "-=0.5")
                  .fromTo("#subtitle", { y: 30, opacity: 0 }, { y: 0, opacity: 1, duration: 0.7 }, "-=0.3")
                  .fromTo("#cta", { y: 20, opacity: 0, scale: 0.95 }, { y: 0, opacity: 1, scale: 1, duration: 0.6 }, "-=0.2");

                gsap.to("h1", {
                    backgroundPosition: "100% 50%",
                    duration: 4,
                    repeat: -1,
                    yoyo: true,
                    ease: "sine.inOut"
                });

                const particlesEl = document.getElementById("particles");
                for (let i = 0; i < 40; i++) {
                    const p = document.createElement("div");
                    p.className = "particle";
                    p.style.left = Math.random() * 100 + "%";
                    p.style.top = Math.random() * 100 + "%";
                    p.style.opacity = Math.random() * 0.5 + 0.1;
                    p.style.width = p.style.height = (Math.random() * 3 + 1) + "px";
                    particlesEl.appendChild(p);

                    gsap.to(p, {
                        y: () => gsap.utils.random(-80, 80),
                        x: () => gsap.utils.random(-40, 40),
                        opacity: () => gsap.utils.random(0.05, 0.4),
                        duration: () => gsap.utils.random(4, 10),
                        repeat: -1,
                        yoyo: true,
                        ease: "sine.inOut",
                        delay: Math.random() * 3
                    });
                }

                document.addEventListener("mousemove", (e) => {
                    gsap.to("#cursorGlow", { x: e.clientX, y: e.clientY, duration: 0.8, ease: "power2.out" });
                });

                gsap.fromTo("#glowLine", { scaleX: 0, transformOrigin: "left" }, { scaleX: 1, opacity: 1, duration: 1.5, ease: "power2.inOut" });
            </script>
        </body>
        </html>
        """
