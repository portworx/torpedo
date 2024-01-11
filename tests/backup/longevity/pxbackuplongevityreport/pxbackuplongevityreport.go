package pxbackuplongevityreport

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests/backup/longevity/pxbackuplongevitytypes"
)

var (
	Results      = make(map[string]ResultForReport)
	ResultsMutex = sync.RWMutex{}
)

type ResultForReport struct {
	Status bool
	Data   EventResponse
}

type ResultForSuccessFailurePercentage struct {
	Failed map[string]int
	Passed map[string]int
}

var SuccessFailurePercentage = ResultForSuccessFailurePercentage{
	Failed: make(map[string]int),
	Passed: make(map[string]int),
}

type ResultForSuccessStatus struct {
	TotalPass int
	TotalFail int
}

var SuccessStatus = ResultForSuccessStatus{
	TotalPass: 0,
	TotalFail: 0,
}

type ResultTimeAnalysisForEvents struct {
	TotalEvents          map[string]int
	EventCompletionTimes map[string]map[time.Time]float32
	TotalTimeTaken       map[string]float32
}

var TimeAnalysisForEvents = ResultTimeAnalysisForEvents{
	TotalEvents:          make(map[string]int),
	EventCompletionTimes: make(map[string]map[time.Time]float32),
	TotalTimeTaken:       make(map[string]float32),
}

var (
	reportHTMLStart = `
<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
.collapsible {
  color: white;
  cursor: pointer;
  padding: 18px;
  width: 100%;
  border: none;
  text-align: left;
  outline: none;
  font-size: 15px;
}

.active, .collapsible:hover {
  background-color: #555;
}

.content {
  padding: 0 18px;
  display: none;
  overflow: hidden;
  background-color: #f1f1f1;
}

table {
 font-family: arial, sans-serif;
 border-collapse: collapse;
 width: 100%;
}
  
td, th {
 border: 1px solid #000000; 
 text-align: left;
 padding: 8px;
}
  
</style>
</head>
<body>
`

	reportHTMLEnd = `
<script>
var coll = document.getElementsByClassName("collapsible");
var i;

for (i = 0; i < coll.length; i++) {
  coll[i].addEventListener("click", function() {
    this.classList.toggle("active");
    var content = this.nextElementSibling;
    if (content.style.display === "block") {
      content.style.display = "none";
    } else {
      content.style.display = "block";
    }
  });
}
</script>

</body>
</html>
`

	collapsibleTemplate = `
<button type="button" class="collapsible" style="background-color: COLOROFBG">NAMEOFTHEEVENT</button>
<div class="content">
  DATAFROMEVENT
</div>
`

	tableStart = `
<table>
<tr style="background-color: grey">
<th>EventName</th>
<th>Errors</th>
<th>HighlightEvents</th>
<th>TimeTaken(Minutes)</th>
</tr>
`
	tableEnd = `
</table>
`
)

