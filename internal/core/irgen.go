package core

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"regexp"
	"strings"
	"time"
	"path/filepath"
	"bytes"
	"os"
	"encoding/csv"
	"net/http"
	"net/url"
	"sort"
	"context"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
)

var (
	CurrentDir string
	ContentSelector = "body"
	WantedTitleLen = 3
	inFile string
	reCleanHTML = regexp.MustCompile(`^\s*(.*?)\s*$`)
	Extractor ExtractorType
	Article ArticleType
)

type ArticleType struct {
	Name, Lang string
}

// WARNING: Don't confuse Note.Context (image enriched field of the note-to-be) and Go's context!
type NoteType struct {
	QNode *goquery.Selection
	ID, Title, Txt, Context string
	Tags []string
	hasContent bool
}


/* TODO
add waiting/process bar while parsing image-wikipages
fix bug in wikiPrefForHiRes providing non-img URLs to the img downloader
minimize m.Log.Fatal() usage bc it crashes the GUI
FIX CORE: "1 Notes in total"????


split Execute func below
*/


func Execute(ctx context.Context, m *meta.Meta) {
	m.LogConfig("config state at execution")
	userGivenPath := m.Targ
	m.Log.Debug().Msg("Execution started")
	if m.Config.CollectionMedia == "" {
		m.Log.Error().Msg("Images can't be automatically imported because the path to collection.media has not been provided.")
	}
	inFile = userGivenPath
	Article.Name = filepath.Base((userGivenPath)[:len(userGivenPath) - len(filepath.Ext(userGivenPath))])
	var outFile string
	if m.Config.DestDir == "" {
		outFile = strings.TrimSuffix(userGivenPath, filepath.Ext(userGivenPath)) + ".txt"
	} else {
		outFile = m.Config.DestDir + strings.TrimSuffix(filepath.Base(userGivenPath), filepath.Ext(userGivenPath)) + ".txt"
	}
	m.Log.Debug().
		Bool("AbsPath?", filepath.IsAbs(userGivenPath)).
		Bool("canStat?", canStat(userGivenPath)).
		Str("path||url", userGivenPath).
		Msg("init")	
	var file []byte
	var err error
	if filepath.IsAbs(userGivenPath) {
		if !canStat(userGivenPath) {
			m.Log.Fatal().Msg("No input file specified or default file location unaccessible: " + userGivenPath)
		}
		Extractor = local
		file, err = os.ReadFile(userGivenPath)
		if err != nil {
			m.Log.Fatal().Err(err).Msg("can stat but not read specified input file, check permissions")
		}
	} else {
		for _, extractor := range extractors {
			if extractor.Validator.MatchString(userGivenPath) {
				Extractor = extractor
				if extractor.Validator.NumSubexp() > 0 {
					Article.Lang = extractor.Validator.FindStringSubmatch(userGivenPath)[1]
				}
				if extractor.Validator.NumSubexp() > 1 {
					Article.Name, _ = url.QueryUnescape(extractor.Validator.FindStringSubmatch(userGivenPath)[2])
					Article.Name = strings.ReplaceAll(Article.Name, "_", " ")
				}
				outFile = fmt.Sprint(m.Config.DestDir, Extractor.Name, "–",Article.Name, ".txt")
				break
			}
		}
	}
	m.Log.Info().
		Str("source", Extractor.Name).
		Str("lang", Article.Lang).
		Str("Out", outFile).
		Msg("init")
	if Extractor.Name != "local" {
		resp, err := http.Get(userGivenPath)
		if resp.StatusCode != http.StatusOK {
			m.Log.Error().Str("Received response status", resp.Status).Msg("HTTP")
		} else {
			m.Log.Info().Str("Received response status", resp.Status).Msg("HTTP")
		}
		if err != nil {
			m.Log.Fatal().Err(err).Msg("couldn't access URL")
		}
		file, err = io.ReadAll(resp.Body)
		if err != nil {
			m.Log.Fatal().Err(err).Msg("reading retrieved data failed")
		}
	}
	launch := time.Now()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	if err != nil {
		m.Log.Fatal().Err(err).Msg("couldn't prepare the document for parsing")
	}
	Extractor.Clean(doc, Article.Lang)
	n := doc.Find(Extractor.ContentSelector)
	// drag the headings up until they are direct children of the content-containing tag
	// this make things safe to monkey-patch with Cut()
	n.Children().EachWithBreak(func(i int, s *goquery.Selection) bool {
		h := s.Find("h1,h2,h3,h4,h5,h6")
		if h.Nodes == nil {
			return false
		}
		s.ReplaceWithSelection(h)
		return true
	})
	Extractor.TakeImgAlong(ctx, m, n)
	// this returns InnerHTML
	content, err := n.Html()
	if err != nil {
		m.Log.Fatal().Err(err).Msg("couldn't access HTML content of file")
	}
	content = "<cutpattern>" + Cut(content) + "</cutpattern>"
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		m.Log.Fatal().Err(err).Msg("couldn't prepare the document for 2nd parsing")
	}
	Preprocess(m, doc)
	var Notes []NoteType
	doc.Find("cutpattern").Each(func(i int, s *goquery.Selection) {
		node := s.Nodes[0]
		if Text(node) == "" {
			return
		}
		loc, ok := LocRegister[node]
		if !ok {
			m.Log.Error().
				Str("sample", stringCapLen(InnerHTML(node), 200)).
				Msg("loc not found for node " + node.Data)
		}
		TitleStack := loc.Stack()
		if len(TitleStack) > 1 && Extractor.MustSkip(TitleStack) {
			return
		}
		Note := NoteType {
			QNode: s,
			ID: fmt.Sprintf("%s_%s %s", Article.Name, loc.miniStr(), fmtTl(TitleStack, -1)),
			Title: fmtTl(TitleStack, m.Config.MaxTitles),
			Txt: InnerHTML(s.Nodes[0]),
		}
		Note.Context = Note.MkCxt(m, loc, TitleStack)
		// keep this after MkCxt to be able to ez check for duplicate img
		Note.Txt = gohtml.Format(Note.Txt)
		Notes = append(Notes, Note)
	})
	csvout, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		m.Log.Fatal().Err(err).Msg("couldn't access output CSV file for writing")
	}
	writer := csv.NewWriter(csvout)
	writer.Comma = '\t'
	defer csvout.Close()
	defer writer.Flush()
	for _, Note := range Notes {
		_ = writer.Write([]string{Note.ID, Note.Title, Note.Txt, Note.Context})
	}
	m.Log.Info().Int("total notes", len(Notes)).Msg("")
	elapsed := time.Since(launch)
	m.Log.Info().Msgf("Done in %s", elapsed)
}



