
/* Function to toggle the submenu visibility */
function toggleSubmenu() {
    const submenu = document.getElementById("submenu");
    submenu.style.display = submenu.style.display === "block" ? "none" : "block"; // Toggle submenu
}


/* 
 * TEMPORAY NEED TO REPLACE WITH SERVER SIDE & DB GENERATION 
 *
 * Function to add a new report based on the selected option 
 *
 */

function addReport(reportID) {
    const grid = document.getElementById("report-grid");


    // Create a new report link
    const reportLink = document.createElement("a");
    
    /* ===== GENERATING LINK TO REPORT ===== */

    reportLink.href = `/report/${reportID}`; 


    reportLink.className = "report-item";

    // Add to the grid
    grid.appendChild(reportLink);

    // Hide the submenu after selection
    // document.getElementById("submenu").style.display = "none";
}


function createReport(reportType) {
    // Show the modal
    document.getElementById('submenu').style.display = 'none';
    document.getElementById('modal').style.display = 'flex';

    // Handle the submit button click
    document.getElementById('submitButton').onclick = function() {
        const reportName = document.getElementById('reportNameInput').value;

        if (reportName) {
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
                        addReport(data.reportID);
                    } else {
                        console.error("Failed to create report:", data.reportID);
                    }
                })
                .catch(error => console.error('Error creating report:', error));

            // Hide the modal
            document.getElementById('modal').style.display = 'none';
            document.getElementsByClassName('container').style.display = 'flex';

        } else {
            console.error("Report name is required.");

        }
    };

    // Handle the cancel button click
    document.getElementById('cancelButton').onclick = function() {
        document.getElementById('modal').style.display = 'none';
        document.getElementsByClassName('container').style.display = 'flex';

    };
}
