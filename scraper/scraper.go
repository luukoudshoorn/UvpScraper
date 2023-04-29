package scraper

import (
	"regexp"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Run struct {
	plaats         string
	KSR            bool
	MSR            bool
	LSR            bool
	BSR            bool
	JSR            bool
	inschrijflink  string
	informatielink string
	uitslaglink    string
	organisator    string
	runDatum       time.Time
	inschrijfOpen  time.Time
	inschrijfSluit time.Time
}

func GetRuns() []Run {
	c := colly.NewCollector()
	c.Visit("https://www.uvponline.nl/uvponlineU/index.php/uvproot/wedstrijdschema/2023")

	var runs []Run

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		run := parseRun(e)
		runs = append(runs, run)
	})
	return runs
}

func parseRun(e *colly.HTMLElement) Run {
	var run Run

	categorien := e.ChildText(".agendacircuit")
	run.KSR = strings.Contains(categorien, "K")
	run.MSR = strings.Contains(categorien, "M")
	run.LSR = strings.Contains(categorien, "L")
	run.BSR = strings.Contains(categorien, "B")
	run.JSR = strings.Contains(categorien, "J")

	//Als de inschijving open is, bevat deze class de link
	run.inschrijflink = e.ChildAttr(".inschrijflink_open/a", "href")
	if len(run.inschrijflink) > 0 {
		c := colly.NewCollector()
		c.Visit(run.inschrijflink)
		c.OnHTML("div.form_description", func(e *colly.HTMLElement) {
			r := regexp.MustCompile(`en nog mogelijk tot \d{2}-\d{2}-\d{4}`)
			date := r.FindString(e.Text)
			if len(date) > 0 {
				run.inschrijfSluit, _ = time.Parse("dd-MM-yyyy", date[20:])
			}
		})
	}

	//Is de inschrijving al gesloten?
	inschrijfGesloten := e.ChildAttr(".inschrijflink_closed/a", "href")
	if len(inschrijfGesloten) > 0 {
		run.inschrijflink = inschrijfGesloten
		//Geen idee wanneer de inschrijving sloot, maar het is in het verleden
		run.inschrijfSluit = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	//De inschrijf open bevat ook die link, maar daar willen we juist de datum uit de text halen
	inschrijfOpenText := e.ChildText(".inschrijflink_waiting/a")
	if len(inschrijfOpenText) > 0 {
		eersteSpatie := strings.Index(inschrijfOpenText, " ")
		var err error
		run.inschrijfOpen, err = time.Parse("dd-MM-yyyy hh:mm", inschrijfOpenText[eersteSpatie+1:])
		if err == nil {
			//Pak als bonus ook de inschrijflink mee, want die kunnen we hierboven nog niet gepakt hebben als de inschrijving nog niet open is
			run.inschrijflink = e.ChildAttr(".inschrijflink_waiting/a", "href")
		}
	}

	run.organisator = e.ChildText(".wedstrijdlink/a")
	run.informatielink = e.ChildAttr(".wedstrijdlink/a", "href")

	//Als er een uitslag is, zit ofwel uitslaglink_definitief danwel uitslaglink_voorlopig er in
	run.uitslaglink = e.ChildAttr(".uitslaglink_definitief/a", "href")
	if len(run.uitslaglink) > 0 {
		run.uitslaglink = e.ChildAttr(".uitslaglink_voorlopig/a", "href")
	}

	return run
}
