package mcf

import (
	"encoding/xml"
	"time"
)

type RootXML struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Item    []Item   `xml:"item"`
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	PublishDate Date     `xml:"pubDate"`
}

type Date struct {
	time.Time
}

func (d *Date) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var v string
	err := decoder.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	parse, err := time.Parse(time.RFC1123, v)
	if err != nil {
		return err
	}
	*d = Date{parse}
	return nil
}
