import fs from "node:fs/promises";
import path from "node:path";

/**
 * @typedef {Object.<string, Object.<number, string>>} DomainHistory
 * Key is the domain name, value is an object where:
 * - key is the threshold day number
 * - value is the date of last alert in YYYY-MM-DD format
 */

/**
 * @typedef {Object.<string, DomainHistory>} HistoryData
 */

/**
 * @typedef {Error & { code?: string }} FileSystemError
 */

export class HistoryManager {
  constructor() {
    /** @private @type {string} */
    this.historyPath = path.join(process.cwd(), "logs", "data", "alert-history.json");
  }

  /**
   * Load the alert history from disk
   * @returns {Promise<HistoryData>}
   */
  async loadHistory() {
    try {
      await fs.mkdir(path.dirname(this.historyPath), { recursive: true });
      const data = await fs.readFile(this.historyPath, "utf-8");
      return JSON.parse(data);
    } catch (error) {
      const fsError = /** @type {FileSystemError} */ (error);
      if (fsError.code === "ENOENT") {
        return {};
      }
      throw error;
    }
  }

  /**
   * Save the alert history to disk
   * @param {HistoryData} history - The history data to save
   * @returns {Promise<void>}
   */
  async saveHistory(history) {
    await fs.mkdir(path.dirname(this.historyPath), { recursive: true });
    // Create a backup before saving
    try {
      const existingData = await fs.readFile(this.historyPath, "utf-8");
      const backupPath = `${this.historyPath}.backup`;
      await fs.writeFile(backupPath, existingData, "utf-8");
    } catch (error) {
      const fsError = /** @type {FileSystemError} */ (error);
      if (fsError.code !== "ENOENT") {
        throw error;
      }
    }
    await fs.writeFile(
      this.historyPath,
      JSON.stringify(history, null, 2),
      "utf-8"
    );
  }
}
