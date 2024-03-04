-- 메시지 발송
INSERT INTO sms.msg(payload) VALUES('{
  "to": "01075690804",
  "from": "01031252016",
  "text": "테스트 메시지"
}'::json);