package main

import (
	"fmt"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	
	"github.com/PuerkitoBio/goquery"
)


type ExtractorType struct {
	Validator               *regexp.Regexp
	Name, ContentSelector   string
	Clean                   func(*goquery.Document, string)
	MustSkip                func([]*html.Node) bool
}


var extractors = []ExtractorType{wiki}

var local = ExtractorType{
	Name: "local",
	ContentSelector: "body",
	Clean: func(doc *goquery.Document, lang string) {},
	MustSkip: func(TitleStack []*html.Node) bool {return false},
}




var wiki = ExtractorType{
	Name: "Wikipedia",
	Validator: regexp.MustCompile(`https?://([a-z]+).wikipedia.org/wiki/([^?]+)`), // TODO https://en.wikipedia.org/w/index.php?title=xxxxxxxxxxx
	ContentSelector: ".mw-content-ltr",
	Clean: func(doc *goquery.Document, lang string) {
		doc.Find(".sistersitebox, #toc, table.navbox-inner").Remove()
		doc.Find("table.metadata, span.mw-editsection").Remove()
		doc.Find("table.mw-collapsible").Children().First().Unwrap()		
		doc.Find("h1").Remove()
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, found := s.Attr("href")
			if found {
				s.SetAttr("href", "https://"+lang + ".wikipedia.org" +href)
			}
		})
		doc.Find("h2,h3,h4,h5,h6").Each(func(i int, s *goquery.Selection) {
			h := s.Nodes[0]
			x, _ := strconv.Atoi(h.Data[1:])
			s.Nodes[0].Data = fmt.Sprint("h", x-1)
		})
	},
	MustSkip: func(TitleStack []*html.Node) bool {
		if contains([]string{"Notes", "See also", "External links", "References" , "Citations", "Bibliography"}, Text(TitleStack[1])) {
			return true
		}
		return false
	},
}
