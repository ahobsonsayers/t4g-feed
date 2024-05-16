package t4g

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/foolin/pagser"
)

// numbersRegex is a regex to match a the first consecutive numbers in a link.
var numbersRegex = regexp.MustCompile(`\d+`)

// PagserExtractAttrNumbers is a pagser function that extract the first consecutive numbers form an attribute
// Based on: https://github.com/foolin/pagser/blob/v0.1.6/builtin_functions.go#L43
func PagserAttrNumbers(node *goquery.Selection, args ...string) (any, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("attrNumbers(name) must have one argument")
	}
	attrValue, _ := node.Attr(args[0])
	return strconv.Atoi(numbersRegex.FindString(attrValue))
}

func NewHTMLParser() *pagser.Pagser {
	parser := pagser.New()
	parser.RegisterFunc("attrNumbers", PagserAttrNumbers)
	return parser
}
