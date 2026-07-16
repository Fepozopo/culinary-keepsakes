package website

import (
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	homeTemplateFile       = "homepage.md"
	allRecipesTemplateFile = "all-recipes.md"
	authorsTemplateFile    = "authors.md"
	authorTemplateFile     = "author.md"
)

// Recipe is the validated metadata, optimized image filenames, and public URL for one recipe source file.
type Recipe struct {
	Title      string
	Author     string
	AuthorSlug string
	DateAdded  time.Time
	Image      string
	CardImage  string
	URL        string
}

// GenerateContentIndexes builds the recipe listing Markdown files under contentDirPath from recipe metadata and content templates.
// It validates recipe metadata, images, author slugs, and template placeholders before replacing generated pages.
func GenerateContentIndexes(contentDirPath string) error {
	recipes, err := readRecipes(contentDirPath)
	if err != nil {
		return err
	}

	templateDirPath := filepath.Join(contentDirPath, "_templates")
	homeTemplate, err := readContentTemplate(templateDirPath, homeTemplateFile)
	if err != nil {
		return err
	}
	allRecipesTemplate, err := readContentTemplate(templateDirPath, allRecipesTemplateFile)
	if err != nil {
		return err
	}
	authorsTemplate, err := readContentTemplate(templateDirPath, authorsTemplateFile)
	if err != nil {
		return err
	}
	authorTemplate, err := readContentTemplate(templateDirPath, authorTemplateFile)
	if err != nil {
		return err
	}

	recipesByAuthor, authors, err := groupRecipesByAuthor(recipes)
	if err != nil {
		return err
	}

	// Sort recipes by date added (descending) and then by title (ascending) for the homepage.
	recentRecipes := append([]Recipe(nil), recipes...)
	sort.Slice(recentRecipes, func(left, right int) bool {
		if recentRecipes[left].DateAdded.Equal(recentRecipes[right].DateAdded) {
			return recentRecipes[left].Title < recentRecipes[right].Title
		}
		return recentRecipes[left].DateAdded.After(recentRecipes[right].DateAdded)
	})
	if len(recentRecipes) > 6 {
		recentRecipes = recentRecipes[:6]
	}

	// Sort recipes alphabetically by title for the all-recipes page.
	alphabeticalRecipes := append([]Recipe(nil), recipes...)
	sortRecipesByTitle(alphabeticalRecipes)

	homepage, err := renderContentTemplate(homeTemplate, map[string]string{
		"{{ Recent Recipes }}": renderRecipeGridCards(recentRecipes),
	})
	if err != nil {
		return fmt.Errorf("failed to render homepage: %w", err)
	}
	if err := writeGeneratedContent(filepath.Join(contentDirPath, "index.md"), homepage); err != nil {
		return err
	}

	allRecipesPage, err := renderContentTemplate(allRecipesTemplate, map[string]string{
		"{{ All Recipes }}": renderRecipeCatalogCards(alphabeticalRecipes),
	})
	if err != nil {
		return fmt.Errorf("failed to render all-recipes page: %w", err)
	}
	if err := writeGeneratedContent(filepath.Join(contentDirPath, "all-recipes", "index.md"), allRecipesPage); err != nil {
		return err
	}

	authorsPage, err := renderContentTemplate(authorsTemplate, map[string]string{
		"{{ Authors }}": renderAuthorCards(authors),
	})
	if err != nil {
		return fmt.Errorf("failed to render authors page: %w", err)
	}
	if err := writeGeneratedContent(filepath.Join(contentDirPath, "authors", "index.md"), authorsPage); err != nil {
		return err
	}

	for _, author := range authors {
		authorRecipes := recipesByAuthor[author]
		sortRecipesByTitle(authorRecipes)
		authorPage, err := renderContentTemplate(authorTemplate, map[string]string{
			"{{ Author }}":         html.EscapeString(author),
			"{{ Author Recipes }}": renderRecipeGridCards(authorRecipes),
		})
		if err != nil {
			return fmt.Errorf("failed to render author page for %q: %w", author, err)
		}

		authorPath := filepath.Join(contentDirPath, "authors", normalizeSlug(author), "index.md")
		if err := writeGeneratedContent(authorPath, authorPage); err != nil {
			return err
		}
	}

	return nil
}

