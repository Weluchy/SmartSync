import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 30,
  duration: '30s',
};

export default function () {
  const url = 'http://localhost:8000/tasks/54/status'; 
  
  const payload = JSON.stringify({
    status: 'in_progress',
  });

  const params = {
    headers: {
      'Authorization': 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzkyMTc4NjIsInVzZXJfaWQiOjQsInVzZXJuYW1lIjoiYWRtaW4xIn0.UBdKxH1IET_WOoaCzhhn_rbJsgtTd6HnhN15FgibU40', 
      'Content-Type': 'application/json',
    },
  };

  const res = http.patch(url, payload, params);
  
  check(res, {
    'статус 200 (успешно)': (r) => r.status === 200,
  });
  
  sleep(0.5);
}