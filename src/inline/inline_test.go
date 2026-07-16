package inline

import (
	"reflect"
	"testing"

	"github.com/Fepozopo/culinary-keepsakes/src/nodes"
)

func TestSplitNodesDelimiter(t *testing.T) {
	node := nodes.TextNode{
		Type: nodes.Normal,
		Text: "This is text with a `code block` word",
		URL:  "",
	}
	result, err := SplitNodesDelimiter([]nodes.TextNode{node}, "`", nodes.Code)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected := []nodes.TextNode{
		{Type: nodes.Normal, Text: "This is text with a ", URL: ""},
		{Type: nodes.Code, Text: "code block", URL: ""},
		{Type: nodes.Normal, Text: " word", URL: ""},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	node = nodes.TextNode{
		Type: nodes.Normal,
		Text: "This is **bold** text",
		URL:  "",
	}
	result, err = SplitNodesDelimiter([]nodes.TextNode{node}, "**", nodes.Bold)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected = []nodes.TextNode{
		{Type: nodes.Normal, Text: "This is ", URL: ""},
		{Type: nodes.Bold, Text: "bold", URL: ""},
		{Type: nodes.Normal, Text: " text", URL: ""},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	node = nodes.TextNode{
		Type: nodes.Normal,
		Text: "This is *italic* text",
		URL:  "",
	}
	result, err = SplitNodesDelimiter([]nodes.TextNode{node}, "*", nodes.Italic)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected = []nodes.TextNode{
		{Type: nodes.Normal, Text: "This is ", URL: ""},
		{Type: nodes.Italic, Text: "italic", URL: ""},
		{Type: nodes.Normal, Text: " text", URL: ""},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Unmatched delimiter (should raise an error)
	node = nodes.TextNode{
		Type: nodes.Normal,
		Text: "This is unmatched `code block text",
		URL:  "",
	}
	_, err = SplitNodesDelimiter([]nodes.TextNode{node}, "`", nodes.Code)
	if err == nil {
		t.Errorf("Expected error, got none")
	}

	// No splitting needed (no delimiter present)
	node = nodes.TextNode{
		Type: nodes.Normal,
		Text: "No delimiter here",
		URL:  "",
	}
	result, err = SplitNodesDelimiter([]nodes.TextNode{node}, "`", nodes.Code)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected = []nodes.TextNode{node}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Mixed nodes (only "text" nodes are split)
	node1 := nodes.TextNode{
		Type: nodes.Normal,
		Text: "Normal text",
		URL:  "",
	}
	node2 := nodes.TextNode{
		Type: nodes.Bold,
		Text: "**bold** text",
		URL:  "",
	}
	node3 := nodes.TextNode{
		Type: nodes.Normal,
		Text: "Code `inline` text",
		URL:  "",
	}
	result, err = SplitNodesDelimiter([]nodes.TextNode{node1, node2, node3}, "`", nodes.Code)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	expected = []nodes.TextNode{
		node1,
		node2,
		{Type: nodes.Normal, Text: "Code ", URL: ""},
		{Type: nodes.Code, Text: "inline", URL: ""},
		{Type: nodes.Normal, Text: " text", URL: ""},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestSplitNodesUnderscoreItalic verifies underscore-delimited italics without treating identifier underscores as emphasis.
func TestSplitNodesUnderscoreItalic(t *testing.T) {
	node := nodes.TextNode{Type: nodes.Normal, Text: "Use _italics_ but preserve recipe_name.", URL: ""}

	result := SplitNodesUnderscoreItalic([]nodes.TextNode{node})
	expected := []nodes.TextNode{
		{Type: nodes.Normal, Text: "Use ", URL: ""},
		{Type: nodes.Italic, Text: "italics", URL: ""},
		{Type: nodes.Normal, Text: " but preserve recipe_name.", URL: ""},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

// TestTextToTextNodesSupportsUnderscoreItalic verifies the full inline parser recognizes underscore-delimited italics.
func TestTextToTextNodesSupportsUnderscoreItalic(t *testing.T) {
	result := TextToTextNodes("An _italic_ word")
	expected := []nodes.TextNode{
		{Type: nodes.Normal, Text: "An ", URL: ""},
		{Type: nodes.Italic, Text: "italic", URL: ""},
		{Type: nodes.Normal, Text: " word", URL: ""},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

// TestExtractMarkdownImages verifies image URL and alternative-text extraction.
func TestExtractMarkdownImages(t *testing.T) {
	t.Run("extract_markdown_images", func(t *testing.T) {
		text := "This is text with a ![rick roll](https://i.imgur.com/aKaOqIh.gif) and ![obi wan](https://i.imgur.com/fJRm4Vk.jpeg)"
		result := ExtractMarkdownImages(text)
		expected := [][2]string{{"rick roll", "https://i.imgur.com/aKaOqIh.gif"}, {"obi wan", "https://i.imgur.com/fJRm4Vk.jpeg"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("extractMarkdownImages() = %v, want %v", result, expected)
		}
	})

	t.Run("extract_markdown_images_with_no_images", func(t *testing.T) {
		text := "This is text with no images"
		result := ExtractMarkdownImages(text)
		expected := [][2]string{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("extractMarkdownImages() = %v, want %v", result, expected)
		}
	})
}

func TestExtractMarkdownLinks(t *testing.T) {
	t.Run("extract_markdown_links", func(t *testing.T) {
		text := "This is text with a link [to boot dev](https://www.boot.dev) and [to youtube](https://www.youtube.com/@bootdotdev)"
		result := ExtractMarkdownLinks(text)
		expected := [][2]string{{"to boot dev", "https://www.boot.dev"}, {"to youtube", "https://www.youtube.com/@bootdotdev"}}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("extractMarkdownLinks() = %v, want %v", result, expected)
		}
	})

	t.Run("extract_markdown_links_with_no_links", func(t *testing.T) {
		text := "This is text with no links"
		result := ExtractMarkdownLinks(text)
		expected := [][2]string{}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("extractMarkdownLinks() = %v, want %v", result, expected)
		}
	})
}

func TestSplitNodesImage(t *testing.T) {
	t.Run("split_nodes_image", func(t *testing.T) {
		node := nodes.TextNode{
			Type: nodes.Normal,
			Text: "This is text with a ![rick roll](https://i.imgur.com/aKaOqIh.gif) image",
			URL:  "",
		}
		result := SplitNodesImage([]nodes.TextNode{node})
		expected := []nodes.TextNode{
			{Type: nodes.Normal, Text: "This is text with a ", URL: ""},
			{Type: nodes.Image, Text: "rick roll", URL: "https://i.imgur.com/aKaOqIh.gif"},
			{Type: nodes.Normal, Text: " image", URL: ""},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("split_nodes_image_with_no_images", func(t *testing.T) {
		node := nodes.TextNode{
			Type: nodes.Normal,
			Text: "This is text with no images",
			URL:  "",
		}
		result := SplitNodesImage([]nodes.TextNode{node})
		expected := []nodes.TextNode{node}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("split_nodes_image_with_mixed_nodes", func(t *testing.T) {
		node1 := nodes.TextNode{Type: nodes.Normal, Text: "Normal text", URL: ""}
		node2 := nodes.TextNode{Type: nodes.Normal, Text: "![](https://i.imgur.com/aKaOqIh.gif)", URL: ""}
		node3 := nodes.TextNode{Type: nodes.Normal, Text: "Code `inline` text", URL: ""}
		result := SplitNodesImage([]nodes.TextNode{node1, node2, node3})
		expected := []nodes.TextNode{
			node1,
			{Type: nodes.Image, Text: "", URL: "https://i.imgur.com/aKaOqIh.gif"},
			node3,
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestSplitNodesLink(t *testing.T) {
	t.Run("split_nodes_link", func(t *testing.T) {
		node := nodes.TextNode{
			Type: nodes.Normal,
			Text: "This is text with a link [to boot dev](https://www.boot.dev) and [to youtube](https://www.youtube.com/@bootdotdev)",
			URL:  "",
		}
		result := SplitNodesLink([]nodes.TextNode{node})
		expected := []nodes.TextNode{
			{Type: nodes.Normal, Text: "This is text with a link ", URL: ""},
			{Type: nodes.Link, Text: "to boot dev", URL: "https://www.boot.dev"},
			{Type: nodes.Normal, Text: " and ", URL: ""},
			{Type: nodes.Link, Text: "to youtube", URL: "https://www.youtube.com/@bootdotdev"},
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("split_nodes_link_with_no_links", func(t *testing.T) {
		node := nodes.TextNode{
			Type: nodes.Normal,
			Text: "This is text with no links",
			URL:  "",
		}
		result := SplitNodesLink([]nodes.TextNode{node})
		expected := []nodes.TextNode{node}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("split_nodes_link_with_mixed_nodes", func(t *testing.T) {
		node1 := nodes.TextNode{Type: nodes.Normal, Text: "Normal text", URL: ""}
		node2 := nodes.TextNode{Type: nodes.Normal, Text: "[to boot dev](https://www.boot.dev)", URL: ""}
		node3 := nodes.TextNode{Type: nodes.Normal, Text: "Code `inline` text", URL: ""}
		result := SplitNodesLink([]nodes.TextNode{node1, node2, node3})
		expected := []nodes.TextNode{
			node1,
			{Type: nodes.Link, Text: "to boot dev", URL: "https://www.boot.dev"},
			node3,
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestTextToTextNodes(t *testing.T) {
	text := "This is **text** with an *italic* word and a `code block` and an ![obi wan image](https://i.imgur.com/fJRm4Vk.jpeg) and a [link](https://boot.dev)"
	result := TextToTextNodes(text)
	expected := []nodes.TextNode{
		{Type: nodes.Normal, Text: "This is ", URL: ""},
		{Type: nodes.Bold, Text: "text", URL: ""},
		{Type: nodes.Normal, Text: " with an ", URL: ""},
		{Type: nodes.Italic, Text: "italic", URL: ""},
		{Type: nodes.Normal, Text: " word and a ", URL: ""},
		{Type: nodes.Code, Text: "code block", URL: ""},
		{Type: nodes.Normal, Text: " and an ", URL: ""},
		{Type: nodes.Image, Text: "obi wan image", URL: "https://i.imgur.com/fJRm4Vk.jpeg"},
		{Type: nodes.Normal, Text: " and a ", URL: ""},
		{Type: nodes.Link, Text: "link", URL: "https://boot.dev"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
