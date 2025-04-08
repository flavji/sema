
/* Function to toggle the submenu visibility */
function toggleSubmenu() {
    const submenu = document.getElementById("submenu");
    submenu.style.display = submenu.style.display === "block" ? "none" : "block"; // Toggle submenu
}


function addReport(reportID, reportTitle) {
    const grid = document.getElementById("report-grid");

    // Create a new report link
    const reportLink = document.createElement("a");
    reportLink.href = `/report/${reportID}`;
    reportLink.className = "report-item";

    // Create a span for the title
    const titleSpan = document.createElement("span");
    titleSpan.className = "report-title";

    // Truncate if title is too long
    const maxLength = 11; // Adjust as needed
    titleSpan.textContent = reportTitle.length > maxLength
        ? reportTitle.substring(0, maxLength) + "..."
        : reportTitle;

    // Add title to the link
    reportLink.appendChild(titleSpan);

    // Add to the grid
    grid.appendChild(reportLink);
}


function createReport() {
    
    // Show the modal
    document.getElementById('submenu').style.display = 'none';

    // put a report type levels here 


    document.getElementById('modal').style.display = 'flex';


    const ealLevelMap = {
        1: "one",
        2: "two",
        3: "three",
        4: "four",
        5: "five",
        6: "six",
        7: "seven"
    };

    // Handle the submit button click
    document.getElementById('submitButton').onclick = function() {
        const reportName = document.getElementById('reportNameInput').value;
        const ealLevel = document.getElementById('ealLevelSelect').value; // Get selected EAL level

        // Ensure both report name and EAL level are provided
        if (reportName && ealLevel) {

            
            reportType = "reportTemplate_evaluation_assurance_level_" + ealLevelMap[ealLevel]
            

            fetch("/api/reports", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    type: reportType,
                    name: reportName
                })
            })
                .then(response => response.json())
                .then(data => {
                    if (data.Success) {
                        console.log("New report ID:", data.reportID);
                        addReport(data.reportID, reportName);
                    } else {
                        error("Failed to create report:", reportName, ": ", data.reportID);
                        console.error("Failed to create report:", data.reportID);
                    }
                })
                .catch(error => {
                    alert('Error creating report:', error);
                    console.error('Error creating report:', error);
                });

            // Hide the modal
            document.getElementById('modal').style.display = 'none';
            // document.getElementsByClassName('container').style.display = 'flex';

        } else {
            alert("Report name is required.");
            console.error("Report name is required.");

        }
    };

    // Handle the cancel button click
    document.getElementById('cancelButton').onclick = function() {
        document.getElementById('modal').style.display = 'none';
        document.getElementsByClassName('container').style.display = 'flex';

    };
}


document.addEventListener("DOMContentLoaded", function() {
    if (window.reports) {
        window.reports.forEach(report => {
            addReport(report.reportID, report.reportTitle);
        });
    }
});

