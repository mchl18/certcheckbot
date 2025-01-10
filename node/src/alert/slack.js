/**
 * Handles sending alerts to Slack
 */
export class SlackNotifier {
  /**
   * @param {string} webhookUrl - The Slack webhook URL
   */
  constructor(webhookUrl) {
    this.webhookUrl = webhookUrl;
  }

  /**
   * Send an alert to Slack about certificate expiration
   * @param {string} domain - The domain name
   * @param {number} daysToExpiration - Days until certificate expires
   * @param {Date} expirationDate - Certificate expiration date
   * @param {number} threshold - Alert threshold in days
   */
  async sendAlert(domain, daysToExpiration, expirationDate, threshold) {
    const message = {
      text: `🚨 *SSL Certificate Expiration Alert*\nThe SSL certificate for *${domain}* will expire in *${daysToExpiration}* days (${expirationDate.toISOString()}).\nThreshold reached: ${threshold} days\nPlease take action to renew the certificate before it expires.`,
    };

    try {
      const response = await fetch(this.webhookUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(message),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
    } catch (error) {
      console.error("Failed to send Slack alert:", error);
      throw error;
    }
  }
}
