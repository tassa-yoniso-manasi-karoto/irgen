
package cmd

import (
	"regexp"
	"strconv"
	"fmt"
	"strings"
	"golang.org/x/net/html"
	
	"github.com/PuerkitoBio/goquery"
)


// The following types along with the couple first var declaration 
// thereafter should help clarify the model of cache in use.

 
// loc of type Location describe the location of a cutpattern in a doc,
// in relationship with its burying among heading tags (<h1>, <h2>...)
type Location [7]int
/*
For instance a loc = [7]int{0, 5, 2, 1, 5, 3, 3} translates as:
After the 5th <h1> of the doc, 2th <h2> of this <h1>, 1th <h3> of this
<h2>...etc.
The first element, located at [0], is a purposeless dummy, that
allows the value of the headings levels (e.g. <h1>) to be aligned 
with their respectives index values in the loc array (e.g. loc[1]).
when it comments-cuts are implemented back several card can be created from 
the same loc and idx 0 will serve the purpose of keeping track of the nbr of
the current of card.
*/

type LocationData struct {
	Heading			*html.Node
	ShrdObjectSlices	[][]ObjectT
}

var (
// CUTPATTERN-NODE TO [7]int LOCATION
	LocRegister = make(map[*html.Node]Location)	
// [7]int LOCATION TO 	(1) the ACTUAL POINTER OF THE NODE OF THE HEADING AND
//                      (2) an array of results from that reflects the array of desired MkCxt-function the user
	DataRegister = make(map[Location]LocationData)
)


func Preprocess(doc *goquery.Document) {
	var (
		currentLoc = new(Location)
		reHeading = regexp.MustCompile(`h[0-6]`)
	)
	body := doc.Find("body")
	d := 0
	body.Children().Each(func(i int, s *goquery.Selection) {
		n := s.Nodes[0]
		hx := reHeading.MatchString(n.Data)
		if hx {
			// currentLoc[0] provides the num of the cutpattern
			// among those that shares the same heading hierachical
			// position. Reinit it when a heading is found.
			currentLoc[0] = 0
			// update loc
			x, _ := strconv.Atoi(n.Data[1:])
			currentLoc[x] += 1 
			
			// reset the count of sub heading if needed
			for i := range(currentLoc) {
				if i > x {
					currentLoc[i] = 0
				}
			}
			// then update its data and link them in the DataRegister for later
			//! be aware ShrdObjects is used only internally and has no dummy at 0
			ObjectSlices := make([][]ObjectT, len(pref.Fn))
			DataRegister[*currentLoc] = LocationData{n, ObjectSlices}
		} else if n.Data == "cutpattern" {
			LocRegister[n] = *currentLoc
			currentLoc[0] += 1 
		} else {
			fmt.Printf("CACHE: A \"%s\" tag was encountered during preprocessing.\n", n.Data)
		}
		d += 1
	})
}


// tStack[0] is dummy, tStack[1] is the _first_ heading node preceding n (e.g. h6),
// tStack[2] is the first heading node preceding tStack[1] with a higher section level (e.g. h5) ...etc
func (loc Location) Stack() []*html.Node {
	var (
		locData LocationData
		found bool
		t *html.Node
		
		tStack = []*html.Node{nil}
	)
	
	c := make(chan Location)
	go GetParentLocs(c, loc)

	for parentHeadingLoc := range(c) {
		locData, found = DataRegister[parentHeadingLoc]
		if found { 
			t = locData.Heading
			tStack = append(tStack, t)
		}
	}
	return tStack
}


// Truncating loc by x (= replacing the x last values of loc array by 0) 
// will provide the loc (location) of the xth parent.
func GetParentLocs(c chan Location, loc Location) {
	for i := range(loc) {
		// start from the end of the slice ("-1" compensate the dummy at loc[0])
		currentLvl := len(loc)-i-1
		// process the truncated loc only if it is filled w/ localisation info
		if loc[currentLvl] != 0 {
			c <- loc
			loc[currentLvl] = 0
		}
		if loc.IsEmpty() {
			break
		}
	}
	close(c)
}




func (loc Location) miniStr() (s string) {
	for i, val := range loc[1:] {
		if moreHeadingsToCome(loc[i+1:]) {
			s += fmt.Sprint(val)+"."
		}
	}
	s = strings.TrimSuffix(s, ".")
	s+=" §"+fmt.Sprint(loc[0]+1)
	return
}


func (loc Location) IsEmpty() (b bool) {
	sum := 0
	for _, j := range(loc) {
		sum += j
	}
	if sum == 0 {
		b = true	
	}
	return
}

func moreHeadingsToCome(loc []int) bool {
	for _, n := range loc {
		if n != 0 {
			return true
		}
	}
	return false
}


