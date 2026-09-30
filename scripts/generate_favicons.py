import os
import subprocess
from PIL import Image

SVG_CONTENT = """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
  <defs>
    <linearGradient id="cspGrad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#6366f1"/>
      <stop offset="100%" stop-color="#4338ca"/>
    </linearGradient>
    <linearGradient id="borderGrad" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#a5b4fc" stop-opacity="0.6"/>
      <stop offset="100%" stop-color="#3730a3" stop-opacity="0.8"/>
    </linearGradient>
    <filter id="subtleGlow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="4" stdDeviation="8" flood-color="#1e1b4b" flood-opacity="0.35"/>
    </filter>
  </defs>
  <!-- Background Squircle -->
  <rect x="32" y="32" width="448" height="448" rx="112" fill="url(#cspGrad)" stroke="url(#borderGrad)" stroke-width="12" filter="url(#subtleGlow)"/>
  
  <!-- Centered Shield + Checkmark -->
  <g transform="translate(256 256) scale(15) translate(-12 -12)">
    <!-- Inner shield subtle tint for depth -->
    <path d="M12 2.944a11.955 11.955 0 018.618 3.04A12.02 12.02 0 0121 9c0 5.591-3.824 10.29-9 11.622C6.824 19.29 3 14.591 3 9c0-1.042.133-2.052.382-3.016A11.955 11.955 0 0112 2.944z"
          fill="#312e81" fill-opacity="0.25"/>
    <!-- Outer Shield Stroke & Checkmark -->
    <path d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
          fill="none" stroke="#ffffff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
  </g>
</svg>
"""

WEBMANIFEST_CONTENT = """{
  "name": "CSP Scout",
  "short_name": "CSP Scout",
  "description": "CSP Collector & Policy Generator",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#020617",
  "theme_color": "#4f46e5",
  "icons": [
    {
      "src": "/favicon-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/favicon-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
"""

def main():
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    static_dir = os.path.join(root_dir, "frontend", "static")
    assets_dir = os.path.join(root_dir, "frontend", "src", "lib", "assets")
    os.makedirs(static_dir, exist_ok=True)
    os.makedirs(assets_dir, exist_ok=True)

    # 1. Write SVG to static/favicon.svg and src/lib/assets/favicon.svg
    svg_static_path = os.path.join(static_dir, "favicon.svg")
    svg_assets_path = os.path.join(assets_dir, "favicon.svg")
    with open(svg_static_path, "w", encoding="utf-8") as f:
        f.write(SVG_CONTENT)
    with open(svg_assets_path, "w", encoding="utf-8") as f:
        f.write(SVG_CONTENT)
    print(f"Wrote {svg_static_path}")
    print(f"Wrote {svg_assets_path}")

    # 2. Write site.webmanifest
    manifest_path = os.path.join(static_dir, "site.webmanifest")
    with open(manifest_path, "w", encoding="utf-8") as f:
        f.write(WEBMANIFEST_CONTENT)
    print(f"Wrote {manifest_path}")

    # 3. Render 512x512 master raster via Chrome headless
    temp_png = os.path.join(static_dir, "favicon-512x512.png")
    chrome_cmd = [
        r"C:\Program Files\Google\Chrome\Application\chrome.exe",
        "--headless",
        "--disable-gpu",
        "--default-background-color=00000000",
        f"--screenshot={temp_png}",
        "--window-size=512,512",
        f"file:///{svg_static_path.replace(os.sep, '/')}"
    ]
    print("Running Chrome headless screenshot...")
    res = subprocess.run(chrome_cmd, capture_output=True, text=True)
    if res.returncode != 0:
        print(f"Chrome error: {res.stderr}")
        return

    # 4. Generate raster sizes with Pillow Lanczos resampling
    master = Image.open(temp_png)
    
    # 192x192
    p192 = master.resize((192, 192), Image.Resampling.LANCZOS)
    p192.save(os.path.join(static_dir, "favicon-192x192.png"))

    # 180x180 (Apple touch icon)
    p180 = master.resize((180, 180), Image.Resampling.LANCZOS)
    p180.save(os.path.join(static_dir, "apple-touch-icon.png"))

    # 32x32
    p32 = master.resize((32, 32), Image.Resampling.LANCZOS)
    p32.save(os.path.join(static_dir, "favicon-32x32.png"))

    # 16x16
    p16 = master.resize((16, 16), Image.Resampling.LANCZOS)
    p16.save(os.path.join(static_dir, "favicon-16x16.png"))

    # Multi-resolution ICO (16, 32, 48)
    ico_path = os.path.join(static_dir, "favicon.ico")
    master.save(ico_path, format="ICO", sizes=[(16, 16), (32, 32), (48, 48)])
    print(f"Generated all PNGs and multi-res ICO at {static_dir}")

if __name__ == "__main__":
    main()