// readRecipes finds every recipe Markdown file and returns validated recipes in filesystem-independent order.
func readRecipes(contentDirPath string) ([]Recipe, error) {
	recipeDirPath := filepath.Join(contentDirPath, "recipes")
	staticImagesDirPath := filepath.Join(filepath.Dir(contentDirPath), "static", "images")
	recipes := []Recipe{}

	err := filepath.Walk(recipeDirPath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || info.Name() != "index.md" {
			return nil
		}

		recipe, err := parseRecipe(path, recipeDirPath, staticImagesDirPath)
		if err != nil {
			return err
		}
		recipes = append(recipes, recipe)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read recipes: %w", err)
	}
	if len(recipes) == 0 {
		return nil, errors.New("no recipes found")
	}

	return recipes, nil
}

// parseRecipe reads one recipe file and validates all metadata required for generated recipe listings.
func parseRecipe(recipePath, recipeDirPath, staticImagesDirPath string) (Recipe, error) {
	markdown, err := os.ReadFile(recipePath)
	if err != nil {
		return Recipe{}, fmt.Errorf("failed to read recipe %q: %w", recipePath, err)
	}

	metadata, body, err := parseFrontMatter(string(markdown))
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: %w", recipePath, err)
	}

	title, err := requiredMetadata(metadata, "title")
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: %w", recipePath, err)
	}
	author, err := requiredMetadata(metadata, "author")
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: %w", recipePath, err)
	}
	dateAdded, err := requiredMetadata(metadata, "date_added")
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: %w", recipePath, err)
	}
	image, err := requiredMetadata(metadata, "image")
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: %w", recipePath, err)
	}

	if filepath.Base(image) != image || image == "." {
		return Recipe{}, fmt.Errorf("recipe %q: image must be a filename in static/images", recipePath)
	}
	if strings.ToLower(filepath.Ext(image)) != ".webp" {
		return Recipe{}, fmt.Errorf("recipe %q: image must reference the WebP recipe-page version", recipePath)
	}
	if _, err := os.Stat(filepath.Join(staticImagesDirPath, image)); err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: recipe-page image %q is unavailable: %w", recipePath, image, err)
	}
	cardImage := cardImageFilename(image)
	if _, err := os.Stat(filepath.Join(staticImagesDirPath, cardImage)); err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: card thumbnail %q is unavailable: %w", recipePath, cardImage, err)
	}
	if !strings.Contains(body, "/images/"+image) {
		return Recipe{}, fmt.Errorf("recipe %q: listing image %q must also appear in the recipe body", recipePath, image)
	}

	heading, err := ExtractTitle(body)
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: %w", recipePath, err)
	}
	if heading != title {
		return Recipe{}, fmt.Errorf("recipe %q: metadata title %q does not match heading %q", recipePath, title, heading)
	}

	parsedDate, err := time.Parse("2006-01-02", dateAdded)
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: invalid date_added %q; use YYYY-MM-DD", recipePath, dateAdded)
	}

	relativeRecipeDir, err := filepath.Rel(recipeDirPath, filepath.Dir(recipePath))
	if err != nil {
		return Recipe{}, fmt.Errorf("recipe %q: failed to determine URL: %w", recipePath, err)
	}
	authorSlug := normalizeSlug(author)
	if authorSlug == "" {
		return Recipe{}, fmt.Errorf("recipe %q: author %q cannot be converted to a URL slug", recipePath, author)
	}

	return Recipe{
		Title:      title,
		Author:     author,
		AuthorSlug: authorSlug,
		DateAdded:  parsedDate,
		Image:      image,
		CardImage:  cardImage,
		URL:        "/recipes/" + filepath.ToSlash(relativeRecipeDir) + "/",
	}, nil
}

