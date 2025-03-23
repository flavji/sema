/* The quill editors */
let editors = {};

/* Have one websocket to update the sections editors */
let currentSocket = null;
let currentSection = null;


/* Applys change to the relevant editor */
function applyDeltaToEditor(delta) {
  const editorId = delta.editorId; // The editor id
  if (editors[editorId]) {
    editors[editorId].updateContents(delta.delta); // update the delta of the editor
    console.log(`Apply new delta:`, delta.delta,` at `, editorId);
  }
}

/* Send changes user made in editor to the server */
function sendDeltaToServer(delta) {
  if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
    const message = {
      type: 'delta',
      delta: delta
    };
    msg = JSON.stringify(message);
    currentSocket.send(msg);
    console.log(`Sending new delta to web server:`, msg);
  }
}


function broadcastSectionContents(reportId, section) {
  let contents = {}; // Initialize contents as an empty object

  Object.keys(editors).forEach(editorId => {
    const delta = editors[editorId].getContents(); // Get the delta from the editor
    contents[editorId] = {  // Add each editor's delta to the contents object type: 'delta', 
      type: 'delta',
      delta: {editorId: editorId, delta: delta}
    };
    console.log(editorId);
  });


  // Check if the socket is open and then send the contents
  if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
    currentSocket.send(JSON.stringify({ type: 'sync', reportid: reportId, section: section, contents: contents}));
    console.log(`Sending sync contents to websever:`, contents)
  }
}

function updateRepoSectionContents(reportId, section) {
  let contents = {}; // Initialize contents as an empty object

  Object.keys(editors).forEach(editorId => {
    const delta = editors[editorId].getContents(); // Get the delta from the editor
    contents[editorId] = {  // Add each editor's delta to the contents object type: 'delta', 
      type: 'delta',
      delta: {editorId: editorId, delta: delta}
    };
    console.log(editorId);
  });

  // Check if the socket is open and then send the contents
  if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
    currentSocket.send(JSON.stringify({ type: 'updateRepo', reportid: reportId, section: section, contents: contents}));
    console.log(`Sending repo contents to websever:`, contents)
  }
}


/* Probably don't need this function, just gets the reportid from the url */
function getReportId() {
  const pathParts = window.location.pathname.split('/');
  // Assuming the report ID is always the second part of the URL after '/report/'
  return pathParts[2];  
}


/* Open and manage the websocket */
function openSocket(section) {
  // Close previous sections socket
  /*
  if (currentSocket) { 
    currentSocket.close(); 
    console.log(`Closed WebSocket for previous section`);
  }
  */

  /* Open new websocket */
  const reportId = getReportId(); // handle error
  const webSocketUrl = `ws://192.168.1.88:8080/report/${encodeURIComponent(reportId)}/section/${encodeURIComponent(section)}`

  currentSocket = new WebSocket(webSocketUrl, ["binary"]);
  console.log("Attempting to open socket:", webSocketUrl);

  /* Trigger async "join" message to server */
  currentSocket.onopen = function() {
    console.log(`Opened socket for ${section}`);
    currentSocket.send(JSON.stringify({type: 'join', reportid: reportId, section: section}));

  };

  /* Trigger async message hander */
  currentSocket.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log(`Received data at ${section}: `, data) 
    if (data.type == 'delta') {
      applyDeltaToEditor(data.delta);
       
    } else if (data.action == 'request_contents') {
      broadcastSectionContents(reportId, section);   
      console.log(`Sent section contents for ${section}`);
    }

  };

  /* Async error handler */
  currentSocket.onerror = function(error) {
    console.log(`Error at ${section}: `, error) 
  };

  /* Handle WebSocket closure */
  currentSocket.onclose = function(event) {
    console.log(`Closed socket for previous section: `, event.reason, event.code);

  };

}

function cleanPreviousSection() {
  /* Handle WebSocket closure */
  reportId = getReportId();
  console.log("closing ", currentSection, "with editors:", editors);
  updateRepoSectionContents(reportId, currentSection);
  editors = {};
  if (currentSocket) {
    currentSocket.send(JSON.stringify({type: 'close', section: currentSection}));
    currentSocket.close();
  }
  console.log("previous secton cleaned");
}



