package main

import (
	"fmt"
	"log"
	"os"
	"time"

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
		Page  string
		Token string
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
	icon := conf.Icons.Allclear
	color := ""
	if haveActives() {
		color = "red"
		icon = conf.Icons.Openincident
	}
	title := fmt.Sprintf("%s SPI A:%d ", icon, len(ActiveIncidents))
	if conf.Scheduled.Enabled {
		title += fmt.Sprintf("S:%d ", len(ScheduledIncidents))
	}
	if conf.Resolved.Enabled {
		title += fmt.Sprintf("R:%d ", len(PastIncidents))
	}
	title += fmt.Sprintf("|color=%s\n", color)
	fmt.Printf(title)
	fmt.Printf("%s Open StatusPage Incidents: %d|color=%s\n", icon, len(ActiveIncidents), color)

	if conf.Scheduled.Enabled {
		fmt.Printf("Scheduled Maintenance Incidents: %d\n", len(ScheduledIncidents))
	}
	if conf.Resolved.Enabled {
		fmt.Printf("Resolved Incidents: %d\n", len(PastIncidents))
	}
	fmt.Printf("---\n")
}

func statusColor(i statuspage.Incident) string {
	switch *i.Status {
	case "investigating":
		return "lightblue"
	case "resolved":
		return "green"
	case "monitoring":
		switch *i.Impact {
		case "major":
			return "orange"
		case "minor":
			return "yellow"
		case "monitoring":
			return "#123def"
		default:
			return "#123def"
		}
	default:
		return ""
	}
}

func impactColor(i statuspage.Incident) string {
	switch *i.Impact {
	case "minor":
		return "#123def"
	case "major":
		return "red"
	default:
		return ""
	}
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
	fmt.Print("---\n")
	fmt.Printf("Scheduled Incidents (%d)|color=orange\n", len(ScheduledIncidents))
	for x, incident := range ScheduledIncidents {
		if x > conf.Scheduled.Limit-1 {
			return nil
		}
		fmt.Printf("Incident: %s\n", *incident.Name)
		fmt.Printf("--Scheduled for: %s\n", formatDateTime(*incident.ScheduledFor))
		fmt.Printf("--Scheduled until: %s\n", formatDateTime(*incident.ScheduledUntil))
		fmt.Printf("--Open in StatusPage.io|href=%s\n", *incident.Shortlink)
		fmt.Printf("--Impact: %s\n", *incident.Impact)
		fmt.Printf("--Status: %s|color=%s\n", *incident.Status, statusColor(incident))
		fmt.Print("--Affected Components:\t")
		if incident.Components != nil {
			fmt.Print("\n")
			for _, c := range *incident.Components {
				fmt.Printf("--_\t%s\n", *c.Name)
			}
			//println("\n")
		} else {
			fmt.Print("None\n")
		}

		for _, iu := range incident.IncidentUpdates {
			fmt.Printf("--Updated At:%s\n", formatDateTime(*iu.CreatedAt))
		}
	}
	return nil
}

func showResolved() error {
	if !conf.Resolved.Enabled {
		return nil
	}
	fmt.Print("---\n")
	fmt.Printf("Past Incidents (%d)|color=green\n", len(PastIncidents))
	for x, incident := range PastIncidents {
		if x > conf.Resolved.Limit-1 {
			break
		}
		fmt.Printf("Incident: %s|color=%s\n", *incident.Name, impactColor(incident))
		fmt.Printf("--Created: %s\n", formatDateTime(*incident.CreatedAt))
		fmt.Printf("--Open in StatusPage.io|href=%s\n", *incident.Shortlink)
		fmt.Printf("--Impact: %s|color=%s\n", *incident.Impact, impactColor(incident))
		fmt.Printf("--Status: %s|color=%s\n", *incident.Status, statusColor(incident))
		fmt.Print("--Affected Components:\t")
		if incident.Components != nil {
			fmt.Print("\n")
			for _, c := range *incident.Components {
				fmt.Printf("--\t%s,", *c.Name)
			}
			println("\n")
		} else {
			fmt.Print("None\n")
		}
		if incident.UpdatedAt != nil {
			fmt.Printf("--Last Update: %s\n", formatDateTime(*incident.UpdatedAt))
		}
	}
	return nil
}

func showStatus(c *cli.Context) error {
	showTitle()
	fmt.Printf("Active Incidents (%d)|color=red\n", len(ActiveIncidents))
	for _, incident := range ActiveIncidents {
		fmt.Printf("Incident: %s|color=%s\n", *incident.Name, impactColor(incident))
		fmt.Printf("--Created: %s\n", formatDateTime(*incident.CreatedAt))
		fmt.Printf("--Open in StatusPage.io|href=%s\n", *incident.Shortlink)
		fmt.Printf("--Impact: %s|color=%s\n", *incident.Impact, impactColor(incident))
		fmt.Printf("--Status: %s|color=%s\n", *incident.Status, statusColor(incident))
		fmt.Print("--Affected Components:\t")
		if incident.Components != nil {
			fmt.Print("\n")
			for x, c := range *incident.Components {
				if x == 0 {
					fmt.Printf("--.\t%s\n", *c.Name)
				} else if x == len(*incident.Components)-1 {
					fmt.Printf("--.\t%s\n", *c.Name)
				} else {
					fmt.Printf("--.\t%s\n,", *c.Name)
				}
			}
		} else {
			fmt.Print("None listed\n")
		}
		if incident.UpdatedAt != nil {
			fmt.Printf("--Last Update: %s\n", formatDateTime(*incident.UpdatedAt))
		}
	}
	//url, err := getPageURL()
	//if err != nil {
	//log.Fatal(err)
	//}
	showResolved()
	showScheduled()
	//fmt.Printf("---\nGo To StatusPage|href=%s\n", url)
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
