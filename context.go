
package main

import (
	"fmt"
	"strconv"
	"strings"
	//"os"
	"golang.org/x/net/html"
	
	"github.com/PuerkitoBio/goquery"
	"github.com/k0kubun/pp"
	"github.com/gookit/color"
	"github.com/rs/zerolog"
)

var (
	mapfunc = map[string]Capillary {
		"UsingRef": NoteType.UsingRef,
		"FromSuperior": NoteType.FromSuperior,
		"FromSuperiorAndDescendants": NoteType.FromSuperiorAndDescendants,
		//"CaptionInspector": NoteType.CaptionInspector,
	}
	isReusable = map[string]bool {
		"UsingRef": false,
		"FromSuperior": true,
		"FromSuperiorAndDescendants": true,
	}
)

type Capillary func(NoteType, []*html.Node, []string, string, int) (ObjectSlice []ObjectT)


// TODO func MoveTxtAddendumToSrc(n *html.Node, TXT string) (string, SRC string){


func (Note NoteType) MkCxt(loc Location, tStack []*html.Node/*, tRefStack []string*/) (src string) {
// in some books headings may contain direct reference to a pic / table,
// tRefStack should contain these from Preprocess but the corresponding Capillary hasn't been rewritten atm
	var tRefStack []string 
	// ignore card(s) not preceed by a heading
	if loc.IsEmpty() {
		return
	}
	for num, fstr := range(pref.Fn) {
		for _, obj := range Note.ObjProvider(num, fstr, loc, tStack, tRefStack) {
			main, _ := goquery.OuterHtml(obj.Selec)
			if !strings.Contains(Note.Txt, main) && !strings.Contains(src, main) {
				src += obj.Fmt()
			}
		}
	}
	return
}


// num = func number (position) in Fn
func (Note NoteType) ObjProvider(num int, fstr string, loc Location, tStack []*html.Node, tRefStack []string) (ObjectSlice []ObjectT) {
	if DataRegister[loc].ShrdObjectSlices[num] == nil {
		// load the func of type "Capillary" and its scope of
		// execution (target heading level)
		f, ok := mapfunc[fstr]; if ok == false {
			panic(fmt.Sprintf("Unknown Capillary: \"%s\"\n", fstr))
		}
		scope := pref.FnScope[num]
		
		// get object from that func
		ObjectSlice = f(Note, tStack, tRefStack, fstr, scope)
		
		// cache ObjectSlice in case redesired later
		if isReusable[fstr] {
			DataRegister[loc].ShrdObjectSlices[num] = ObjectSlice
		}
	} else {
		// or get object from cache
		ObjectSlice = DataRegister[loc].ShrdObjectSlices[num] 
	}
	return ObjectSlice
}


// TODO Rewrite + if scope=0 search whole doc
//var reRef = regexp.MustCompile(`(?m)[^\s]+?.*?((?:Abb|Tab)\.\s*?\d*?\.\d*)`)
func (Note NoteType) UsingRef(tStack []*html.Node, tRefStack []string, fstr string, scope int) (ObjectSlice []ObjectT) {
	return
}

// share the img in the higher section level between the to-be-created notes of that section level
func (Note NoteType) FromSuperior(tStack []*html.Node, _ []string, fstr string, scope int) (ObjectSlice []ObjectT) {
	return Note.superior(0, tStack, fstr, scope)
}

func (Note NoteType) FromSuperiorAndDescendants(tStack []*html.Node, _ []string, fstr string, scope int) (ObjectSlice []ObjectT) {
	return Note.superior(1, tStack, fstr, scope)
}

func (Note NoteType) superior(includeDescendants int, tStack []*html.Node, fstr string, scope int) (ObjectSlice []ObjectT) {
	logger := zerolog.Nop() //zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	logger.Debug().
		Str("fstr", fstr).
		Int("scope", scope).
		Msg("superior")
	//printStack(logger, tStack)
	// Only run if a suitable heading is available at this loc
	// 1st offset = len() induced, 2nd offset for the dummy at idx 0
	idxMax := len(tStack)-1-1
	logger.Debug().
		Int("len", len(tStack)).
		Int("scope", scope).
		Int("idxMax", idxMax).
		Msg("superior")
	var h *html.Node
	var x int
	var s *goquery.Selection
	if idxMax < scope {
		//logger.Debug().Int("maxed out to", idxMax)
		//scope = idxMax
		logger.Debug().
			Int("idxMax", idxMax).
			Int("scope", scope).
			Msg("Out-of-bounds scope constrained to document root")
		s = Note.QNode.Parent().Children()
		if includeDescendants == 1 {
			// because 0 is superior in importance to the int of h1
			x = 0
		} else {
			x = 1 // FIXME
		}
	} else {	
		if idxMax-1 < 0 {
			logger.Warn().Msg("Returned")
			return
		}
		// this is the type of tag at which the parser must stop
		// using plus sign because tStack is RELATIVE, the lower the idx, the closer the heading is to the cutpattern
		h = tStack[scope+includeDescendants]
		x, _ = strconv.Atoi(h.Data[1:])
		s = &goquery.Selection{Nodes: []*html.Node{h}}
		s = s.NextAll()
	}
	/*logger.Debug("query:",
		"searching from", h.Data,
		"stopping if finding a h", x,
	)*/
	// CAVEAT: NEXTALL COMBINED FIND(*) DOES NOT ITERATE OVER THE SIBLING NODE, ONLY OVER ITS DESCENDANTS
	s.EachWithBreak(func(i int, selec *goquery.Selection) bool {
		node := selec.Nodes[0]
		if IsEqualOrMoreImportantHeading(node, x) {
			return false
		}		
		selec.Find("*").Each(func(i int, selec *goquery.Selection) {
			node := selec.Nodes[0]
			if node.Data == "img" || node.Data == "table" {
				if p := selec.ParentsFiltered("figure, div.tmulti"); len(p.Nodes) != 0 {
					selec = p
				}
				ObjectSlice = append(ObjectSlice, ObjectT{
					Type: node.Data,
					Origin: fstr,
					Scope: scope,
					Selec: selec,
				})
			}
		})
		return true
	})
	return
}

/*func printStack(logger *slog.Logger, tStack []*html.Node) {
	logger.Debug("TSTACK=")
	for i := 0; i<8; i++ {
		if len(tStack) > i && tStack[i] != nil {
			logger.Debug(fmt.Sprint("\t",i," ")+Text(tStack[i]))
		} else {
			logger.Debug("\tnil")
		}
	}
}
*/

func placeholder2() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}


func IsEqualOrMoreImportantHeading(n *html.Node, x int) (b bool) {
	y, err := strconv.Atoi(n.Data[1:])
	// if the number of the heading user-provided is 0, treat it as
	// the root of the document (-> document-wide search)
	if (err == nil && n.Data[:1] == "h" && x != 0 && x >= y) {
		return true
	}
	return
}