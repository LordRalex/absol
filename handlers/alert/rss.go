package alert

import (
	"encoding/xml"
	"strings"
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
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        RssLink  `xml:"link"`
	Details     string   `xml:"-"`
	Id          string   `xml:"-"`
}

type Date struct {
	time.Time
}

type RssLink struct {
	string
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

func (d *RssLink) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var v string
	err := decoder.DecodeElement(&v, &start)
	if err != nil {
		return err
	}

	//HACK: Our elmah's don't really work with links...
	v = strings.Replace(v, "rss/detail", "json", 1)

	*d = RssLink{v}
	return nil
}

func (i *Item) ParseId() {
	v := i.Link.string
	id := v[len(v) - 64:]
	i.Id = id
}