// cardImageFilename returns the thumbnail filename generated beside a recipe's full WebP image.
func cardImageFilename(image string) string {
	return strings.TrimSuffix(image, filepath.Ext(image)) + "-card.webp"
}

// parseFrontMatter separates an optional YAML-style metadata block from a Markdown document body.
func parseFrontMatter(markdown string) (map[string]string, string, error) {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	if !strings.HasPrefix(markdown, "---\n") {
		return map[string]string{}, markdown, nil
	}

	// The front matter block must be closed with a matching "---" delimiter on its own line.
	closingDelimiter := strings.Index(markdown[4:], "\n---\n")
	if closingDelimiter == -1 {
		return nil, "", errors.New("front matter is missing its closing delimiter")
	}
	closingDelimiter += 4
	metadataLines := strings.Split(markdown[4:closingDelimiter], "\n")
	// Each line must be a key-value pair separated by a colon, with optional whitespace and quotes.
	metadata := make(map[string]string, len(metadataLines))
	for _, line := range metadataLines {
		key, value, found := strings.Cut(line, ":")
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if !found || key == "" || value == "" {
			return nil, "", fmt.Errorf("invalid front matter line %q", line)
		}
		if _, exists := metadata[key]; exists {
			return nil, "", fmt.Errorf("duplicate front matter field %q", key)
		}
		metadata[key] = value
	}

	return metadata, markdown[closingDelimiter+5:], nil
}

// requiredMetadata returns a non-empty required metadata field or an actionable validation error.
func requiredMetadata(metadata map[string]string, field string) (string, error) {
	value := strings.TrimSpace(metadata[field])
	if value == "" {
		return "", fmt.Errorf("missing required %s metadata", field)
	}
	return value, nil
}

// groupRecipesByAuthor groups recipes by their exact display name and rejects different names that would share a URL slug.
func groupRecipesByAuthor(recipes []Recipe) (map[string][]Recipe, []string, error) {
	recipesByAuthor := make(map[string][]Recipe)
	authorNamesBySlug := make(map[string]string)
	for _, recipe := range recipes {
		if existingAuthor, exists := authorNamesBySlug[recipe.AuthorSlug]; exists && existingAuthor != recipe.Author {
			return nil, nil, fmt.Errorf("authors %q and %q both use the URL slug %q", existingAuthor, recipe.Author, recipe.AuthorSlug)
		}
		authorNamesBySlug[recipe.AuthorSlug] = recipe.Author
		recipesByAuthor[recipe.Author] = append(recipesByAuthor[recipe.Author], recipe)
	}

	authors := make([]string, 0, len(recipesByAuthor))
	for author := range recipesByAuthor {
		authors = append(authors, author)
	}
	sort.Slice(authors, func(left, right int) bool {
		return strings.ToLower(authors[left]) < strings.ToLower(authors[right])
	})
	return recipesByAuthor, authors, nil
}

// sortRecipesByTitle orders recipes by case-insensitive title, then author name, and finally URL for fully equal entries.
func sortRecipesByTitle(recipes []Recipe) {
	sort.Slice(recipes, func(left, right int) bool {
		leftTitle := strings.ToLower(recipes[left].Title)
		rightTitle := strings.ToLower(recipes[right].Title)
		if leftTitle != rightTitle {
			return leftTitle < rightTitle
		}
		leftAuthor := strings.ToLower(recipes[left].Author)
		rightAuthor := strings.ToLower(recipes[right].Author)
		if leftAuthor != rightAuthor {
			return leftAuthor < rightAuthor
		}
		return recipes[left].URL < recipes[right].URL
	})
}

