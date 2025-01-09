const https = require("https");
const fs = require("fs");
const path = require("path");
require("dotenv").config();

// Configuration
const DOMAINS = process.env.DOMAINS.split(","); // Domains to check
const THRESHOLD_DAYS = process.env.THRESHOLD_DAYS.split(",").map(Number); // Alert thresholds in days
const SLACK_WEBHOOK_URL = process.env.SLACK_WEBHOOK_URL; // Slack webhook URL

if (!SLACK_WEBHOOK_URL) {
  throw new Error("SLACK_WEBHOOK_URL is not set");
}
if (!DOMAINS) {
  throw new Error("DOMAINS is not set");
}
if (!THRESHOLD_DAYS) {
  throw new Error("THRESHOLD_DAYS is not set");
}

const HISTORY_FILE = path.join(__dirname, "cert-alerts.json");
const LOG_FILE = path.join(__dirname, "cert-checker.log");
const CHECK_INTERVAL = 6 * 60 * 60 * 1000; // 6 hours in milliseconds

// Initialize logging
function log(message, level = "INFO", details = {}) {
  const timestamp = new Date().toISOString();
  const processInfo = {
    pid: process.pid,
    memory: Math.round(process.memoryUsage().heapUsed / 1024 / 1024) + "MB",
    uptime: Math.round(process.uptime()) + "s",
  };

  const logMessage = `[${timestamp}] [${level}] [PID:${processInfo.pid}] [MEM:${
    processInfo.memory
  }] [UPTIME:${processInfo.uptime}] ${message}\n${
    Object.keys(details).length ? JSON.stringify(details, null, 2) + "\n" : ""
  }`;

  // Write to log file
  fs.appendFileSync(LOG_FILE, logMessage);

  // Also log to console
  console.log(logMessage);
}

// Get certificate info
function checkCertificate(domain) {
  log(`Starting SSL certificate check for domain: ${domain}`, "INFO", {
    checkTime: new Date().toISOString(),
    domain,
    checkType: "SSL_CERTIFICATE",
  });

  const options = {
    host: domain,
    port: 443,
    method: "GET",
  };

  const req = https.request(options, (res) => {
    const cert = res.socket.getPeerCertificate();
    const expirationDate = new Date(cert.valid_to);
    log(`Certificate details retrieved for ${domain}`, "INFO", {
      domain,
      issuer: cert.issuer,
      subject: cert.subject,
      validFrom: new Date(cert.valid_from).toISOString(),
      validTo: expirationDate.toISOString(),
      serialNumber: cert.serialNumber,
      fingerprint: cert.fingerprint,
      protocol: res.socket.getProtocol(),
    });

    const daysToExpiration = Math.floor(
      (expirationDate - new Date()) / (1000 * 60 * 60 * 24)
    );

    log(`Certificate expiration analysis for ${domain}`, "INFO", {
      domain,
      daysRemaining: daysToExpiration,
      expirationDate: expirationDate.toISOString(),
      status: daysToExpiration <= 30 ? "WARNING" : "OK",
    });

    checkAndSendAlert(domain, daysToExpiration, expirationDate);
  });

  req.on("error", (error) => {
    log(`Failed to check certificate for ${domain}`, "ERROR", {
      domain,
      errorCode: error.code,
      errorMessage: error.message,
      errorStack: error.stack,
    });
  });

  req.end();
}

// Load alert history
function loadAlertHistory() {
  try {
    if (fs.existsSync(HISTORY_FILE)) {
      const history = JSON.parse(fs.readFileSync(HISTORY_FILE, "utf8"));
      log(`Alert history loaded successfully`, "INFO", {
        domainsCount: Object.keys(history).length,
        domains: Object.keys(history),
        historyFile: HISTORY_FILE,
      });
      return history;
    }
  } catch (error) {
    log(`Alert history load failed`, "ERROR", {
      errorMessage: error.message,
      errorStack: error.stack,
      historyFile: HISTORY_FILE,
    });
  }
  return {};
}

