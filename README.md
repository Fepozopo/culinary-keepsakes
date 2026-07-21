# Culinary Keepsakes

Culinary Keepsakes is a project designed to store recipes I've come across over the years that I thought were worth saving.

## Features

- Store and manage recipes using markdown files.
- View recipes in a user-friendly interface.
- Customizable build process using `build.sh`.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/Fepozopo/culinary-keepsakes.git
   ```
2. Navigate to the project directory:
   ```bash
   cd culinary-keepsakes
   ```

## Usage

1. Add a recipe Markdown file at `content/recipes/<recipe-slug>/<author-slug>/index.md`. Begin the file with the required metadata:
   ```md
   ---
   title: Recipe Title
   author: Author Name
   date_added: YYYY-MM-DD
   image: recipe-image.webp
   categories: [Dessert, Holiday]
   ---
   ```
   The optional `categories` field is a free-form YAML flow list. A recipe is listed on every generated category page named in that list. The `date_added` value is the date the recipe was added to this site and should not change when the recipe is edited. The title must match the recipe's `#` heading, and the declared image must also appear in the recipe body.
2. Place a PNG, JPEG, or WebP source image in `new_images/` before running the build script. The script creates an optimized recipe-page WebP image (`recipe-image.webp`) and a 600×600 cropped card thumbnail (`recipe-image-card.webp`) in `static/images/`, then moves the source image to `backup/original_images/`. Use the recipe-page `.webp` filename in both the metadata and recipe image Markdown; card pages use the derived thumbnail automatically.
3. Build the project:
   ```bash
   ./build.sh
   ```
   The build regenerates `content/index.md`, `content/all-recipes/index.md`, `content/authors/`, `content/categories/`, and their individual listing pages before producing the committed HTML files under `docs/`. It automatically creates author and category pages from recipe metadata.
4. Optionally choose to run the local server, then open `http://localhost:8080/`.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch:
   ```bash
   git checkout -b feature-name
   ```
3. Commit your changes:
   ```bash
   git commit -m "Add feature-name"
   ```
4. Push to your branch:
   ```bash
   git push origin feature-name
   ```
5. Open a pull request.

## Contact

For questions or feedback, please reach out to [Fepozopo](https://github.com/Fepozopo).