// normalizeSlug converts a display name into the stable lowercase URL segment used for author pages.
func normalizeSlug(value string) string {
	var slug strings.Builder
	pendingSeparator := false
	for _, character := range strings.ToLower(value) {
		switch {
		case unicode.IsLetter(character) || unicode.IsNumber(character):
			if pendingSeparator && slug.Len() > 0 {
				slug.WriteByte('-')
			}
			slug.WriteRune(character)
			pendingSeparator = false
		case character == '\'':
			// Apostrophes are omitted so names such as "Peet's Coffee" retain their established slug.
		default:
			pendingSeparator = slug.Len() > 0
		}
	}
	return strings.Trim(slug.String(), "-")
}

// readContentTemplate loads one required template from content/_templates.
func readContentTemplate(templateDirPath, filename string) (string, error) {
	templatePath := filepath.Join(templateDirPath, filename)
	contents, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read content template %q: %w", templatePath, err)
	}
	return string(contents), nil
}

// renderContentTemplate replaces every required placeholder exactly once so template mistakes fail before generated pages are written.
func renderContentTemplate(template string, replacements map[string]string) (string, error) {
	result := template
	for placeholder, value := range replacements {
		if strings.Count(result, placeholder) != 1 {
			return "", fmt.Errorf("template must contain exactly one %s placeholder", placeholder)
		}
		result = strings.Replace(result, placeholder, value, 1)
	}
	return result, nil
}

// writeGeneratedContent writes a generated Markdown file after ensuring its parent directory exists.
func writeGeneratedContent(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create generated content directory for %q: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write generated content %q: %w", path, err)
	}
	return nil
}

// renderRecipeGridCards returns lazily loaded recipe-thumbnail links for the homepage and author-page grid layouts.
func renderRecipeGridCards(recipes []Recipe) string {
	var cards strings.Builder
	for _, recipe := range recipes {
		fmt.Fprintf(&cards, `  <a href="%s" style="text-align: center; text-decoration: none; color: inherit;">
    <img src="/images/%s" alt="%s" loading="lazy" decoding="async" style="width: 100%%; aspect-ratio: 1/1; object-fit: cover; max-width: 300px; margin: 0 auto; display: block; border-radius: 8px;" />
    <div style="margin-top: 0.5em; font-weight: bold;">%s</div>
    <div class="author">%s</div>
  </a>
`, recipe.URL, html.EscapeString(recipe.CardImage), html.EscapeString(recipe.Title), html.EscapeString(recipe.Title), html.EscapeString(recipe.Author))
	}
	return strings.TrimSuffix(cards.String(), "\n")
}

// renderRecipeCatalogCards returns lazily loaded recipe-thumbnail cards for the all-recipes catalog layout.
func renderRecipeCatalogCards(recipes []Recipe) string {
	var cards strings.Builder
	for _, recipe := range recipes {
		fmt.Fprintf(&cards, `  <div class="col col-4">
    <a class="card card-link" href="%s">
      <img src="/images/%s" alt="%s" loading="lazy" decoding="async">
      <div class="card-body">
        <div style="font-weight:bold;">%s</div>
        <div class="author">%s</div>
      </div>
    </a>
  </div>
`, recipe.URL, html.EscapeString(recipe.CardImage), html.EscapeString(recipe.Title), html.EscapeString(recipe.Title), html.EscapeString(recipe.Author))
	}
	return strings.TrimSuffix(cards.String(), "\n")
}

// renderAuthorCards returns alphabetically ordered author links for the authors directory.
func renderAuthorCards(authors []string) string {
	var cards strings.Builder
	for _, author := range authors {
		fmt.Fprintf(&cards, "  <div class=\"col col-4\"><a class=\"card card-link\" href=\"/authors/%s/\"><div class=\"card-body\">%s</div></a></div>\n", normalizeSlug(author), html.EscapeString(author))
	}
	return strings.TrimSuffix(cards.String(), "\n")
}

// stripFrontMatter removes an optional metadata block before Markdown rendering while preserving documents without metadata.
func stripFrontMatter(markdown string) (string, error) {
	_, body, err := parseFrontMatter(markdown)
	if err != nil {
		return "", err
	}
	return body, nil
}
