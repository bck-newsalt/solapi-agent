-- DROP SCHEMA sms;

CREATE SCHEMA sms AUTHORIZATION postgres;

DROP TABLE IF EXISTS sms.msg;
CREATE TABLE sms.msg (
  id SERIAL4 NOT NULL,
  createdAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updatedAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  scheduledAt TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  sendAttempts SMALLINT NOT NULL DEFAULT 0,
  reportAttempts SMALLINT NOT NULL DEFAULT 0,
  "to" VARCHAR(20) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(payload, 'to'))) STORED,
  "from" VARCHAR(20) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(payload, 'from'))) STORED,
  groupId VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'groupId'))) STORED,
  messageId VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'messageId'))) STORED,
  status VARCHAR(20) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'status'))) STORED,
  statusCode VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'statusCode'))) STORED,
  statusMessage VARCHAR(255) GENERATED ALWAYS AS (trim('"' from json_extract_path_text(result, 'statusMessage'))) STORED,
  payload JSON,
  result JSON DEFAULT NULL,
  sent BOOLEAN NOT NULL DEFAULT false,
  CONSTRAINT msg_pkey PRIMARY KEY (id)
);
CREATE INDEX ix_msg_id ON sms.msg USING btree (id);
CREATE INDEX ix_msg_createdAt ON sms.msg USING btree (createdAt);
CREATE INDEX ix_msg_updatedAt ON sms.msg USING btree (updatedAt);
CREATE INDEX ix_msg_scheduledAt ON sms.msg USING btree (scheduledAt);
CREATE INDEX ix_msg_sendAttempts ON sms.msg USING btree (sendAttempts);
CREATE INDEX ix_msg_reportAttempts ON sms.msg USING btree (reportAttempts);
CREATE INDEX ix_msg_to ON sms.msg USING btree ("to");
CREATE INDEX ix_msg_from ON sms.msg USING btree ("from");
CREATE INDEX ix_msg_groupId ON sms.msg USING btree (groupId);
CREATE INDEX ix_msg_messageId ON sms.msg USING btree (messageId);
CREATE INDEX ix_msg_status ON sms.msg USING btree (status);
CREATE INDEX ix_msg_statusCode ON sms.msg USING btree (statusCode);
CREATE INDEX ix_msg_sent ON sms.msg USING btree (sent);
