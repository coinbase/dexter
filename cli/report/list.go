package report

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/coinbase/dexter/engine"
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/util"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

//
// A report struct contains the information learned about a report
// from listing the reports S3 directory.
//
type Report struct {
	Investigation      engine.Investigation
	ID                 string
	HostsUploaded      int
	RecipientsUploaded []string
}

func listReports(cmd *cobra.Command, args []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Investigation",
		"Issuer",
		"Tasks",
		"Scope",
		"Recipients",
		"Hosts Uploaded",
	})

	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
	)

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
	)

	for _, rep := range CurrentReports() {
		table.Append([]string{
			rep.ID,
			rep.Investigation.Issuer.Name,
			strings.Join(helpers.TaskStrings(rep.Investigation.TaskList), ",\n"),
			strings.Join(rep.Investigation.ScopeFactsStrings(), ",\n"),
			strings.Join(rep.RecipientsUploaded, ",\n"),
			strconv.Itoa(rep.HostsUploaded),
		})

	}

	table.Render()
}

//
// List all currently available reports.
//
func CurrentReports() []Report {
	allReportIDs := make([]string, 0)
	reportedHosts := make(map[string][]string)
	reportedUsers := make(map[string][]string)
	reports := make([]Report, 0)

	reportFiles, err := helpers.ListS3Path("reports/")
	if err != nil {
		color.HiRed(err.Error())
		return []Report{}
	}
	cachedInvestigations := engine.CurrentInvestigations()
	for _, filename := range reportFiles {
		if !strings.HasSuffix(filename, ".zip.enc") {
			continue
		}
		re := regexp.MustCompile(`reports/(.+?)-(.+)\.(.+)\.zip\.enc`)
		matches := re.FindStringSubmatch(filename)
		if len(matches) < 4 {
			log.WithFields(log.Fields{
				"at":       "report.CurrentReports",
				"filename": filename,
			}).Error("regex mismatch on filename")
			continue
		}
		uuid := matches[1]
		hostname := matches[2]
		recipientName := matches[3]

		if !util.StringsInclude(allReportIDs, uuid) {
			allReportIDs = append(allReportIDs, uuid)
		}
		if !util.StringsInclude(reportedHosts[uuid], hostname) {
			reportedHosts[uuid] = append(reportedHosts[uuid], hostname)
		}
		if !util.StringsInclude(reportedUsers[uuid], recipientName) {
			reportedUsers[uuid] = append(reportedUsers[uuid], recipientName)
		}
	}
	for _, uuid := range allReportIDs {
		investigation, err := engine.InvestigationByIDWithCache(cachedInvestigations, uuid)
		if err != nil {
			color.HiRed("Error loading investigation for report " + uuid + ", investigation was probably pruned: " + err.Error())
		}
		reports = append(reports, Report{
			ID:                 uuid,
			Investigation:      investigation,
			HostsUploaded:      len(reportedHosts[uuid]),
			RecipientsUploaded: reportedUsers[uuid],
		})
	}
	return reports
}
