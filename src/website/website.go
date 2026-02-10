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

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

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

// GeneratePage generates an HTML page from a markdown file using a template.
// It reads the markdown file, converts it to HTML, and replaces placeholders in the template.
// The generated HTML is written to the destination path.
// The basepath is used to replace href and src attributes in the generated HTML.
func GeneratePage(fromPath, templatePath, destPath, basepath string) error {
	// Print a message to indicate that the page is being generated
	fmt.Printf("Generating page: %s -> %s using template: %s\n", fromPath, destPath, templatePath)

	// Read the markdown file at fromPath
	markdown, err := os.ReadFile(fromPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Read the template file
	template, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	// Convert the markdown to an HTML string
	htmlNode := blocks.MarkdownToHTMLNode(string(markdown))
	html, err := htmlNode.ToHTML()
	if err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Extract the title from the markdown
	title, err := ExtractTitle(string(markdown))
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

// GeneratePagesRecursive generates HTML pages from markdown files in the content directory recursively.
// It uses the specified template and writes the output to the destination directory.
// The basepath is used to replace href and src attributes in the generated HTML.
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

		// Skip directories, but ensure subdirectories are created in the destination path
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
