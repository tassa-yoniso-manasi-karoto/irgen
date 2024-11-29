
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
	"path"
	"errors"
	"os"
	"net/url"
	"context"
	"slices"
	
	"github.com/PuerkitoBio/goquery"
	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/schollz/progressbar/v3"
	
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/meta"
	"github.com/tassa-yoniso-manasi-karoto/irgen/internal/common"
)

type ExtractorType struct {
	Validator			   *regexp.Regexp
	Name, ContentSelector   string
	Clean				   func(*goquery.Document, string)
	MustSkip				func([]*html.Node) bool
	IMGProcessor		func(context.Context, *meta.Meta, *goquery.Selection)
}

type ThumbnailType struct {
	sort.IntSlice
	Href string
	X, Y int
	Pass bool
}

var (
	extractors = []ExtractorType{wiki}
	SupportedIMGExt = []string{".jpg", ".jpeg", ".png", ".tif", ".tiff", ".gif", ".svg", ".webp", ".avif"}
)

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
		doc.Find("table.sidebar").Remove()
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
		if contains([]string{"Notes", "See also", "External links", "References" , "Citations", "Footnotes", "Bibliography"}, Text(TitleStack[1])) {
			return true
		}
		return false
	},
	IMGProcessor: func(ctx context.Context, m *meta.Meta, n *goquery.Selection) {
		imgs := n.Find("img")
		
		totalImages := 0
		imgs.Each(func(i int, s *goquery.Selection) {
			_, found := s.Parent().Attr("href")
			if !found {
				return
			}
			class, found := s.Parent().Attr("class")
			if !found || class != "mw-file-description" {
				return
			}
			totalImages++
		})

		m.Log.Debug().Int("totalImages", totalImages).Msg("Starting image resolution analysis")

		var URLs, filenames []string
		currentImage := 0

		var bar *progressbar.ProgressBar
		if !m.GUIMode {
			bar = progressbar.NewOptions(totalImages,
				progressbar.OptionEnableColorCodes(true),
				progressbar.OptionSetWidth(20),
				progressbar.OptionSetDescription("[cyan]Analyzing resolutions..."),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "[green]=[reset]",
					SaucerHead:    "[green]>[reset]",
					SaucerPadding: " ",
					BarStart:      "[",
					BarEnd:        "]",
				}),
			)
		}
		
		imgs.Each(func(i int, s *goquery.Selection) {
			href, found := s.Parent().Attr("href")
			if !found {
				return
			}
			class, found := s.Parent().Attr("class")
			if !found || class != "mw-file-description" {
				return
			}

			currentImage++
			progress := float64(currentImage) / float64(totalImages) * 100
			filename, _ := url.QueryUnescape(path.Base(href))

			prettyName := strings.TrimPrefix(filename, "File:")
			if m.GUIMode {
				runtime.EventsEmit(ctx, "download-progress", common.DownloadProgress{
					Current:		currentImage,
					Total:			totalImages,
					Progress:		progress,
					CurrentFile:		prettyName,
					Speed:			"",
					Operation:		"Analyzing resolutions for",
				})
			} else {
				bar.Describe(fmt.Sprintf("[cyan]%s[reset] %s", "Find res. for ", common.StringCapLen(prettyName, 25)))
				bar.Set(currentImage)
			}

			href = wikiPrefForHiRes(m, href)
			filename, _ = url.QueryUnescape(path.Base(href))
			// minimize clutter
			s.RemoveAttr("width")
			s.RemoveAttr("height")
			s.RemoveAttr("srcset")
			s.RemoveAttr("decoding")
			s.RemoveAttr("class")
			s.RemoveAttr("data-file-height")
			s.RemoveAttr("data-file-width")
			// update the src in the <img> with the new file
			s.SetAttr("src", filename)
			// remove wiki file description link
			s.Unwrap()
			path := filepath.Join(m.Config.CollectionMedia, filename)
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				URLs = append(URLs, "https://"+strings.TrimPrefix(href, "//"))
				filenames = append(filenames, filename)
			} else {
				m.Log.Trace().Str("path", path).Msg("Img exist already")
			}
		})

		if !m.GUIMode {
			bar.Finish()
			fmt.Println() // Add newline after progress bar
		}

		m.Log.Trace().Strs("URLs", URLs).Strs("filenames", filenames).Msg("Downloads starting")
		common.DownloadFiles(ctx, m, URLs, filenames)
		
		if m.GUIMode {
			runtime.EventsEmit(ctx, "download-progress", nil)
		}
		m.Log.Trace().Msg("Downloads completed")
	},
}




func (Extractor ExtractorType) TakeImgAlong(ctx context.Context, m *meta.Meta, n *goquery.Selection) {
	if Extractor.Name == "local"  {
		origDir := filepath.Dir(m.Targ)
		files, _ := ioutil.ReadDir(origDir)
		var total int
		for _, file := range files {
			origPath := filepath.Join(origDir, file.Name())
			destPath := filepath.Join(m.Config.CollectionMedia, file.Name())
			_, err := os.Stat(destPath)
			if errors.Is(err, os.ErrNotExist) && slices.Contains(SupportedIMGExt, filepath.Ext(origPath)) {
				origFile, err := os.Open(origPath)
				if err != nil {
					m.Log.Error().Err(err).Str("origPath", origPath).Msg("can't read img to copy")
				}
				destFile, err := os.OpenFile(m.Config.CollectionMedia + file.Name(), os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					m.Log.Error().Err(err).Str("origPath", origPath).Msg("can't access destination where img must be copied")
				}
				// very counterintuitive but io.Copy is: Copy(dst Writer, src Reader)
				_, err = io.Copy(destFile, origFile)
				if err != nil {
					m.Log.Error().Err(err).Str("origPath", origPath).Msg("file copying error")
				}
				origFile.Close()
				destFile.Close()
				total += 1
			}
		}
		m.Log.Info().Msg(fmt.Sprint(total, " images copied."))
	} else {
		Extractor.IMGProcessor(ctx, m, n)
	}
}


func wikiPrefForHiRes(m *meta.Meta, href string) (wanted string) {
	resp, err := http.Get(href)
	if err != nil {
		m.Log.Error().Err(err).Str("href", href).Msg("error during GET request to img")
		return
	}
	if resp.StatusCode != http.StatusOK {
		m.Log.Error().Str("HTTP status code", resp.Status).Msg("")
		return
	}
	file, err := io.ReadAll(resp.Body)
	if err != nil {
		m.Log.Error().Err(err).Str("href", href).Msg("error ready body response")
		return
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	if err != nil {
		m.Log.Error().Err(err).Msg("couldn't prepare the image-containing wikipage for parsing")
		return
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
			wanted = Thumbnail.Href
			m.Log.Trace().Str("wanted", wanted).Str("href", href).Msg("THERE IS resized variant")
		}
	}
	// some low res img don't have any resized variants
	if len(Thumbnails) == 0 {
		doc.Find("a.internal").Each(func(index int, s *goquery.Selection) {
			wanted, _ = s.Attr("href")
			m.Log.Trace().Str("wanted", wanted).Str("href", href).Msg("NO resized variant")
		})
	}
	return
}



func placeholder6zui9876() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}

