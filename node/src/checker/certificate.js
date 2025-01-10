import https from "node:https";
import { SlackNotifier } from "../alert/slack.js";
import { HistoryManager } from "../storage/history.js";

/**
 * @typedef {Object} CertificateDetails
 * @property {Object} issuer - Certificate issuer details
 * @property {Object} subject - Certificate subject details
 * @property {string} valid_from - Certificate validity start date
 * @property {string} valid_to - Certificate validity end date
 * @property {string} serialNumber - Certificate serial number
 * @property {string} fingerprint - Certificate fingerprint
 */

/**
 * @typedef {Object} HistoryRecord
 * @property {Object.<string, Object.<number, string>>} [domain] - Domain-specific alert history
 */

/**
 * Handles SSL certificate checking functionality
 */
export class CertificateChecker {
  /**
   * @param {string[]} domains - List of domains to check
   * @param {number[]} thresholdDays - List of threshold days for alerts
   * @param {string} slackWebhookUrl - Slack webhook URL
   * @param {import('../logger').Logger} logger - Logger instance
   */
  constructor(domains, thresholdDays, slackWebhookUrl, logger) {
    /** @private @type {string[]} */
    this.domains = domains;
    /** @private @type {number[]} */
    this.thresholdDays = thresholdDays;
    /** @private @type {import('../logger').Logger} */
    this.logger = logger;
    this.slackNotifier = new SlackNotifier(slackWebhookUrl);
    this.historyManager = new HistoryManager();
  }

  async checkAll() {
    for (const domain of this.domains) {
      await this.checkCertificate(domain);
    }
  }

  /**
   * Check certificate for a single domain
   * @param {string} domain - Domain to check
   * @returns {Promise<void>}
   * @throws {Error} When certificate check fails
   */
  async checkCertificate(domain) {
    this.logger.info(`Starting SSL certificate check for domain: ${domain}`, {
      checkTime: new Date().toISOString(),
      domain,
      checkType: "SSL_CERTIFICATE",
    });

    return new Promise((resolve, reject) => {
      const options = {
        host: domain,
        port: 443,
        method: "GET",
      };

      const req = https.request(options, async (res) => {
        // @ts-ignore
        const cert = res.socket.getPeerCertificate();
        const expirationDate = new Date(cert.valid_to);

        this.logger.info(`Certificate details retrieved for ${domain}`, {
          domain,
          issuer: cert.issuer,
          subject: cert.subject,
          validFrom: new Date(cert.valid_from).toISOString(),
          validTo: expirationDate.toISOString(),
          serialNumber: cert.serialNumber,
          fingerprint: cert.fingerprint,
          // @ts-ignore
          protocol: res.socket.getProtocol(),
        });

        const daysToExpiration = Math.floor(
          (expirationDate.getTime() - new Date().getTime()) /
            (1000 * 60 * 60 * 24)
        );

        this.logger.info(`Certificate expiration analysis for ${domain}`, {
          domain,
          daysRemaining: daysToExpiration,
          expirationDate: expirationDate.toISOString(),
          status: daysToExpiration <= 30 ? "WARNING" : "OK",
        });

        await this.checkAndSendAlert(domain, daysToExpiration, expirationDate);
        resolve();
      });
      req.on("error", (error) => {
        this.logger.error(`Failed to check certificate for ${domain}`, {
          domain,
          // @ts-ignore
          errorCode: error?.code,
          errorMessage: error.message,
          errorStack: error.stack,
        });
        reject(error);
      });

      req.end();
    });
  }

  /**
   * Check if alert should be sent and send it
   * @param {string} domain - Domain name
   * @param {number} daysToExpiration - Days until certificate expires
   * @param {Date} expirationDate - Certificate expiration date
   * @returns {Promise<void>}
   */
  async checkAndSendAlert(domain, daysToExpiration, expirationDate) {
    const history = await this.historyManager.loadHistory();
    const today = new Date().toISOString().split("T")[0];

    if (!history[domain]) {
      history[domain] = {};
    }

    let thresholdToAlert = null;
    for (const threshold of this.thresholdDays.sort((a, b) => a - b)) {
      if (daysToExpiration <= threshold) {
        thresholdToAlert = threshold;
        break;
      }
    }

    if (thresholdToAlert) {
      const lastAlert = history[domain][thresholdToAlert];
      if (!lastAlert || lastAlert !== today) {
        await this.slackNotifier.sendAlert(
          domain,
          daysToExpiration,
          expirationDate,
          thresholdToAlert
        );
        history[domain][thresholdToAlert] = today;
        await this.historyManager.saveHistory(history);
      }
    }
  }
}
