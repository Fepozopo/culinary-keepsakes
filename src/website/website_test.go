package website

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractTitleWithH1Header verifies that ExtractTitle returns the first level-one heading.
func TestExtractTitleWithH1Header(t *testing.T) {
	markdown := `
    # This is the title
    `

	result, err := ExtractTitle(markdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "This is the title"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestExtractTitleWithNoH1Header verifies that ExtractTitle rejects Markdown without a level-one heading.
func TestExtractTitleWithNoH1Header(t *testing.T) {
	markdown := `
    This is the title
    `

	_, err := ExtractTitle(markdown)
	if err == nil {
		t.Fatal("expected an error but got none")
	}

	expectedError := "no h1 header found"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

// TestExtractTitleWithWhitespace verifies that ExtractTitle tolerates surrounding Markdown whitespace.
func TestExtractTitleWithWhitespace(t *testing.T) {
	markdown := `
        # This is a heading

        This is a paragraph of text. It has some **bold** and *italic* words inside of it.

        * This is the first list item in a list block
        * This is a list item
        * This is another list item
    `

	result, err := ExtractTitle(markdown)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "This is a heading"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

// TestGenerateContentIndexesBuildsMetadataDrivenListings verifies generated pages use date, title, image, and author metadata.
func TestGenerateContentIndexesBuildsMetadataDrivenListings(t *testing.T) {
	rootDir := t.TempDir()
	contentDir := filepath.Join(rootDir, "content")

	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", homeTemplateFile), "# Home\n\n{{ Recent Recipes }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", allRecipesTemplateFile), "# All Recipes\n\n{{ All Recipes }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", authorsTemplateFile), "# Authors\n\n{{ Authors }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", authorTemplateFile), "# {{ Author }}'s Recipes\n\n{{ Author Recipes }}\n")
	writeWebsiteTestFile(t, filepath.Join(rootDir, "static", "images", "apple.webp"), "image")
	writeWebsiteTestFile(t, filepath.Join(rootDir, "static", "images", "zucchini.webp"), "image")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "recipes", "apple", "alice", "index.md"), `---
title: Apple Pie
author: Alice Example
date_added: 2025-01-10
image: apple.webp
---

# Apple Pie

![Apple Pie](/images/apple.webp)
`)
	writeWebsiteTestFile(t, filepath.Join(contentDir, "recipes", "zucchini", "zed", "index.md"), `---
title: Zucchini Bread
author: Zed Example
date_added: 2025-02-10
image: zucchini.webp
---

# Zucchini Bread

![Zucchini Bread](/images/zucchini.webp)
`)

	if err := GenerateContentIndexes(contentDir); err != nil {
		t.Fatalf("GenerateContentIndexes returned an error: %v", err)
	}

	homepage := readWebsiteTestFile(t, filepath.Join(contentDir, "index.md"))
	if strings.Index(homepage, "Zucchini Bread") > strings.Index(homepage, "Apple Pie") {
		t.Error("homepage recipes are not ordered newest first")
	}

	allRecipes := readWebsiteTestFile(t, filepath.Join(contentDir, "all-recipes", "index.md"))
	if strings.Index(allRecipes, "Apple Pie") > strings.Index(allRecipes, "Zucchini Bread") {
		t.Error("all recipes are not ordered alphabetically")
	}

	authors := readWebsiteTestFile(t, filepath.Join(contentDir, "authors", "index.md"))
	if strings.Index(authors, "Alice Example") > strings.Index(authors, "Zed Example") {
		t.Error("authors are not ordered alphabetically")
	}

	zedAuthorPage := readWebsiteTestFile(t, filepath.Join(contentDir, "authors", "zed-example", "index.md"))
	if !strings.Contains(zedAuthorPage, "Zucchini Bread") {
		t.Error("a new author page did not include that author's recipe")
	}
}

// TestGenerateContentIndexesRejectsMissingListingImage verifies that recipes cannot generate listings without an available image.
func TestGenerateContentIndexesRejectsMissingListingImage(t *testing.T) {
	rootDir := t.TempDir()
	contentDir := filepath.Join(rootDir, "content")

	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", homeTemplateFile), "# Home\n\n{{ Recent Recipes }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", allRecipesTemplateFile), "# All Recipes\n\n{{ All Recipes }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", authorsTemplateFile), "# Authors\n\n{{ Authors }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "_templates", authorTemplateFile), "# {{ Author }}'s Recipes\n\n{{ Author Recipes }}\n")
	writeWebsiteTestFile(t, filepath.Join(contentDir, "recipes", "missing-image", "example", "index.md"), `---
title: Missing Image
author: Example Author
date_added: 2025-01-10
image: missing.webp
---

# Missing Image

![Missing Image](/images/missing.webp)
`)

	err := GenerateContentIndexes(contentDir)
	if err == nil || !strings.Contains(err.Error(), "listing image") {
		t.Fatalf("expected a missing listing image error, got %v", err)
	}
}

// TestSortRecipesByTitleUsesAuthorAsTieBreaker verifies identical titles sort by their authors before URLs.
func TestSortRecipesByTitleUsesAuthorAsTieBreaker(t *testing.T) {
	recipes := []Recipe{
		{Title: "Chocolate Chip Cookies", Author: "Zed Baker", URL: "/recipes/cookies/zed/"},
		{Title: "Chocolate Chip Cookies", Author: "Alice Baker", URL: "/recipes/cookies/alice/"},
		{Title: "Apple Pie", Author: "Zed Baker", URL: "/recipes/apple-pie/zed/"},
	}

	sortRecipesByTitle(recipes)

	if recipes[0].Title != "Apple Pie" || recipes[1].Author != "Alice Baker" || recipes[2].Author != "Zed Baker" {
		t.Fatalf("recipes were not ordered by title and then author: %#v", recipes)
	}
}

// writeWebsiteTestFile creates a test fixture file and its parent directories.
func writeWebsiteTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}
}

// readWebsiteTestFile returns a required test fixture file.
func readWebsiteTestFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}
	return string(contents)
}
