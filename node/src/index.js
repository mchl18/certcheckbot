import { config } from "dotenv";
import { CertificateChecker } from "./checker/certificate.js";
import { Logger } from "./logger/index.js";
import path from "node:path";

// Load environment variables from project root
config({ path: path.join(process.cwd(), '..', '.env') });

// Configuration validation
const DOMAINS = process.env.DOMAINS?.split(",") || [];
const THRESHOLD_DAYS = process.env.THRESHOLD_DAYS?.split(",").map(Number) || [];
const SLACK_WEBHOOK_URL = process.env.SLACK_WEBHOOK_URL;
const CHECK_INTERVAL = 6 * 60 * 60 * 1000; // 6 hours in milliseconds

if (!SLACK_WEBHOOK_URL) {
  throw new Error("SLACK_WEBHOOK_URL is not set");
}
if (!DOMAINS.length) {
  throw new Error("DOMAINS is not set");
}
if (!THRESHOLD_DAYS.length) {
  throw new Error("THRESHOLD_DAYS is not set");
}

// Initialize logger
const logger = new Logger(path.join(process.cwd(), "logs", "cert-checker.log"));

// Initialize certificate checker
const certChecker = new CertificateChecker(
  DOMAINS,
  THRESHOLD_DAYS,
  SLACK_WEBHOOK_URL,
  logger
);

/**
 * Run the certificate check for all domains
 */
async function runCheck() {
  logger.info("Beginning certificate check cycle", {
    domains: DOMAINS,
    thresholds: THRESHOLD_DAYS,
    checkInterval: CHECK_INTERVAL,
  });

  try {
    await certChecker.checkAll();
    logger.info("Certificate check cycle completed successfully");
  } catch (error) {
    const err = /** @type {Error} */ (error);
    logger.error("Certificate check cycle failed", {
      errorMessage: err.message,
      errorStack: err.stack,
    });
  }
}

// Run initial check
logger.info("Certificate monitoring service initialization", {
  startTime: new Date().toISOString(),
  configuration: {
    domains: DOMAINS,
    thresholds: THRESHOLD_DAYS,
    checkInterval: CHECK_INTERVAL,
  },
});

runCheck();

// Schedule periodic checks
setInterval(runCheck, CHECK_INTERVAL);
