from datetime import datetime

project = "mmmbed"
master_doc = "index"
extensions = ["myst_parser", "hawkmoth"]
source_suffix = {".md": "markdown", ".rst": "restructuredtext"}
html_theme = "alabaster"
copyright = f"Zachary Fogg {datetime.now().year}"
templates_path = ["_templates"]
html_static_path = ["_static"]
html_css_files = ["custom.css"]
html_sidebars = {"**": ["searchbox.html", "localtoc.html"]}
exclude_patterns = [".git", "build", "bin"]
