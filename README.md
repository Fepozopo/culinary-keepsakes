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

1. Add new recipes by creating a new folder in `content/recipes` and adding a markdown file with the recipe details.
2. Place any new images in the `new_images` folder before running the build script. This ensures they are processed and included correctly.
3. Add the recipe to the all-recipes index by editing `content/all-recipes/index.md` to include a link to the new recipe.
4. Update the homepage `content/index.md` to feature the new recipe. Recipe entries should be added in the grid layout section, which should show the most recent recipes.
5. Build the project using the updated `build.sh` script:
   ```bash
   ./build.sh
   ```
   The script will generate the necessary files and start a local server.
6. Open your browser and navigate to `http://localhost:8080/`.

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
