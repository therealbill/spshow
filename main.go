package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johnmccabe/bitbar"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	statuspage "github.com/yfronto/go-statuspage-api"
	"gopkg.in/gcfg.v1"
)

var (
	app                *cli.App
	conf               Config
	configfile         string
	client             *statuspage.Client
	ActiveIncidents    []statuspage.Incident
	AllIncidents       []statuspage.Incident
	PastIncidents      []statuspage.Incident
	ScheduledIncidents []statuspage.Incident
	dateFormat         = "2006-01-02 15:04 MST"
	bar                bitbar.Plugin
)

func getClient() (c *statuspage.Client, err error) {
	loadCredentials()
	client, err = statuspage.NewClient(conf.Main.Token, conf.Main.Page)
	if err != nil {
		log.Print("NO client")
		log.Fatal(err)
	}
	return client, err
}

type Config struct {
	Main struct {
		Page    string
		Token   string
		Title   string
		Baricon string
	}
	Scheduled struct {
		Enabled bool
		Limit   int
	}
	Resolved struct {
		Enabled bool
		Limit   int
	}
	Icons struct {
		Openincident string
		Allclear     string
	}
}

func loadCredentials() error {
	fp, err := homedir.Expand(configfile)
	if err != nil {
		log.Fatal(err)
	}
	err = gcfg.ReadFileInto(&conf, fp)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func loadAllIncidents(c *cli.Context) (err error) {
	loadCredentials()
	getClient()
	all, err := client.GetAllIncidents()
	if err != nil {
		log.Fatal(err)
	}
	AllIncidents = all

	open, err := client.GetOpenIncidents()
	if err != nil {
		log.Fatal(err)
	}
	ActiveIncidents = open

	sched, err := client.GetScheduledIncidents()
	if err != nil {
		log.Fatal(err)
	}
	ScheduledIncidents = sched

	var pi []statuspage.Incident
	for _, i := range AllIncidents {
		if *i.Status == "resolved" {
			pi = append(pi, i)
		}
	}
	PastIncidents = pi
	return err
}
func haveActives() bool {
	return len(ActiveIncidents) > 0
}

func showTitle() {
	wtftrue := true
	style_clear := bitbar.Style{}
	style_active := bitbar.Style{Color: "red", Trim: &wtftrue, Length: 24}
	//style_active := bitbar.Style{Color: "red"}
	//icon := conf.Icons.Allclear
	//if haveActives() { icon = conf.Icons.Openincident }

	title := fmt.Sprintf("%s %d", conf.Main.Title, len(ActiveIncidents))

	if conf.Scheduled.Enabled {
		title += fmt.Sprintf("/%d", len(ScheduledIncidents))
	}
	if conf.Resolved.Enabled {
		title += fmt.Sprintf("/%d", len(PastIncidents))
	}
	if haveActives() {
		//icon = conf.Icons.Openincident
		//aline := fmt.Sprintf("%s Open Incidents: %d", icon, len(ActiveIncidents))
		bar.StatusLine(title).Style(style_active).DropDown(false).Image(conf.Main.Baricon)
		//bar.StatusLine(aline).Style(style_active)
	} else {
		//aline := fmt.Sprintf("%s Open Incidents: %d", icon, len(ActiveIncidents))
		bar.StatusLine(title).Style(style_clear).Image("https://global.localizecdn.com/0-images/integrations/statuspage.png")
		//bar.StatusLine(aline).Style(style_clear)
	}

	if conf.Scheduled.Enabled {
		//bar.StatusLine(fmt.Sprintf("Scheduled Maintenance Incidents: %d", len(ScheduledIncidents)))
	}
	if conf.Resolved.Enabled {
		//bar.StatusLine(fmt.Sprintf("Resolved Incidents: %d", len(PastIncidents)))
	}
}

func statusStyle(i statuspage.Incident, s bitbar.Style) bitbar.Style {
	switch *i.Status {
	case "investigating":
		s.Color = "lightblue"
	case "resolved":
		s.Color = "green"
	case "monitoring":
		switch *i.Impact {
		case "major":
			s.Color = "orange"
		case "minor":
			s.Color = "yellow"
		case "monitoring":
			s.Color = "#123def"
		default:
			s.Color = "#123def"
		}
	default:
		s.Color = ""
	}
	return s
}

func impactStyle(i statuspage.Incident, s bitbar.Style) bitbar.Style {
	switch *i.Impact {
	case "minor":
		s.Color = "#123def"
	case "major":
		s.Color = "red"
	default:
	}
	return s
}

func getPageURL() (string, error) {
	return "Not Implemented in Client Yet", nil
}

func formatDateTime(t time.Time) string {
	localLoc, _ := time.LoadLocation("Local")
	localTime := t.In(localLoc)
	return localTime.Format(dateFormat)
}

func showScheduled() error {
	if !conf.Scheduled.Enabled {
		return nil
	}
	menu := bar.SubMenu
	ss := bitbar.Style{
		Font:  "UbuntuMono-Bold",
		Color: "orange",
		Size:  18,
	}
	menu.Line(fmt.Sprintf("Scheduled Incidents (%d)", len(ScheduledIncidents))).Style(ss)
	for x, incident := range ScheduledIncidents {
		if x > conf.Scheduled.Limit-1 {
			return nil
		}
		menu.Line(fmt.Sprintf("Incident: %s", *incident.Name))
		im := menu.NewSubMenu()
		im.Line(fmt.Sprintf("Scheduled for: %s", formatDateTime(*incident.ScheduledFor)))
		im.Line(fmt.Sprintf("Scheduled until: %s", formatDateTime(*incident.ScheduledUntil)))
		im.Line(fmt.Sprintf("Open in StatusPage.io|href=%s", *incident.Shortlink))
		im.Line(fmt.Sprintf("Impact: %s", *incident.Impact))
		im.Line(fmt.Sprintf("Status: %s", *incident.Status)).Style(statusStyle(incident, ss))
		for _, iu := range incident.IncidentUpdates {
			im.Line(fmt.Sprintf("Updated At:%s", formatDateTime(*iu.CreatedAt)))
		}
		im.Line(fmt.Sprintf("Affected Components:\t"))
		cm := im.NewSubMenu()
		if incident.Components != nil {
			for _, c := range *incident.Components {
				cm.Line(*c.Name)
			}
		} else {
			cm.Line("None")
		}

	}
	return nil
}

func showResolved() error {
	if !conf.Resolved.Enabled {
		return nil
	}
	menu := bar.SubMenu
	ss := bitbar.Style{
		Font:  "UbuntuMono-Bold",
		Color: "green",
		Size:  18,
	}
	is := bitbar.Style{
		Font:  "UbuntuMono-Bold",
		Color: "green",
		Size:  14,
	}
	menu.Line(fmt.Sprintf("Past Incidents (%d)", len(PastIncidents))).Style(ss)
	for x, incident := range PastIncidents {
		if x > conf.Resolved.Limit-1 {
			break
		}
		menu.Line(fmt.Sprintf("Incident: %s", *incident.Name))
		im := menu.NewSubMenu()
		im.Line(fmt.Sprintf("Created: %s\n", formatDateTime(*incident.CreatedAt)))
		im.Line("Open in StatusPage.io").Href(*incident.Shortlink)
		im.Line(fmt.Sprintf("Impact: %s", *incident.Impact)).Style(impactStyle(incident, is))
		im.Line(fmt.Sprintf("Status: %s", *incident.Status)).Style(statusStyle(incident, is))
		for _, iu := range incident.IncidentUpdates {
			im.Line(fmt.Sprintf("Updated At:%s", formatDateTime(*iu.CreatedAt)))
		}
		im.Line("Affected Components")
		cm := im.NewSubMenu()
		if incident.Components != nil {
			for _, c := range *incident.Components {
				cm.Line(*c.Name)
			}
		} else {
			cm.Line("None")
		}
	}
	return nil
}

func handleActives() {
	if haveActives() {
		menu := bar.NewSubMenu()
		s := bitbar.Style{Font: "UbuntuMono-Bold", Color: "red"}
		menu.Line(fmt.Sprintf("Active Incidents (%d)", len(ActiveIncidents))).Style(s).Size(16).Font("Avenir-Bold")
		for _, incident := range ActiveIncidents {
			menu.Line(fmt.Sprintf("%s Incident: %s", conf.Icons.Openincident, *incident.Name)).Style(impactStyle(incident, s))
			im := menu.NewSubMenu()
			im.Line(fmt.Sprintf("Created: %s", formatDateTime(*incident.CreatedAt)))
			im.Line("Open in StatusPage.io").Href(*incident.Shortlink)
			im.Line(fmt.Sprintf("Impact: %s", *incident.Impact)).Style(impactStyle(incident, s))
			im.Line(fmt.Sprintf("Status: %s", *incident.Status)).Style(statusStyle(incident, s))
			im.Line("Affected Components:  ")
			cm := menu.NewSubMenu()
			if incident.Components != nil {
				for x, c := range *incident.Components {
					if x == 0 {
						cm.Line(*c.Name)
					} else if x == len(*incident.Components)-1 {
						cm.Line(*c.Name)
					} else {
						cm.Line(*c.Name)
					}
				}
			} else {
				cm.Line("None listed\n")
			}
			if incident.UpdatedAt != nil {
				im.Line(fmt.Sprintf("Last Update: %s\n", formatDateTime(*incident.UpdatedAt)))
			}
		}
	}
}

func showStatus(c *cli.Context) error {
	showTitle()
	handleActives()
	showScheduled()
	//url, err := getPageURL()
	//if err != nil {
	//log.Fatal(err)
	//}
	showResolved()
	//fmt.Printf("---\nGo To StatusPage|href=%s\n", url)
	fmt.Print(bar.Render())
	return nil
}

func main() {
	app = cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config,c",
			Usage:       "Config file to use",
			Value:       "~/.spshow",
			Destination: &configfile,
		},
		cli.StringFlag{
			Name:        "authToken,a",
			Usage:       "Statuspage.io auth token",
			EnvVar:      "AUTHTOK",
			Destination: &conf.Main.Token,
		},
		cli.StringFlag{
			Name:        "pageid,p",
			Usage:       "Statuspage.io Page ID",
			EnvVar:      "PAGEID",
			Destination: &conf.Main.Page,
		},
	}
	app.Action = showStatus
	app.Before = loadAllIncidents
	app.Run(os.Args)
}
