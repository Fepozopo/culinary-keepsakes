package blocks

import (
	"reflect"
	"testing"

	"github.com/Fepozopo/culinary-keepsakes/src/nodes"
)

// Unit tests for MarkdownToBlocks
func TestMarkdownToBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "Basic markdown",
			input: "# This is a heading\n\nThis is a paragraph of text. It has some **bold** and *italic* words inside of it.\n\n* This is the first list item in a list block\n* This is a list item\n* This is another list item",
			expected: []string{
				"# This is a heading",
				"This is a paragraph of text. It has some **bold** and *italic* words inside of it.",
				"* This is the first list item in a list block\n* This is a list item\n* This is another list item",
			},
		},
		{
			name:  "Markdown with excessive empty lines",
			input: "# This is a heading\n\n\n\nThis is a paragraph of text. It has some **bold** and *italic* words inside of it.\n\n\n* This is the first list item in a list block\n* This is a list item\n* This is another list item",
			expected: []string{
				"# This is a heading",
				"This is a paragraph of text. It has some **bold** and *italic* words inside of it.",
				"* This is the first list item in a list block\n* This is a list item\n* This is another list item",
			},
		},
		{
			name: "Markdown with leading/trailing whitespace",
			input: `
            # This is a heading

            This is a paragraph of text. It has some **bold** and *italic* words inside of it.

            * This is the first list item in a list block
            * This is a list item
            * This is another list item
        `,
			expected: []string{
				"# This is a heading",
				"This is a paragraph of text. It has some **bold** and *italic* words inside of it.",
				"* This is the first list item in a list block\n* This is a list item\n* This is another list item",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := MarkdownToBlocks(test.input)
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d blocks, got %d", len(test.expected), len(result))
			}
			for i := range result {
				if result[i] != test.expected[i] {
					t.Errorf("Block %d: expected %q, got %q", i, test.expected[i], result[i])
				}
			}
		})
	}
}

func TestBlockToBlockType(t *testing.T) {
	tests := []struct {
		block string
		want  BlockType
	}{
		{
			block: "* This is the first list item in a list block\n* This is a list item\n* This is another list item",
			want:  UnorderedList,
		},
		{
			block: "# This is a heading",
			want:  Heading,
		},
		{
			block: "* This is a list item",
			want:  UnorderedList,
		},
		{
			block: "1. This is a list item\n2. This is another list item",
			want:  OrderedList,
		},
		{
			block: "```This is a code block```",
			want:  Code,
		},
		{
			block: "> This is a quote",
			want:  Quote,
		},
		{
			block: "This is a paragraph of text",
			want:  Paragraph,
		},
		{
			// Invalid list item format
			block: "1. This is a list item\n1. This is another list item",
			want:  Paragraph,
		},
		{
			// Invalid heading format
			block: "#This is a heading",
			want:  Paragraph,
		},
		{
			// Invalid code block format
			block: "```This is a code block",
			want:  Paragraph,
		},
	}

	for _, tt := range tests {
		t.Run(tt.block, func(t *testing.T) {
			got := BlockToBlockType(tt.block)
			if got != tt.want {
				t.Errorf("BlockToBlockType(%q) = %q, want %q", tt.block, got, tt.want)
			}
		})
	}
}

func TestTextToChildren(t *testing.T) {
	result := TextToChildren("This is **bold** and *italic*.")
	expected := []nodes.HTMLNode{
		&nodes.LeafNode{Tag: "", Value: "This is ", Props: nil},
		&nodes.LeafNode{Tag: "b", Value: "bold", Props: nil},
		&nodes.LeafNode{Tag: "", Value: " and ", Props: nil},
		&nodes.LeafNode{Tag: "i", Value: "italic", Props: nil},
		&nodes.LeafNode{Tag: "", Value: ".", Props: nil},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestBlockToHTMLNodeHeading(t *testing.T) {
	block := "# Heading 1"
	result := BlockToHTMLNode(block)
	expected := &nodes.ParentNode{
		Tag: "h1",
		Children: []nodes.HTMLNode{
			&nodes.LeafNode{Tag: "", Value: "Heading 1", Props: nil},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestBlockToHTMLNodeParagraph(t *testing.T) {
	block := "This is a paragraph."
	result := BlockToHTMLNode(block)
	expected := &nodes.ParentNode{
		Tag: "p",
		Children: []nodes.HTMLNode{
			&nodes.LeafNode{Tag: "", Value: "This is a paragraph.", Props: nil},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestBlockToHTMLNodeUnorderedList(t *testing.T) {
	block := "* Item 1\n* Item 2"
	result := BlockToHTMLNode(block)
	expected := &nodes.ParentNode{
		Tag: "ul",
		Children: []nodes.HTMLNode{
			&nodes.ParentNode{
				Tag: "li",
				Children: []nodes.HTMLNode{
					&nodes.LeafNode{Tag: "", Value: "Item 1", Props: nil},
				},
			},
			&nodes.ParentNode{
				Tag: "li",
				Children: []nodes.HTMLNode{
					&nodes.LeafNode{Tag: "", Value: "Item 2", Props: nil},
				},
			},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestMarkdownToHTMLNodeKeepsImagesWithOrderedSteps verifies that step images do not terminate an ordered list.
func TestMarkdownToHTMLNodeKeepsImagesWithOrderedSteps(t *testing.T) {
	markdown := "1. Prepare the peaches.\n\n![Prepared peaches](/images/peaches-01.webp)\n\n2. Roast the peaches.\n\n![Roasted peaches](/images/peaches-02.webp)"

	html, err := MarkdownToHTMLNode(markdown).ToHTML()
	if err != nil {
		t.Fatalf("MarkdownToHTMLNode() returned an error: %v", err)
	}

	want := `<div><ol><li>Prepare the peaches.<p><img alt="Prepared peaches" src="/images/peaches-01.webp" /></p></li><li>Roast the peaches.<p><img alt="Roasted peaches" src="/images/peaches-02.webp" /></p></li></ol></div>`
	if html != want {
		t.Errorf("MarkdownToHTMLNode() = %q, want %q", html, want)
	}
}

// TestMarkdownToHTMLNode converts headings, paragraphs, and unordered lists into HTML nodes.
func TestMarkdownToHTMLNode(t *testing.T) {
	markdown := "# Title\n\nThis is a **bold** paragraph.\n\n* Item 1\n* Item 2"
	result := MarkdownToHTMLNode(markdown)
	expected := &nodes.ParentNode{
		Tag: "div",
		Children: []nodes.HTMLNode{
			&nodes.ParentNode{
				Tag: "h1",
				Children: []nodes.HTMLNode{
					&nodes.LeafNode{Tag: "", Value: "Title", Props: nil},
				},
			},
			&nodes.ParentNode{
				Tag: "p",
				Children: []nodes.HTMLNode{
					&nodes.LeafNode{Tag: "", Value: "This is a ", Props: nil},
					&nodes.LeafNode{Tag: "b", Value: "bold", Props: nil},
					&nodes.LeafNode{Tag: "", Value: " paragraph.", Props: nil},
				},
			},
			&nodes.ParentNode{
				Tag: "ul",
				Children: []nodes.HTMLNode{
					&nodes.ParentNode{
						Tag: "li",
						Children: []nodes.HTMLNode{
							&nodes.LeafNode{Tag: "", Value: "Item 1", Props: nil},
						},
					},
					&nodes.ParentNode{
						Tag: "li",
						Children: []nodes.HTMLNode{
							&nodes.LeafNode{Tag: "", Value: "Item 2", Props: nil},
						},
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
