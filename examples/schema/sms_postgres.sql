-- 메시지 발송
INSERT INTO sms.msg(payload) VALUES('{
  "to": "01000000001",
  "from": "020000001",
  "text": "테스트 메시지"
}'::json);