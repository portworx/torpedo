package pxbackuplongevityreport

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

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

var reportHTMLStart = `
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

var reportHTMLEnd = `
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

var collapsibleTemplate = `
<button type="button" class="collapsible" style="background-color: COLOROFBG">NAMEOFTHEEVENT</button>
<div class="content">
  DATAFROMEVENT
</div>
`

var (
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
	}

	finalHtml := reportHTMLStart + logHeading + allCollapsibles + reportHTMLEnd

	_, err = f.WriteString(finalHtml)
	if err != nil {
		panic(err)
	}
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