function loadSubsections(section) {
   
  if (currentSection == 'settings') {
    const settingsDiv = document.getElementById('settings');
    settingsDiv.style.display = 'none'; 
  } else {
    cleanPreviousSection();
  }


  currentSection = section;
  console.log('Loading subsections for section:', section);

  const submenu = document.getElementById('submenu');
  const editorsDiv = document.getElementById('editors');
  const mainHeader = document.getElementById('main-header');

  // Check if the submenu, editorsDiv, and mainHeader elements exist
  if (!submenu || !editorsDiv || !mainHeader) {
    console.error('Missing one or more required DOM elements');
    return;
  }

  submenu.innerHTML = ''; // Clear previous submenu links
  editorsDiv.innerHTML = ''; // Clear previous editors
  mainHeader.textContent = section;

  // Ensure subsections data is valid
  if (!subsections || !Array.isArray(subsections)) {
    console.error('Subsections data is invalid:', subsections);
    return;
  }

  // Find the subsection object by matching the title
  const sectionObject = subsections.find(subsection => subsection.Title === section);

  if (!sectionObject) {
    console.error(`Section "${section}" not found in subsections`);
    submenu.innerHTML = '<p style="color: white; padding: 10px;">Section not found</p>';
    return;
  }

  const sectionSubsections = sectionObject.Subsections || [];

  // Check if subsections for this section exist
  if (sectionSubsections.length === 0) {
    submenu.innerHTML = '<p style="color: white; padding: 10px;">No subsections available</p>';
    return;
  }

  console.log('Subsections for section:', section, sectionSubsections);

  sectionSubsections.forEach((subsection, index) => {
    console.log('Processing subsection:', subsection);

    // Add submenu link
    const link = document.createElement('a');
    link.href = `#${subsection}`;
    link.textContent = subsection;
    submenu.appendChild(link);

    // Add Quill editor for subsection
    const editorContainer = document.createElement('div');
    editorContainer.classList.add('editor-container');

    const editorHeader = document.createElement('div');
    editorHeader.classList.add('editor-header');
    editorHeader.textContent = subsection;

    const editorDiv = document.createElement('div');
    editorDiv.id = `editor-${section.replace(/\s+/g, '_')}-${index}`;
    editorDiv.style.height = '200px';
    editorDiv.style.border = '1px solid #ccc';

    editorContainer.appendChild(editorHeader);
    editorContainer.appendChild(editorDiv);
    editorsDiv.appendChild(editorContainer);

    // Check if Quill is loaded and available
    if (typeof Quill === 'undefined') {
      console.error('Quill is not defined. Make sure Quill is included properly.');
      return;
    }

    const toolbarOptions = [
      ['bold', 'italic', 'underline'],
      [{ 'list': 'ordered' }, { 'list': 'bullet' }],
      ['link', 'image']
    ];

    try {
      // Initialize Quill editor
      // editors[editorDiv.id] = new Quill(`#${editorDiv.id}`, {
      editors[subsection] = new Quill(`#${editorDiv.id}`, {
        theme: 'snow',
        placeholder: `Edit content for ${subsection}...`,
        modules: {
          toolbar: toolbarOptions
        }
      });

      // Attach event listener for text changes
      editors[subsection].on('text-change', function (delta, _, source) {
        if (source === 'user') {
          sendDeltaToServer({
            editorId: subsection,
            delta: delta
          });
        }
      });
    } catch (error) {
      console.error('Error initializing Quill editor for subsection:', subsection, error);
    }
  });

  openSocket(section);
}


function loadSettings() {

  cleanPreviousSection();
  currentSection = 'settings';
  const settingsDiv = document.getElementById('settings');

  if (!settingsDiv) {
    console.error('Settings container not found');
    return;
  }

  const submenu = document.getElementById('submenu');
  const editorsDiv = document.getElementById('editors');
  const mainHeader = document.getElementById('main-header');

  // Check if the submenu, editorsDiv, and mainHeader elements exist
  if (!submenu || !editorsDiv || !mainHeader) {
    console.error('Missing one or more required DOM elements');
    return;
  }

  submenu.innerHTML = ''; // Clear previous submenu links
  editorsDiv.innerHTML = ''; // Clear previous editors
  mainHeader.textContent = 'Settings';

  
  settingsDiv.style.display = 'flex';
  settingsDiv.innerHTML = ''; // Clear previous content

  const reportButton = document.createElement('button');
  reportID = getReportId();
  reportButton.textContent = 'Report Generate';
  reportButton.onclick = function () {
    fetch("/api/generateReport", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        type: 'reportGeneration',
        reportID: reportID
      })
    })
    /*
      .then(response => response.json())
      .then(data => {
        if (data.Success) {
          console.log("New report ID:", data.reportID);
        } else {
          console.error("Failed to create report:", data.reportID);
        }
      })
      .catch(error => console.error('Error creating report:', error));
      */

  };

  const addUserButton = document.createElement('button');
  addUserButton.textContent = 'Add Users';
  addUserButton.onclick = function () {
    console.log('Add Users clicked');
  };

  settingsDiv.appendChild(reportButton);
  settingsDiv.appendChild(addUserButton);
}


// Function to dynamically load section links
function loadSectionLinks() {

  const sectionsLinksDiv = document.getElementById('sections-links');

  // Ensure the sections data is available
  if (!subsections || !Array.isArray(subsections)) {
    console.error('Subsections data is invalid:', subsections);
    return;
  }

  subsections.forEach(section => {
    const sectionTitle = section.Title;

    // Create a new link element
    const link = document.createElement('a');
    link.href = 'javascript:void(0)';
    link.textContent = sectionTitle;

    // Add onclick event to call loadSubsections with the section title
    link.setAttribute('onclick', `loadSubsections('${sectionTitle}')`);

    // Append the link to the sections container
    sectionsLinksDiv.appendChild(link);

  });
}

window.onload = function() {
  loadSectionLinks(); // Create the section links dynamically
  loadSettings();
};


function cleanup() {
  // Clean up or save necessary data
  console.log("Page is unloading, perform cleanup.");
  reportId = getReportId();
  updateRepoSectionContents(reportId, currentSection);
  editors = {};
  if (currentSocket) {
    currentSocket.send(JSON.stringify({type: 'close', section: currentSection}));
    currentSocket.close();
  }

};

window.addEventListener('beforeunload', function (event) {
  event.preventDefault();
  cleanup();  
  // location.reload() -> go back to settings page
  loadSettings();
  event.returnValue = 'Are you sure you want to leave?';
});


/*
window.addEventListener('unload', function () {
    cleanup(); // More likely to run before the page fully closes
});
*/