func DumpResult() {
	var tempString string

	f, err := os.OpenFile("data1.html", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	logHeading := fmt.Sprintf("<h2>Logged At: %s</h2>", time.Now().Format("02 Jan 06 15:04 MST"))
	allCollapsibles := ""
	for eventName, eventDetails := range Results {
		tempString = ""

		if !eventDetails.Status {
			tempString = strings.Replace(collapsibleTemplate, "COLOROFBG", "red", 1)
		} else {
			tempString = strings.Replace(collapsibleTemplate, "COLOROFBG", "green", 1)
		}

		tempString = strings.Replace(tempString, "NAMEOFTHEEVENT", eventName, 1)
		eventDataToTables := createTableFromData(eventDetails.Data)
		tempString = strings.Replace(tempString, "DATAFROMEVENT", eventDataToTables, 1)

		allCollapsibles += tempString

		UpdateDetailsForCharts(eventDetails.Data)
	}

	finalHtml := reportHTMLStart + logHeading + allCollapsibles + reportHTMLEnd

	_, err = f.WriteString(finalHtml)
	if err != nil {
		panic(err)
	}

	CreateGraphs()

	log.Infof("[%+v] - [%+v]", SuccessFailurePercentage, SuccessStatus)
}

func CreateGraphs() {
	createGraphForFailPass()
	createGraphForFailurePercentage()
}

func createTableFromData(eventData EventResponse) string {
	var createdTable string

	for eventName, response := range eventData.EventBuilders {
		tempString := "<tr>"
		tempString += fmt.Sprintf("<td>%s</td>", eventName)
		if response.Error != nil {
			tempString += fmt.Sprintf("<td>%s</td>", response.Error)
		} else {
			tempString += fmt.Sprintf("<td>%s</td>", "")
		}
		if response.HighlightEvent != "" {
			tempString += fmt.Sprintf("<td>%s</td>", response.HighlightEvent)
		} else {
			tempString += fmt.Sprintf("<td>%s</td>", "")
		}
		tempString += fmt.Sprintf("<td>%f</td>", response.TimeTakenInMinutes)
		tempString += "</tr>"

		createdTable += tempString
	}

	return tableStart + createdTable + tableEnd
}

func createGraphForFailPass() {
	// initialize chart
	pie := charts.NewPie()

	// preformat data
	pieData := []opts.PieData{
		{Name: "Passed", Value: SuccessStatus.TotalPass},
		{Name: "Failed", Value: SuccessStatus.TotalFail},
	}

	// put data into chart
	pie.AddSeries("Overall Test Status", pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: true, Formatter: "{b}: {c}"}),
	)

	// generate chart and write it to io.Writer
	f, _ := os.Create("pie.html")
	pie.Render(f)
}

func createGraphForFailurePercentage() {
	pie := charts.NewPie()

	pieData := []opts.PieData{}

	for eventName, eventCount := range SuccessFailurePercentage.Failed {
		pieData = append(pieData, opts.PieData{Name: eventName, Value: eventCount})
	}

	// put data into chart
	pie.AddSeries("Failure Percentage", pieData).SetSeriesOptions(
		charts.WithLabelOpts(opts.Label{Show: true, Formatter: "{b}: {c}"}),
	)

	// generate chart and write it to io.Writer
	f, _ := os.Create("pie_fails.html")
	pie.Render(f)
}

func UpdateDetailsForCharts(eventResponse EventResponse) {
	if !eventResponse.Status {
		if val, ok := SuccessFailurePercentage.Failed[eventResponse.Name]; ok {
			SuccessFailurePercentage.Failed[eventResponse.Name] = 1 + val
		} else {
			SuccessFailurePercentage.Failed[eventResponse.Name] = 1
		}
		SuccessStatus.TotalFail += 1
	} else {
		if val, ok := SuccessFailurePercentage.Passed[eventResponse.Name]; ok {
			SuccessFailurePercentage.Passed[eventResponse.Name] = 1 + val
		} else {
			SuccessFailurePercentage.Passed[eventResponse.Name] = 1
		}

		SuccessStatus.TotalPass += 1

		if _, ok := TimeAnalysisForEvents.EventCompletionTimes[eventResponse.Name]; ok {
			TimeAnalysisForEvents.EventCompletionTimes[eventResponse.Name][time.Now()] = eventResponse.TimeTakenInMinutes
		} else {
			TimeAnalysisForEvents.EventCompletionTimes[eventResponse.Name] = make(map[time.Time]float32)
			TimeAnalysisForEvents.EventCompletionTimes[eventResponse.Name][time.Now()] = eventResponse.TimeTakenInMinutes
		}

		if _, ok := TimeAnalysisForEvents.TotalEvents[eventResponse.Name]; ok {
			TimeAnalysisForEvents.TotalEvents[eventResponse.Name] += 1
		} else {
			TimeAnalysisForEvents.TotalEvents[eventResponse.Name] = 1
		}

		if _, ok := TimeAnalysisForEvents.TotalTimeTaken[eventResponse.Name]; ok {
			TimeAnalysisForEvents.TotalTimeTaken[eventResponse.Name] += eventResponse.TimeTakenInMinutes
		} else {
			TimeAnalysisForEvents.TotalEvents[eventResponse.Name] = int(eventResponse.TimeTakenInMinutes)
		}
	}

}
