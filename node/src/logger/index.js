import { appendFile, mkdir } from 'node:fs/promises';
import { dirname } from 'node:path';

export class Logger {
  /**
   * @param {string} logFile - The path to the log file
   */
  constructor(logFile) {
    this.logFile = logFile;
    this.ensureLogDirectory();
  }

  /**
   * Ensures the log directory exists
   * @private
   * @returns {Promise<void>}
   */
  async ensureLogDirectory() {
    const dir = dirname(this.logFile);
    await mkdir(dir, { recursive: true });
  }

  /**
   * Logs a message to the console and file
   * @param {string} message - The message to log
   * @param {string} level - The log level (INFO, ERROR, WARNING)
   * @param {Object} details - Additional details to log
   * @returns {Promise<void>}
   */
  async log(message, level = "INFO", details = {}) {
    await this.ensureLogDirectory();
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
    await appendFile(this.logFile, logMessage);

    // Also log to console
    console.log(logMessage);
  }

  /**
   * Logs an informational message
   * @param {string} message - The message to log
   * @param {Object} details - Additional details to log
   * @returns {Promise<void>}
   */
  info(message, details = {}) {
    return this.log(message, "INFO", details);
  }

  /**
   * Logs an error message
   * @param {string} message - The message to log
   * @param {Object} details - Additional details to log
   * @returns {Promise<void>}
   */
  error(message, details = {}) {
    return this.log(message, "ERROR", details);
  }

  /**
   * Logs a warning message
   * @param {string} message - The message to log
   * @param {Object} details - Additional details to log
   * @returns {Promise<void>}
   */
  warning(message, details = {}) {
    return this.log(message, "WARNING", details);
  }
}
