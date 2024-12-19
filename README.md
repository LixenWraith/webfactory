# WebFactory

A component-based static website generator that assembles web pages from reusable components.

## Overview

webfactory is designed to build static websites using a component-based architecture. It processes blueprint files that define page structure and component relationships, combining HTML templates with their associated assets (CSS and JavaScript) to generate complete web pages.

## Key Features

- Component-based architecture
- Blueprint-driven page assembly
- Automatic asset management (CSS/JS)
- Template processing with variable support
- Range/loop functionality in templates
- Asset deduplication and optimization

## Project Structure

```
input/
├── components/        # Project component
│   ├── layout/        # Layout components
│   └── composite/     # Composite comopnent group
│       └── card       # Card component
└── blueprints/        # Page blueprints
output/                # Generated site
```

## Blueprints

Blueprints define the page structure using a simple syntax:

```
1 sample.card
.header=Card Title
.content=Card content goes here
1.1 sample.button
.label=Click Me
```

- Numbers define component hierarchy (1, 1.1, 1.2, etc.)
- Component paths use dot notation
- Variables are prefixed with a dot

## Components

Components consist of HTML templates with optional CSS and JavaScript:

```html
<div class="card">
    <h3>{{.header}}</h3>
    <p>{{.content}}</p>
    {{component}}
</div>
```

Special directives:
- `{{.varname}}` - Variable substitution
- `{{component}}` - Child component insertion
- `{{styles}}` - CSS insertion point
- `{{script}}` - JavaScript insertion point
- `{{range .var}}...{{range end}}` - Loop construct

## Usage

```bash
webfactory -s /path/to/source -t /path/to/output
```

## License

MIT License

> Note: This project is in early development and anything may change in the future.