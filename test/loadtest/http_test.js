import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  vus: 50,             // simulate 50 users
  duration: '30s',     // run test for 30 seconds
};

const BASE_URL = '';

export default function () {
  // hit the register page
  let res = http.get(`${BASE_URL}/register`);
  check(res, { 'register page loaded': (r) => r.status === 200 });

  // fake auth verify call (middleware is disabled anyway)
  res = http.post(`${BASE_URL}/api/auth/verify`, JSON.stringify({ token: 'fake-token' }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'token verify worked': (r) => r.status === 200 });

  // create a new report with random name
  const reportName = `LoadTest_${Math.random().toString(36).substring(7)}`;
  const createPayload = {
    type: 'reportTemplate_evaluation_assurance_level_three',
    name: reportName
  };
  res = http.post(`${BASE_URL}/api/reports`, JSON.stringify(createPayload), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'report created': (r) => r.status === 200 });

  // try to extract the new report ID, fall back if it fails
  let reportID = 'test-report';
  try {
    const body = res.json();
    reportID = body.reportID || reportID;
  } catch (_) {}

  // load the report page
  res = http.get(`${BASE_URL}/report/${reportID}/`);
  check(res, { 'report page loaded': (r) => r.status === 200 });

  // check if current user is admin (normally part of settings page load)
  res = http.get(`${BASE_URL}/report/${reportID}/api/isadmin`);
  check(res, { 'is admin check ok': (r) => r.status === 200 });

  // simulate clicking "Generate Report" button
  res = http.get(`${BASE_URL}/report/${reportID}/api/generateReport?reportID=${reportID}`);
  check(res, { 'generate report attempted': (r) => r.status === 200 || r.status === 400 });

  // simulate renaming the report
  res = http.post(`${BASE_URL}/report/${reportID}/api/renamereport`, JSON.stringify({ reportname: 'Renamed_Report' }), {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'renamed report': (r) => r.status === 200 });

  // open the logs page
  res = http.get(`${BASE_URL}/report/${reportID}/logs`);
  check(res, { 'logs loaded': (r) => r.status === 200 });

  // simulate deleting the report
  res = http.del(`${BASE_URL}/report/${reportID}/api/deletereport`);
  check(res, { 'report deleted': (r) => r.status === 200 });

  // small pause between iterations
  sleep(1);
}