// Save alert history
function saveAlertHistory(history) {
  try {
    fs.writeFileSync(HISTORY_FILE, JSON.stringify(history, null, 2));
    log(`Alert history saved successfully`, "INFO", {
      domainsCount: Object.keys(history).length,
      domains: Object.keys(history),
      historyFile: HISTORY_FILE,
    });
  } catch (error) {
    log(`Alert history save failed`, "ERROR", {
      errorMessage: error.message,
      errorStack: error.stack,
      historyFile: HISTORY_FILE,
    });
  }
}

// Check if we should send an alert and do so
function checkAndSendAlert(domain, daysToExpiration, expirationDate) {
  const today = new Date().toISOString().split("T")[0];
  const history = loadAlertHistory();

  if (!history[domain]) {
    history[domain] = {};
    log(`New domain alert history initialized`, "INFO", {
      domain,
      timestamp: new Date().toISOString(),
    });
  }

  let thresholdToAlert = null;
  for (const threshold of THRESHOLD_DAYS.sort((a, b) => a - b)) {
    if (daysToExpiration <= threshold) {
      thresholdToAlert = threshold;
      break;
    }
  }

  if (thresholdToAlert) {
    const lastAlert = history[domain][thresholdToAlert];
    if (!lastAlert || lastAlert !== today) {
      log(`Alert threshold reached`, "WARNING", {
        domain,
        threshold: thresholdToAlert,
        daysToExpiration,
        expirationDate: expirationDate.toISOString(),
      });
      sendSlackAlert(
        domain,
        daysToExpiration,
        expirationDate,
        thresholdToAlert
      );
      history[domain][thresholdToAlert] = today;
      saveAlertHistory(history);
    } else {
      log(`Alert already sent today`, "INFO", {
        domain,
        threshold: thresholdToAlert,
        lastAlertDate: lastAlert,
      });
    }
  }
}

// Send alert to Slack
async function sendSlackAlert(
  domain,
  daysToExpiration,
  expirationDate,
  threshold
) {
  const message = {
    text: `🚨 *SSL Certificate Expiration Alert*\nThe SSL certificate for *${domain}* will expire in *${daysToExpiration}* days (${expirationDate.toISOString()}).\nThreshold reached: ${threshold} days\nPlease take action to renew the certificate before it expires.`,
  };

  try {
    log(`Initiating Slack alert`, "INFO", {
      domain,
      webhookUrl: SLACK_WEBHOOK_URL,
      messageContent: message,
    });
    console.log(`Would send Slack alert to ${SLACK_WEBHOOK_URL}: `, message);
    const response = await fetch(SLACK_WEBHOOK_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(message),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    log(`Slack alert sent successfully`, "INFO", {
      domain,
      timestamp: new Date().toISOString(),
    });
  } catch (error) {
    log(`Slack alert sending failed`, "ERROR", {
      domain,
      errorMessage: error.message,
      errorStack: error.stack,
      webhookUrl: SLACK_WEBHOOK_URL,
    });
  }
}

// Check all domains
function checkAllDomains() {
  log("Beginning certificate check cycle", "INFO", {
    domains: DOMAINS,
    thresholds: THRESHOLD_DAYS,
    checkInterval: CHECK_INTERVAL,
  });
  for (const domain of DOMAINS) {
    checkCertificate(domain);
  }
}

// Run initial check and schedule periodic checks
log("Certificate monitoring service initialization", "INFO", {
  startTime: new Date().toISOString(),
  configuration: {
    domains: DOMAINS,
    thresholds: THRESHOLD_DAYS,
    checkInterval: CHECK_INTERVAL,
    historyFile: HISTORY_FILE,
    logFile: LOG_FILE,
  },
});
checkAllDomains();
setInterval(checkAllDomains, CHECK_INTERVAL);