type RatingType struct {
	sort.IntSlice
	idx []ThumbnailType
}
func (Rating RatingType) Swap(i, j int) {
	Rating.idx[i], Rating.idx[j] = Rating.idx[j], Rating.idx[i]
	Rating.IntSlice.Swap(i, j)
}

func fmtTl(TitleStack []*html.Node, max int) (s string) {
	added := 0
	for _, n := range(TitleStack[1:]) {
		if max < 0 {
			s = "❱"+Text(n)+s
		} else if added < max {
			s = "<span class=heading>"+Text(n)+"</span><span class=del></span>"+s	
			added += 1		
		}
	}
	s = strings.TrimPrefix(s, "❱")
	if s == "" {
		s = "<span class=heading>"+Article.Name+"</span>"
	}
	return
}


func Text(n *html.Node) string {
	s := goquery.Selection{Nodes: []*html.Node{n}}
	return reCleanHTML.ReplaceAllString(s.Text(), `$1`)
	 
}

var reInnerHTML = regexp.MustCompile(`(?s)^<.*?>(.*)</.*?>$`)
func InnerHTML(n *html.Node) (s string) {
	s = RenderNode(n)
	s = reInnerHTML.ReplaceAllString(s, `$1`)
	return
}


func RenderNode(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

var reHeading = regexp.MustCompile("(?si)<h[0-9]+[^<]*?>(.*?)</h[0-9]+>")
func Cut(str string) string {
	var done []string
	for _, h := range reHeading.FindAllStringSubmatch(str, -1) {
		outer := h[0]
		inner := h[1]
		// ensure only actual html pairs of tags get found
		if reHeading.MatchString(inner) {
			continue
		}
		// string of individual heading may occasionally be the same
		// throughout the document, make sure to not ReplaceAll multiple time
		if contains(done, outer) {
			continue
		}
		str = strings.ReplaceAll(str, outer, "</cutpattern>" + outer + "<cutpattern>")
		done = append(done, outer)
	}
	return str
}

func contains[T comparable](arr []T, i T) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}

func stringCapLen(s string, max int) string{
	trimmed := false
	for len(s) > max {
		s = s[:len(s)-1]
		trimmed = true
	}
	if trimmed {
		s += "…"
	}
	return s
}


func canStat(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func placeholder() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
