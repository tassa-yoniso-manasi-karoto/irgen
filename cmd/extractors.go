package cmd

import (
	"fmt"
	"golang.org/x/net/html"
	"regexp"
	"strconv"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"io"
	"bytes"
	"strings"
	"sort"
	"time"
	"path"
	"errors"
	"os"
	"net/url"
	
	"github.com/schollz/progressbar/v3"
	"github.com/PuerkitoBio/goquery"
	"github.com/rs/zerolog/log"
)


type ExtractorType struct {
	Validator               *regexp.Regexp
	Name, ContentSelector   string
	Clean                   func(*goquery.Document, string)
	MustSkip                func([]*html.Node) bool
	IMGProcessor		func(n *goquery.Selection)
}

type ThumbnailType struct {
	sort.IntSlice
	Href string
	X, Y int
	Pass bool
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
	IMGProcessor: func(n *goquery.Selection) {
		imgs := n.Find("img")
		bar := progressbar.DefaultBytes(
			-1,
			"Downloading images...",
		)
		imgs.Each(func(i int, s *goquery.Selection) {
			href, found := s.Parent().Attr("href")
			if !found {
				return
			}
			//pp.Println(RenderNode(s.Nodes[0]))
			href = wikiPrefForHiRes(href)
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
			p := pref.CollectionMedia+filename
			if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
				//log.Debug().Msg("\nDownloading https://"+strings.TrimPrefix(href, "//"))
				DownloadFile(p, "https://"+strings.TrimPrefix(href, "//"), bar)
				currentTime := time.Now().Local()
				_ = os.Chtimes(p, currentTime, currentTime)
			}
			bar.Add(1)
		})
		fmt.Print("\n")
	},
}



func (Extractor ExtractorType) TakeImgAlong(n *goquery.Selection) {
	if Extractor.Name == "local"  {
		files, _ := ioutil.ReadDir(filepath.Dir(*inFile))
		var total int
		for _, file := range files {
			for _, ext := range []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".gif", ".svg", ".webp", ".avif"} {
				fpath := filepath.Dir(*inFile) + string(os.PathSeparator) + file.Name()
				_, err := os.Stat(pref.CollectionMedia + file.Name())
				if ext == filepath.Ext(file.Name()) && errors.Is(err, os.ErrNotExist) {
					from, err := os.Open(fpath)
					check(err)
					to, err := os.OpenFile(pref.CollectionMedia + file.Name(), os.O_CREATE|os.O_WRONLY, 0644)
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
		Extractor.IMGProcessor(n)
	}
}


func wikiPrefForHiRes(href string) string {
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


func DownloadFile(filepath string, url string, bar *progressbar.ProgressBar) error {
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
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	return err
}

