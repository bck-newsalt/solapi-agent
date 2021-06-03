-- 메시지 발송
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'text', '테스트 메시지'
));

-- LMS 발송
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'text', '한글 45자, 영자 90자 이상 입력되면 자동으로 LMS타입의 문자메시지가 발송됩니다. 0123456789 ABCDEFGHIJKLMNOPQRSTUVWXYZ'
));

-- MMS 발송(JPG 이미지 파일 상대경로 입력 시 /opt/agent/files/ 아래 파일)
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'file', 'sample.jpg',
  'subject', 'MMS 제목',
  'text', '파일이름만 입력하면 /opt/agent/files 아래에서 찾아 MMS로 발송됩니다.'
));

-- MMS 발송(JPG 이미지 파일 절대경로)
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'file', '/home/ubuntu/images/sample.jpg',
  'subject', 'MMS 제목',
  'text', '파일 절대 경로를 입력하면 해당 이미지를 MMS로 발송합니다.'
));

-- MMS 발송(이미지 아이디 발송 예제, 이미 업로드된 이미지의 아이디를 알고 있다면)
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'imageId', '이미지ID',
  'subject', 'MMS 제목',
  'text', '이미지 아이디가 입력되면 MMS로 발송됩니다.'
));

-- 타입 지정
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'text', '테스트 메시지',
  'type', 'SMS'
));

-- 카카오 알림톡 발송
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'text', '홍길동님 가입을 환영합니다.',
  'subject', '대체 발송시 LMS 제목',
  'kakaoOptions', json_object(
    'pfId', 'KA01PF1903260033550428GGGGGGGGGG',
    'templateId', 'KA01TP1903260033550428BBBBBBBBBB',
    'buttons', json_array(json_object(
      'buttonName', '홈페이지',
      'buttonType', 'WL',
      'linkPc', 'https://www.example.com',
      'linkMo', 'https://m.example.com'
    ), json_object(
      'buttonName', '앱 링크',
      'buttonType', 'AL',
      'linkIos', 'iosscheme://',
      'linkAnd', 'androidscheme://'
    ))
  )
));

-- 카카오 친구톡 발송
INSERT INTO msg(payload) VALUES(json_object(
  'to', '01000000001',
  'from', '020000001',
  'subject', '대체 발송시 LMS 제목',
  'kakaoOptions', json_object(
    'pfId', 'KA01PF1903260033550428GGGGGGGGGG',
    'buttons', json_array(json_object(
      'buttonName', '홈페이지',
      'buttonType', 'WL',
      'linkPc', 'https://www.example.com',
      'linkMo', 'https://m.example.com'
    ), json_object(
      'buttonName', '앱 링크',
      'buttonType', 'AL',
      'linkIos', 'iosscheme://',
      'linkAnd', 'androidscheme://'
    ))
  )
));

-- 해외 카카오 알림톡 발송
INSERT INTO msg(payload) VALUES(json_object(
  'country', '1', -- 국가번호 입력
  'to', '01000000001',
  'from', '020000001',
  'text', '홍길동님 가입을 환영합니다.',
  'subject', '대체 발송시 LMS 제목',
  'kakaoOptions', json_object(
    'pfId', 'KA01PF1903260033550428GGGGGGGGGG',
    'templateId', 'KA01TP1903260033550428BBBBBBBBBB',
    'buttons', json_array(json_object(
      'buttonName', '홈페이지',
      'buttonType', 'WL',
      'linkPc', 'https://www.example.com',
      'linkMo', 'https://m.example.com'
    ), json_object(
      'buttonName', '앱 링크',
      'buttonType', 'AL',
      'linkIos', 'iosscheme://',
      'linkAnd', 'androidscheme://'
    ))
  )
));

-- 발송 예약 (2021년 3월 26일 12시 20분에 발송 예약)
INSERT INTO msg(scheduledAt, payload) VALUES(
  '2021-03-26 12:20:00',
  json_object(
    'to', '01000000001',
    'from', '020000001',
    'text', '테스트 메시지'
  )
);
