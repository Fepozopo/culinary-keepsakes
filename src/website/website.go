package website

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Fepozopo/culinary-keepsakes/src/blocks"
)

// CopyStaticToPublic copies all the contents of the static directory to the public directory recursively.
func CopyStaticToPublic() error {
	staticPath := "static"
	publicPath := "docs"

	// Remove all contents of the public directory if it exists
	if _, err := os.Stat(publicPath); err == nil {
		err = os.RemoveAll(publicPath)
		if err != nil {
			return err
		}
	}

	// Create the public directory if it doesn't exist
	err := os.MkdirAll(publicPath, os.ModePerm)
	if err != nil {
		return err
	}

	// Walk through the static directory and copy files to the public directory
	err = filepath.Walk(staticPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Determine the target path in the public directory
		relPath, err := filepath.Rel(staticPath, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(publicPath, relPath)

		if info.IsDir() {
			// Create directories in the public directory
			return os.MkdirAll(targetPath, os.ModePerm)
		} else {
			// Copy files to the public directory
			return copyFile(path, targetPath)
		}
	})

	return err
}

// copyFile copies the file at src to dst and returns any copy or close error.
func copyFile(src, dst string) (err error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = sourceFile.Close()
		}
	}()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		// Closing may flush buffered data, so report that failure when copying succeeded.
		if err == nil {
			err = destinationFile.Close()
		}
	}()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// ExtractTitle pulls the h1 header from the markdown and returns it as a string.
// If there is no h1 header, it returns an error.
func ExtractTitle(markdown string) (string, error) {
	// Split the markdown into blocks
	blocks := blocks.MarkdownToBlocks(markdown)

	// Find the h1 header and strip the "# " prefix
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if strings.HasPrefix(block, "# ") {
			return strings.TrimSpace(block[2:]), nil
		}
	}

	// If no h1 header is found, return an error
	return "", errors.New("no h1 header found")
}

// GeneratePage generates an HTML page from a Markdown file using a template.
// It removes optional recipe metadata before converting Markdown, writes the rendered page to destPath,
// and prefixes root-relative links with basepath.
func GeneratePage(fromPath, templatePath, destPath, basepath string) error {
	// Print a message to indicate that the page is being generated
	fmt.Printf("Generating page: %s -> %s using template: %s\n", fromPath, destPath, templatePath)

	// Read the markdown file at fromPath
	markdown, err := os.ReadFile(fromPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Metadata drives listing generation but must not become visible recipe-page content.
	contentMarkdown, err := stripFrontMatter(string(markdown))
	if err != nil {
		return fmt.Errorf("failed to parse markdown metadata: %w", err)
	}

	// Read the template file
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Convert the Markdown body to an HTML string.
	htmlNode := blocks.MarkdownToHTMLNode(contentMarkdown)
	html, err := htmlNode.ToHTML()
	if err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Extract the title from the Markdown body.
	title, err := ExtractTitle(contentMarkdown)
	if err != nil {
		return fmt.Errorf("failed to extract title: %w", err)
	}

	// Replace the placeholders in the template with the extracted title and HTML string
	result := strings.ReplaceAll(string(template), "{{ Title }}", title)
	result = strings.ReplaceAll(result, "{{ Content }}", html)

	// Replace href and src attributes with basepath
	result = strings.ReplaceAll(result, "href=\"/", fmt.Sprintf("href=\"%s", basepath))
	result = strings.ReplaceAll(result, "src=\"/", fmt.Sprintf("src=\"%s", basepath))

	// Write the result to the destination file
	err = os.WriteFile(destPath, []byte(result), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to destination file: %w", err)
	}

	return nil
}

// GeneratePagesRecursive generates HTML pages from publishable Markdown and HTML files in contentDirPath recursively.
// It excludes content/_templates because those files are build inputs rather than public pages, writes output to destDirPath,
// and prefixes root-relative links with basepath.
func GeneratePagesRecursive(contentDirPath, templatePath, destDirPath, basepath string) error {
	// Ensure the destination directory exists
	err := os.MkdirAll(destDirPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk through the content directory
	err = filepath.Walk(contentDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Templates are source-only inputs and must never become public pages.
		if info.IsDir() && filepath.Base(path) == "_templates" {
			return filepath.SkipDir
		}

		// Skip directories, but ensure publishable subdirectories are created in the destination path.
		if info.IsDir() {
			relPath, err := filepath.Rel(contentDirPath, path)
			if err != nil {
				return err
			}
			newDestDir := filepath.Join(destDirPath, relPath)
			return os.MkdirAll(newDestDir, os.ModePerm)
		}

		// Process markdown files
		if strings.HasSuffix(info.Name(), ".md") {
			relPath, err := filepath.Rel(contentDirPath, path)
			if err != nil {
				return err
			}
			outputFile := filepath.Join(destDirPath, strings.TrimSuffix(relPath, ".md")+".html")
			return GeneratePage(path, templatePath, outputFile, basepath)
		}

		// Process html files
		if strings.HasSuffix(info.Name(), ".html") {
			relPath, err := filepath.Rel(contentDirPath, path)
			if err != nil {
				return err
			}
			outputFile := filepath.Join(destDirPath, relPath)
			return copyFile(path, outputFile)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}

	// Print a message to indicate that all pages have been generated
	fmt.Printf("All pages have been generated in %s\n", destDirPath)
	return nil
}
