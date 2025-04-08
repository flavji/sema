const WebSocket = require('ws');

const totalClients = 200;
const clientsPerPhase = 50;
const phaseDelay = 3000; // 3 seconds between each wave
const sessionDuration = 10000; // how long each user stays connected

// Define multiple reports and sections
const reportIDs = ['report-a', 'report-b', 'report-c'];
const sectionIDs = ['Intro', 'Overview', 'Details', 'Conclusion'];

let clientIndex = 0;

function startClient(i) {
  // Distribute clients across reports/sections evenly
  const reportID = reportIDs[i % reportIDs.length];
  const sectionID = sectionIDs[i % sectionIDs.length];
  const target = `ws://192.168.1.88:8080/report/${reportID}/section/${sectionID}`;

  const ws = new WebSocket(target);

  ws.on('open', () => {
    console.log(`[${i}] Connected to ${reportID}/${sectionID}`);

    // Send join message
    ws.send(JSON.stringify({
      type: 'join',
      reportid: reportID,
      section: sectionID
    }));

    // Send edits every second
    const interval = setInterval(() => {
      ws.send(JSON.stringify({
        type: 'delta',
        delta: {
          editorId: `editor-${i}`,
          delta: {
            ops: [{ insert: `Client ${i} editing ${sectionID}\n` }]
          }
        }
      }));
    }, 1000);

    // Disconnect after sessionDuration
    setTimeout(() => {
      ws.send(JSON.stringify({ type: 'close', section: sectionID }));
      ws.close();
      clearInterval(interval);
    }, sessionDuration);
  });

  ws.on('error', err => {
    console.error(`[${i}] Error:`, err.message);
  });

  ws.on('close', () => {
    console.log(`[${i}] Connection closed`);
  });
}

// Ramp users in phases
function runPhase() {
  const end = Math.min(clientIndex + clientsPerPhase, totalClients);
  for (; clientIndex < end; clientIndex++) {
    startClient(clientIndex);
  }

  if (clientIndex < totalClients) {
    setTimeout(runPhase, phaseDelay);
  }
}

runPhase(); // Start the first wave
