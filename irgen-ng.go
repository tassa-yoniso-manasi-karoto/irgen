package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"
	"path"
	"bytes"
	"os"
	"path/filepath"
	"encoding/csv"
	"flag"
	"net/http"
	"net/url"
	"errors"
	"strconv"
	"sort"
	
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"
	"github.com/schollz/progressbar/v3"
	"github.com/rs/zerolog"
)

var (
	CurrentDir string
	ContentSelector = "body"
	inFile *string
	WantedTitleLen = 3
	reCleanHTML = regexp.MustCompile(`^\s*(.*?)\s*$`)
	Extractor ExtractorType
	Article ArticleType
	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
)

type ArticleType struct {
	Name, Lang string
}

type NoteType struct {
	QNode *goquery.Selection
	ID, Title, Txt, Src string
	Tags []string
	hasContent bool
}

func main() {
	log.Info().Msg("Started")
	inFile = flag.String("i", CurrentDir+"/article.html", "file path or URL of an HTML article\n")
	flag.Parse()
	Article.Name = filepath.Base((*inFile)[:len(*inFile) - len(filepath.Ext(*inFile))])
	var outFile string
	if pref.DestDir == "" {
		outFile = strings.TrimSuffix(*inFile, filepath.Ext(*inFile)) + ".txt"
	} else {
		outFile = pref.DestDir + strings.TrimSuffix(filepath.Base(*inFile), filepath.Ext(*inFile)) + ".txt"
	}
	log.Debug().
		Bool("AbsPath?", filepath.IsAbs(*inFile)).
		Bool("canStat?", canStat(*inFile)).
		Str("path||url", *inFile).
		Msg("init")	
	var file []byte
	var err error
	if filepath.IsAbs(*inFile){
		if !canStat(*inFile) {
			log.Error().Msg("File not found: " + *inFile)
			os.Exit(1)
		}
		Extractor = local
		file, err = os.ReadFile(*inFile)
		check(err)
	} else {
		for _, extractor := range extractors {
			if extractor.Validator.MatchString(*inFile) {
				Extractor = extractor
				if extractor.Validator.NumSubexp() > 0 {
					Article.Lang = extractor.Validator.FindStringSubmatch(*inFile)[1]
				}
				if extractor.Validator.NumSubexp() > 1 {
					Article.Name, _ = url.QueryUnescape(extractor.Validator.FindStringSubmatch(*inFile)[2])
					Article.Name = strings.ReplaceAll(Article.Name, "_", " ")
				}
				outFile = fmt.Sprint(pref.DestDir, Extractor.Name, "–",Article.Name, ".txt")
				break
			}
		}
	}
	log.Info().
		Str("source", Extractor.Name).
		Str("lang", Article.Lang).
		Str("Out", outFile).
		Msg("init")
	if Extractor.Name != "local" {
		resp, err := http.Get(*inFile)
		check(err)
		if resp.StatusCode != http.StatusOK {
			log.Error().Str("Received response status", resp.Status).Msg("HTTP")
		} else {
			log.Info().Str("Received response status", resp.Status).Msg("HTTP")
		}
		file, err = io.ReadAll(resp.Body)
		check(err)
	}	
	launch := time.Now()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	check(err)
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
	Extractor.TakeImgAlong(n)
	// this returns InnnerHTML
	content, err := n.Html()
	check(err)
	content = "<cutpattern>" + Cut(content) + "</cutpattern>"
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(content))
	check(err)
	Preprocess(doc)
	var Notes []NoteType
	doc.Find("cutpattern").Each(func(i int, s *goquery.Selection) {
		node := s.Nodes[0]
		if Text(node) == "" {
			return
		}
		loc, ok := LocRegister[node]
		if !ok {
			fmt.Println("LOC NOT FOUND for", node.Data, stringCapLen(InnerHTML(node), 200))
		}
		TitleStack := loc.Stack()
		if len(TitleStack) > 1 && Extractor.MustSkip(TitleStack) {
			return
		}
		Note := NoteType {
			QNode: s,
			ID: fmt.Sprintf("%s–%s %s", Article.Name, loc.miniStr(), fmtTl(TitleStack, -1)),
			Title: fmtTl(TitleStack, pref.LenStack),
			Txt: InnerHTML(s.Nodes[0]),
		}
		Note.Src = Note.MkCxt(loc, TitleStack)
		// keep this after MkCxt to be able to ez check for duplicate img
		Note.Txt = gohtml.Format(Note.Txt)
		Notes = append(Notes, Note)
	})
	csvout, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	check(err)
	writer := csv.NewWriter(csvout)
	writer.Comma = '\t'
	defer csvout.Close()
	defer writer.Flush()
	for _, Note := range Notes {
		_ = writer.Write([]string{Note.ID, Note.Title, Note.Txt, Note.Src})
	}
	log	.Info().Msg(fmt.Sprint(len(Notes), " Notes in total"))
	elapsed := time.Since(launch)
	log.Info().Msg(fmt.Sprintf("\x1b[1;37;45mTOOK %s\x1b[0m\n", elapsed))
}


