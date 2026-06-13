import os
import re
import yaml

def define_env(env):
    """
    Define custom macros for the MkDocs site.
    """
    @env.macro
    def pagenav():
        """
        Dynamically generates a grid of navigation cards for every document listed
        in mkdocs.yml, extracting titles, descriptions, and subsection links.
        """
        docs_dir = "."
        mkdocs_yml_path = "mkdocs.yml"

        if not os.path.exists(mkdocs_yml_path):
            return "*(Error: mkdocs.yml not found)*"

        # Load mkdocs.yml configuration (PyYAML is guaranteed to be available in MkDocs env)
        with open(mkdocs_yml_path, "r", encoding="utf-8") as f:
            try:
                config = yaml.safe_load(f)
            except Exception as e:
                return f"*(Error parsing mkdocs.yml: {e})*"

        nav = config.get("nav", [])
        if not nav:
            return "*(No navigation found in mkdocs.yml)*"

        # Flatten the navigation tree into a list of (Title, FilePath)
        pages = []
        def parse_nav_item(item):
            if isinstance(item, str):
                pages.append((None, item))
            elif isinstance(item, dict):
                for key, val in item.items():
                    if isinstance(val, str):
                        pages.append((key, val))
                    elif isinstance(val, list):
                        for sub_item in val:
                            parse_nav_item(sub_item)

        for item in nav:
            parse_nav_item(item)

        icon_mapping = {
            "index.md": "🏠",
            "development.md": "💻",
            "distribution.md": "☸️",
            "setup_template.md": "🛠️",
            "default": "📄"
        }

        def get_icon(filename):
            lower_name = filename.lower()
            for key, icon in icon_mapping.items():
                if key in lower_name:
                    return icon
            return icon_mapping["default"]

        cards = []
        cards.append('<div class="card-grid">')

        for title, filepath in pages:
            # Skip the available content directory page itself to prevent recursive loop
            if "available_content.md" in filepath:
                continue

            full_path = os.path.join(docs_dir, filepath)
            if not os.path.exists(full_path):
                continue

            # Read document contents to extract metadata dynamically
            with open(full_path, "r", encoding="utf-8") as f_md:
                lines = f_md.readlines()

            md_title = title
            if not md_title:
                for line in lines:
                    match = re.match(r"^#\s+(.+)$", line)
                    if match:
                        md_title = match.group(1).strip()
                        break
                if not md_title:
                    md_title = os.path.basename(filepath)

            # Extract the first blockquote or text paragraph as description
            description = ""
            found_title = False
            for line in lines:
                striped = line.strip()
                if not found_title:
                    if striped.startswith("# "):
                        found_title = True
                    continue
                if striped:
                    if striped.startswith(">"):
                        description = striped.lstrip("> ").strip()
                        break
                    elif not striped.startswith("#") and not striped.startswith("---") and not striped.startswith("```"):
                        description = striped
                        break
            if len(description) > 130:
                description = description[:127] + "..."

            # Find all H2 headers for anchor links
            sections = []
            for line in lines:
                match = re.match(r"^##\s+(.+)$", line)
                if match:
                    header_text = match.group(1).strip()
                    anchor = header_text.lower()
                    anchor = re.sub(r"[^\w\s\-]", "", anchor)
                    anchor = re.sub(r"[\s\_]+", "-", anchor)
                    sections.append((header_text, anchor))

            icon = get_icon(filepath)
            list_items = []
            for sec_title, anchor in sections[:5]:
                list_items.append(f'      <li><a href="../{filepath}#{anchor}">{sec_title}</a></li>')
            if len(sections) > 5:
                list_items.append(f'      <li><a href="../{filepath}">... and {len(sections) - 5} more sections</a></li>')

            list_content = "\n".join(list_items)
            card = f"""  <!-- Card: {md_title} -->
  <div class="card-item">
    <h3>{icon} <a href="../{filepath}">{md_title}</a></h3>
    <p>{description or 'Documentation guide.'}</p>
    <hr>
    <ul>
{list_content}
    </ul>
  </div>"""
            cards.append(card)

        cards.append('</div>')
        return "\n".join(cards)
