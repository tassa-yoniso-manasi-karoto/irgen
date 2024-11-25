
package core

import (
	"fmt"
	"golang.org/x/net/html"
	
	"github.com/PuerkitoBio/goquery"
)

/*
this file was dedicated to try to guess find a user-predefined sequence that is part of the context of an image/table ie. captions or
a title but yet not properly defined through HTML tags. 
It needs instruction that are tailored made to how figures appear in the document and therefore was full of hardcoded stuff.
Need rewriting from scratch
*/


type ObjectT struct {
	Type, Origin string
	PosRefNode, Scope int
	Nodes []*html.Node
	Selec *goquery.Selection
}


/*
UsingRef			<table> (NODE) || <div> title (NODE)
FromSuperior			<table> (NODE) || <img> (NODE)
InlineCaptionInspector		<div> caption (NODE)
InIMGCaptionInspector		<img> (NODE)
*/

func (Object ObjectT) Fmt() (str string) {
	str = fmt.Sprintf("<section class=\"capillary\" origin=\"%s\" scope=\"%d\">", Object.Origin, Object.Scope)
	tmp, _ := goquery.OuterHtml(Object.Selec)
	str += tmp
	str += "</section>\n\n"
	return
}






