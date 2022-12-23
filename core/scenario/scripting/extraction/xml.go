package extraction

import (
	"bytes"
	"fmt"

	"github.com/antchfx/xmlquery"
)

type XmlExtractor struct {
}

func (xe XmlExtractor) extractFromByteSlice(source []byte, xPath string) (interface{}, error) {
	reader := bytes.NewBuffer(source)
	rootNode, err := xmlquery.Parse(reader)
	if err != nil {
		return nil, err
	}

	// returns the first matched element
	foundNode, err := xmlquery.Query(rootNode, xPath)
	if err != nil {
		return nil, fmt.Errorf("no match")
	}

	return []byte(foundNode.InnerText()), nil
}
