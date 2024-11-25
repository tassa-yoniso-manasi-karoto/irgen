
package core

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
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
)


type ExtractorType struct {
	Validator               *regexp.Regexp
	Name, ContentSelector   string
	Clean                   func(*goquery.Document, string)
	MustSkip                func([]*html.Node) bool
	IMGProcessor		func(*meta.Meta, *goquery.Selection)
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
	IMGProcessor: func(m *meta.Meta, n *goquery.Selection) {
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
			href = wikiPrefForHiRes(m, href)
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
			p := m.Config.CollectionMedia+filename
			if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
				//log.Debug().Msg("\nDownloading https://"+strings.TrimPrefix(href, "//"))
				DownloadFile(p, "https://"+strings.TrimPrefix(href, "//"), bar)
				currentTime := time.Now().Local()
				_ = os.Chtimes(p, currentTime, currentTime)
			}
			// TODO rework due to GUI
			bar.Add(1)
		})
		fmt.Print("\n")
	},
}



func (Extractor ExtractorType) TakeImgAlong(m *meta.Meta, n *goquery.Selection) {
	if Extractor.Name == "local"  {
		files, _ := ioutil.ReadDir(filepath.Dir(inFile))
		var total int
		for _, file := range files {
			for _, ext := range []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".gif", ".svg", ".webp", ".avif"} {
				fpath := filepath.Dir(inFile) + string(os.PathSeparator) + file.Name()
				_, err := os.Stat(m.Config.CollectionMedia + file.Name())
				if ext == filepath.Ext(file.Name()) && errors.Is(err, os.ErrNotExist) {
					from, err := os.Open(fpath)
					if err != nil {
						m.Log.Error().Err(err).Str("fpath", fpath).Msg("can't read img to copy")
					}
					to, err := os.OpenFile(m.Config.CollectionMedia + file.Name(), os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						m.Log.Error().Err(err).Str("fpath", fpath).Msg("can't access destination where img must be copied")
					}
					_, err = io.Copy(to, from)
					if err != nil {
						m.Log.Error().Err(err).Str("fpath", fpath).Msg("file copying error")
					}
					from.Close()
					to.Close()
					total += 1
				}
			}
		}
		log.Info().Msg(fmt.Sprint(total, " images copied."))
	} else {
		Extractor.IMGProcessor(m, n)
	}
}


func wikiPrefForHiRes(m *meta.Meta, href string) string {
	resp, err := http.Get(href)
	if err != nil {
		m.Log.Error().Err(err).Str("href", href).Msg("error during GET request to img")
	}
	if resp.StatusCode != http.StatusOK {
		log.Error().Str("HTTP status code", resp.Status).Msg("")
	}
	file, err := io.ReadAll(resp.Body)
	if err != nil {
		m.Log.Error().Err(err).Str("href", href).Msg("error ready body response")
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	if err != nil {
		m.Log.Fatal().Err(err).Msg("couldn't prepare the image-containing wikipage for parsing")
	}
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
		if m.Config.ResXMax < x {
			Thumbnail.Pass = false
		}
		if m.Config.ResYMax < y {
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

