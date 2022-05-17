CREATE TABLE msg (
  id integer  AUTO_INCREMENT primary key,
  createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  scheduledAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  sendAttempts SMALLINT NOT NULL DEFAULT 0,
  reportAttempts SMALLINT NOT NULL DEFAULT 0,
  `to` VARCHAR(20) AS (JSON_UNQUOTE(JSON_EXTRACT(payload, '$.to'))) STORED,
  `from` VARCHAR(20) AS (JSON_UNQUOTE(JSON_EXTRACT(payload, '$.from'))) STORED,
  groupId VARCHAR(255) AS (JSON_UNQUOTE(JSON_EXTRACT(result, '$.groupId'))) STORED,
  messageId VARCHAR(255) AS (JSON_UNQUOTE(JSON_EXTRACT(result, '$.messageId'))) STORED,
  status VARCHAR(20) AS (JSON_UNQUOTE(JSON_EXTRACT(result, '$.status'))) STORED,
  statusCode VARCHAR(255) AS (JSON_UNQUOTE(JSON_EXTRACT(result, '$.statusCode'))) STORED,
  statusMessage VARCHAR(255) AS (JSON_UNQUOTE(JSON_EXTRACT(result, '$.statusMessage'))) STORED,
  payload JSON,
  result JSON default NULL,
  sent BOOLEAN NOT NULL default false,
  KEY (`createdAt`),
  KEY (`updatedAt`),
  KEY (`scheduledAt`),
  KEY (`sendAttempts`),
  KEY (`reportAttempts`),
  KEY (`to`),
  KEY (`from`),
  KEY (groupId),
  KEY (messageId),
  KEY (status),
  KEY (statusCode),
  KEY (sent)
) DEFAULT CHARSET=utf8mb4;