func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func (Extractor ExtractorType) TakeImgAlong(n *goquery.Selection) {
	if Extractor.Name == "local"  {
		files, _ := ioutil.ReadDir(filepath.Dir(*inFile))
		var total int
		for _, file := range files {
			for _, ext := range []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".gif", ".svg", ".webp", ".avif"} {
				fpath := filepath.Dir(*inFile) + "/" + file.Name()
				_, err := os.Stat(pref.Collection + file.Name())
				if ext == filepath.Ext(file.Name()) && errors.Is(err, os.ErrNotExist) {
					from, err := os.Open(fpath)
					check(err)
					to, err := os.OpenFile(pref.Collection + file.Name(), os.O_CREATE|os.O_WRONLY, 0644)
					check(err)
					_, err = io.Copy(to, from)
					check(err)
					from.Close()
					to.Close()
					total += 1
				}
			}
		}
		log.Info().Msg(fmt.Sprint(total, " images copied."))
	} else {
		imgs := n.Find("img")
		bar := progressbar.NewOptions(imgs.Length(), progressbar.OptionSetDescription(fmt.Sprint("Downloading ",imgs.Length()," images...")))
		imgs.Each(func(i int, s *goquery.Selection) {
			/*src, found := s.Attr("src")
			if !found {
				log.Warn().Msg("Img tag without src attribute")
			}*/
			href, found := s.Parent().Attr("href")
			if !found {
				return
			}
			//pp.Println(RenderNode(s.Nodes[0]))
			href = Extractor.PrefForHiRes(href)
			//log.Debug().Msg("\n"+src+" →→→→→→→ "+path.Base(href))
			s.RemoveAttr("width")
			s.RemoveAttr("height")
			s.RemoveAttr("srcset")
			s.RemoveAttr("decoding")
			s.RemoveAttr("class")
			s.RemoveAttr("data-file-height")
			s.RemoveAttr("data-file-width")
			filename, _ := url.QueryUnescape(path.Base(href))
			s.SetAttr("src", filename)
			s.Unwrap()
			//title := a.Attr("title")
			//if strings.Contains(strings.ToLower(title), "map")
			p := pref.Collection+filename
			if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
				log.Debug().Msg("\nDownloading https://"+strings.TrimPrefix(href, "//"))
				DownloadFile(p, "https://"+strings.TrimPrefix(href, "//"))
				currentTime := time.Now().Local()
				_ = os.Chtimes(p, currentTime, currentTime)								
			}
			bar.Add(1)
		})
		fmt.Print("\n")
	}
}

type ThumbnailType struct {
	sort.IntSlice
	Href string
	X, Y int
	Pass bool
}

func (Extractor ExtractorType) PrefForHiRes(href string) string {
	resp, err := http.Get(href)
	check(err)
	if resp.StatusCode != http.StatusOK {
		log.Error().Str("IMG: Received response status", resp.Status).Msg("HTTP")
	}
	file, err := io.ReadAll(resp.Body)
	check(err)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	check(err)
	var Thumbnails []ThumbnailType
	var xs []int
	doc.Find("a.mw-thumbnail-link").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		Thumbnail := ThumbnailType{Href: href}
		sz := strings.ReplaceAll(strings.TrimSuffix(s.Text(), " pixels"), ",", "")
		xstr, ystr, _ := strings.Cut(sz, " × ")
		x, _ := strconv.Atoi(xstr)
		y, _ := strconv.Atoi(ystr)
		Thumbnail.Pass = true
		if pref.ResXMax < x {
			Thumbnail.Pass = false
		}
		if pref.ResYMax < y {
			Thumbnail.Pass = false
		}
		Thumbnail.X = x
		Thumbnail.Y = y
		xs = append(xs, x)
		Thumbnails = append(Thumbnails, Thumbnail)
	})
	Rating := RatingType{sort.IntSlice(xs), Thumbnails}
	sort.Stable(Rating)
	for _, Thumbnail := range Thumbnails {
		if Thumbnail.Pass {
			href = Thumbnail.Href
		}
	}
	return href
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
